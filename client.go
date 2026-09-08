package mt5

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const defaultPingCommand = 51

// Message represents a parsed MT5 protocol message received from the server.
type Message struct {
	// Command is the protocol command identifier.
	Command uint16
	// Code is the response/status code.
	Code uint8
	// Body is the message payload after the 5-byte header.
	Body []byte
	// Raw is the complete decrypted message including the header.
	Raw []byte
}

// MessageHandler is a function that processes a received [Message].
type MessageHandler func(Message) error

// ConnectHook is called after a successful connection. The reconnect
// parameter is true for all connections after the first.
type ConnectHook func(client *Client, reconnect bool) error

// DisconnectHook is called when the client disconnects due to an error.
type DisconnectHook func(client *Client, err error)

// Config holds the parameters for creating a new [Client].
type Config struct {
	// URL is the WebSocket endpoint (e.g. "wss://broker.example/ws").
	URL string
	// Dialer is the WebSocket dialer. Defaults to websocket.DefaultDialer.
	Dialer *websocket.Dialer
	// RequestHeader is sent during the WebSocket handshake.
	RequestHeader http.Header
	// Crypto handles payload encryption and decryption.
	// Defaults to [DefaultCrypto] with [DefaultObfuscatedKey].
	Crypto Crypto
	// Codec handles command encoding, framing, and message parsing.
	// Defaults to [DefaultCodec].
	Codec Codec
	// SkipCommands is a set of command IDs to silently discard.
	// Defaults to {25, 51} (heartbeat-related).
	SkipCommands map[uint16]bool
	// PingInterval controls how often heartbeat pings are sent.
	// Defaults to 5 seconds.
	PingInterval time.Duration
	// PingCommand is the command ID used for heartbeat pings.
	// Defaults to 51.
	PingCommand uint16
	// ReconnectBackoff is the delay between reconnection attempts.
	// Defaults to 3 seconds.
	ReconnectBackoff time.Duration
	// ReconnectMaxAttempts limits reconnection attempts. Zero disables
	// reconnection, negative values allow unlimited attempts.
	ReconnectMaxAttempts int
	// AutoRequestKey causes the client to automatically send a key
	// request after connecting.
	AutoRequestKey bool
	// RequestKeyBody is the payload sent with the automatic key request.
	RequestKeyBody []byte
	// MessageBuffer is the size of the message channel buffer.
	// Defaults to 64.
	MessageBuffer int
	// ErrorBuffer is the size of the error channel buffer.
	// Defaults to 64.
	ErrorBuffer int
	// OnConnect is called after each successful connection.
	OnConnect ConnectHook
	// OnDisconnect is called when the connection drops.
	OnDisconnect DisconnectHook
}

// Client is a low-level WebSocket client for the MT5 binary protocol.
// It manages the connection lifecycle, encryption, heartbeats, and
// message dispatch.
//
// Use [New] to create a Client, then call [Client.Run] to start the
// connection loop.
type Client struct {
	cfg Config

	mu      sync.RWMutex
	writeMu sync.Mutex

	conn    *websocket.Conn
	closed  bool
	closeCh chan struct{}

	handlersMu sync.RWMutex
	handlers   map[uint16][]MessageHandler
	onAny      []MessageHandler

	msgCh chan Message
	errCh chan error
}

// New creates a new [Client] with the given configuration. It does not
// establish a connection; call [Client.Connect] or [Client.Run] for that.
func New(config Config) (*Client, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("url is required")
	}

	if config.Crypto == nil {
		crypto, err := NewDefaultCrypto(DefaultObfuscatedKey)
		if err != nil {
			return nil, err
		}
		config.Crypto = crypto
	}

	if config.Codec == nil {
		config.Codec = NewDefaultCodec()
	}

	if config.PingInterval <= 0 {
		config.PingInterval = 5 * time.Second
	}

	if config.PingCommand == 0 {
		config.PingCommand = defaultPingCommand
	}

	if config.ReconnectBackoff <= 0 {
		config.ReconnectBackoff = 3 * time.Second
	}

	if config.MessageBuffer <= 0 {
		config.MessageBuffer = 64
	}

	if config.ErrorBuffer <= 0 {
		config.ErrorBuffer = 64
	}

	if config.SkipCommands == nil {
		config.SkipCommands = map[uint16]bool{25: true, 51: true}
	}

	if config.Dialer == nil {
		config.Dialer = websocket.DefaultDialer
	}

	c := &Client{
		cfg:      config,
		closeCh:  make(chan struct{}),
		handlers: make(map[uint16][]MessageHandler),
		msgCh:    make(chan Message, config.MessageBuffer),
		errCh:    make(chan error, config.ErrorBuffer),
	}

	return c, nil
}

// Messages returns a read-only channel of dispatched messages.
func (c *Client) Messages() <-chan Message {
	return c.msgCh
}

// Errors returns a read-only channel of non-fatal errors encountered
// during operation.
func (c *Client) Errors() <-chan error {
	return c.errCh
}

// On registers a handler for a specific command ID. Multiple handlers
// can be registered for the same command.
func (c *Client) On(command uint16, handler MessageHandler) {
	c.handlersMu.Lock()
	defer c.handlersMu.Unlock()
	c.handlers[command] = append(c.handlers[command], handler)
}

// OnAny registers a handler that is called for every received message,
// regardless of command ID.
func (c *Client) OnAny(handler MessageHandler) {
	c.handlersMu.Lock()
	defer c.handlersMu.Unlock()
	c.onAny = append(c.onAny, handler)
}

// Connect establishes a WebSocket connection to the server. If
// [Config.AutoRequestKey] is true, a key request is sent immediately
// after connecting.
func (c *Client) Connect(ctx context.Context) error {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return ErrClientClosed
	}
	c.mu.RUnlock()

	conn, _, err := c.cfg.Dialer.DialContext(ctx, c.cfg.URL, c.cfg.RequestHeader)
	if err != nil {
		return err
	}

	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		_ = conn.Close()
		return ErrClientClosed
	}
	c.conn = conn
	c.mu.Unlock()

	if c.cfg.AutoRequestKey {
		if err := c.RequestKey(c.cfg.RequestKeyBody); err != nil {
			_ = c.closeConn()
			return err
		}
	}

	return nil
}

// Run connects to the server and enters the read loop. It automatically
// reconnects on errors according to [Config.ReconnectMaxAttempts] and
// [Config.ReconnectBackoff]. Run blocks until the context is cancelled,
// the client is closed, or reconnection attempts are exhausted.
func (c *Client) Run(ctx context.Context) error {
	attempt := 0
	connectedOnce := false

	for {
		if err := c.Connect(ctx); err != nil {
			if !c.shouldRetry(attempt, err) {
				return err
			}
			attempt++
			if err := c.waitBackoff(ctx); err != nil {
				return nil
			}
			continue
		}

		if c.cfg.OnConnect != nil {
			err := c.cfg.OnConnect(c, connectedOnce)
			if err != nil {
				c.publishError(fmt.Errorf("on connect error: %w", err))
				_ = c.closeConn()
				if !c.shouldRetry(attempt, err) {
					return err
				}
				attempt++
				if err := c.waitBackoff(ctx); err != nil {
					return nil
				}
				continue
			}
		}

		connectedOnce = true
		attempt = 0

		heartbeatDone := make(chan struct{})
		go c.heartbeatLoop(heartbeatDone)

		err := c.readLoop(ctx)
		if err != nil && c.cfg.OnDisconnect != nil {
			c.cfg.OnDisconnect(c, err)
		}
		close(heartbeatDone)
		_ = c.closeConn()

		if err == nil {
			return nil
		}

		if !c.shouldRetry(attempt, err) {
			return err
		}

		attempt++
		if err := c.waitBackoff(ctx); err != nil {
			return nil
		}
	}
}

// Close gracefully shuts down the client and its connection.
func (c *Client) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	close(c.closeCh)
	c.mu.Unlock()

	return c.closeConn()
}

// SendCommand encrypts and sends a command with the given body.
func (c *Client) SendCommand(command uint16, body []byte) error {
	payload, err := c.cfg.Codec.EncodeCommand(command, body)
	if err != nil {
		return err
	}

	cipherText, err := c.cfg.Crypto.Encrypt(payload)
	if err != nil {
		return err
	}

	frame, err := c.cfg.Codec.Pack(cipherText)
	if err != nil {
		return err
	}

	return c.Send(frame)
}

// Send writes a raw framed message to the WebSocket connection.
func (c *Client) Send(frame []byte) error {
	c.mu.RLock()
	if c.closed {
		c.mu.RUnlock()
		return ErrClientClosed
	}
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return ErrNotConnected
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return conn.WriteMessage(websocket.BinaryMessage, frame)
}

// RequestKey sends a session key request to the server.
func (c *Client) RequestKey(param []byte) error {
	if param == nil {
		param = make([]byte, 64)
	}
	return c.SendCommand(0, param)
}

// LoadSessionKeyFromMessage extracts the session key from a server
// message starting at bodyOffset and sets it as the active encryption key.
func (c *Client) LoadSessionKeyFromMessage(msg Message, bodyOffset int) error {
	if bodyOffset < 0 || bodyOffset >= len(msg.Body) {
		return fmt.Errorf("invalid key offset %d", bodyOffset)
	}
	return c.cfg.Crypto.SetKey(msg.Body[bodyOffset:])
}

// SetKeyFromString replaces the encryption key by decoding an obfuscated
// key string.
func (c *Client) SetKeyFromString(key string) error {
	return c.cfg.Crypto.SetKeyFromString(key)
}

// SetSessionKey replaces the encryption key with raw key bytes.
func (c *Client) SetSessionKey(key []byte) error {
	return c.cfg.Crypto.SetKey(key)
}

func (c *Client) readLoop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-c.closeCh:
			return nil
		default:
		}

		c.mu.RLock()
		conn := c.conn
		c.mu.RUnlock()
		if conn == nil {
			return ErrNotConnected
		}

		_, raw, err := conn.ReadMessage()
		if err != nil {
			c.publishError(ErrReadError)
			return ErrReadError
		}

		payload, err := c.cfg.Codec.Unpack(raw)
		if err != nil {
			c.publishError(fmt.Errorf("unpack error: %w", err))
			continue
		}

		plain, err := c.cfg.Crypto.Decrypt(payload)
		if err != nil {
			c.publishError(fmt.Errorf("decrypt error: %w", err))
			continue
		}

		msg, err := c.cfg.Codec.ParseMessage(plain)
		if err != nil {
			c.publishError(fmt.Errorf("parse error: %w", err))
			continue
		}

		if c.cfg.SkipCommands[msg.Command] {
			continue
		}

		fmt.Printf("Recieved msg Code %d Body %d bytes\n", msg.Command, len(msg.Body))

		c.dispatch(msg)
	}
}

func (c *Client) heartbeatLoop(done <-chan struct{}) {
	ticker := time.NewTicker(c.cfg.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-c.closeCh:
			return
		case <-ticker.C:
			if err := c.SendCommand(c.cfg.PingCommand, nil); err != nil {
				c.publishError(fmt.Errorf("ping error: %w", err))
				return
			}
		}
	}
}

func (c *Client) dispatch(msg Message) {
	c.handlersMu.RLock()
	handlers := append([]MessageHandler{}, c.handlers[msg.Command]...)
	onAny := append([]MessageHandler{}, c.onAny...)
	c.handlersMu.RUnlock()

	for _, h := range handlers {
		if err := h(msg); err != nil {
			c.publishError(err)
		}
	}

	for _, h := range onAny {
		if err := h(msg); err != nil {
			c.publishError(err)
		}
	}
}

func (c *Client) shouldRetry(attempt int, err error) bool {
	if err == nil {
		return false
	}

	c.mu.RLock()
	closed := c.closed
	c.mu.RUnlock()
	if closed {
		return false
	}

	if c.cfg.ReconnectMaxAttempts == 0 {
		return false
	}

	if c.cfg.ReconnectMaxAttempts < 0 {
		return true
	}

	return attempt < c.cfg.ReconnectMaxAttempts
}

func (c *Client) waitBackoff(ctx context.Context) error {
	t := time.NewTimer(c.cfg.ReconnectBackoff)
	defer t.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.closeCh:
		return ErrClientClosed
	case <-t.C:
		return nil
	}
}

func (c *Client) closeConn() error {
	c.mu.Lock()
	conn := c.conn
	c.conn = nil
	c.mu.Unlock()

	if conn != nil {
		return conn.Close()
	}
	return nil
}

func (c *Client) publishError(err error) {
	if err == nil {
		return
	}

	select {
	case c.errCh <- err:
	default:
	}
}
