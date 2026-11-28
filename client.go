package mt5

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/kahono0/mt5/models"
)

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
	// ServerURL is the logical server address sent during login.
	ServerURL string
	// AccountID is the MT5 account number.
	AccountID uint64
	// Password is the account password.
	Password string
	// DeviceID is a unique device identifier string.
	DeviceID string

	// SessionKeyOffset is the byte offset into the first message's body
	// where the session key begins. Defaults to 66.
	SessionKeyOffset int
	// SessionKeyBody is the payload sent with the session key request.
	SessionKeyBody []byte
	// KeyExtractor overrides the default session key extraction logic.
	KeyExtractor KeyExtractor
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
	// PingTimeout is the maximum time to wait for a ping acknowledgment.
	// Defaults to 30 seconds.
	// If a ping is not acknowledged within this time, the client will consider the connection lost and attempt to reconnect.
	PingTimeout time.Duration
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
	// OnError is called for every error encountered during operation.
	// Unlike the Errors channel, this callback is never lossy — it is
	// always invoked regardless of channel buffer pressure.
	OnError func(error)
	// OnConnect is called after each successful connection.
	OnConnect ConnectHook
	// OnDisconnect is called when the connection drops.
	OnDisconnect DisconnectHook
}

// handlerEntry pairs a message handler with an optional per-handler
// error callback.
type handlerEntry struct {
	id          uint64
	fn          MessageHandler
	decode      func(Message) (any, error)
	invokeTyped func(any, Message) error
	onDecodeErr func(error, Message)
	onError     func(error)
	onRemove    func()
}

// MessageResult contains either a decoded value or the error from decoding a
// message for a typed subscription.
type MessageResult[T any] struct {
	Value   T
	Message Message
	Err     error
}

// Subscription receives decoded messages for one typed command. Each
// subscription has its own channel, but decoded values are shared with other
// handlers and subscriptions for that message and should be treated as
// read-only. When its buffer is full, new messages are dropped and counted by
// Dropped. Call Cancel to close the channel.
type Subscription[T any] struct {
	C      <-chan MessageResult[T]
	state  *subscriptionState[T]
	cancel func()
}

// Cancel unregisters the subscription and closes its channel. It is safe to
// call more than once.
func (s *Subscription[T]) Cancel() {
	if s != nil && s.cancel != nil {
		s.cancel()
	}
}

// Dropped returns the number of messages discarded because the channel buffer
// was full.
func (s *Subscription[T]) Dropped() uint64 {
	if s == nil || s.state == nil {
		return 0
	}
	return s.state.dropped.Load()
}

type subscriptionState[T any] struct {
	mu      sync.Mutex
	ch      chan MessageResult[T]
	closed  bool
	dropped atomic.Uint64
}

func (s *subscriptionState[T]) offer(result MessageResult[T]) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	select {
	case s.ch <- result:
	default:
		s.dropped.Add(1)
	}
}

func (s *subscriptionState[T]) close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.closed {
		s.closed = true
		close(s.ch)
	}
}

// HandlerOption configures how a registered handler behaves.
type HandlerOption func(*handlerEntry)

// WithErrorHandler sets a per-handler error callback. When set, errors
// returned by this handler are routed to fn instead of the global error
// channel and [Config.OnError] callback.
func WithErrorHandler(fn func(error)) HandlerOption {
	return func(h *handlerEntry) {
		h.onError = fn
	}
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

	handlersMu    sync.RWMutex
	handlers      map[uint16][]handlerEntry
	onAny         []handlerEntry
	nextHandlerID uint64

	authenticated bool

	//msgCh chan Message
	//errCh chan error

	pingAckCh     chan struct{}
	heartbeatDone chan struct{}
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

	if config.PingTimeout <= 0 {
		config.PingTimeout = 30 * time.Second
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
		config.SkipCommands = map[uint16]bool{25: true}
	}

	if config.Dialer == nil {
		config.Dialer = websocket.DefaultDialer
	}

	c := &Client{
		cfg:      config,
		closeCh:  make(chan struct{}),
		handlers: make(map[uint16][]handlerEntry),
		//msgCh:    make(chan Message, config.MessageBuffer),
		//errCh:    make(chan error, config.ErrorBuffer),

		pingAckCh: make(chan struct{}, 1),
	}

	c.On(CommandRequestSessionKey, c.handleSessionKeyMessage)
	c.On(CommandLogin, c.handleLoginAck)
	c.On(CommandLoginSuccess, c.handleLoginSuccess)

	return c, nil
}

// On registers a handler for a specific command ID. Multiple handlers
// can be registered for the same command. Use [WithErrorHandler] to
// route errors from this handler to a dedicated callback instead of
// the global error channel. Returns a function that, when called,
// removes the handler.
func (c *Client) On(command uint16, handler MessageHandler, opts ...HandlerOption) func() {
	e := handlerEntry{fn: handler}
	for _, o := range opts {
		if o != nil {
			o(&e)
		}
	}
	return c.addHandler(command, e)
}

func (c *Client) addHandler(command uint16, e handlerEntry) func() {
	c.handlersMu.Lock()
	c.nextHandlerID++
	e.id = c.nextHandlerID
	c.handlers[command] = append(c.handlers[command], e)
	c.handlersMu.Unlock()

	return func() {
		c.handlersMu.Lock()
		handlers := c.handlers[command]
		var onRemove func()
		for i, h := range handlers {
			if h.id == e.id {
				last := len(handlers) - 1
				onRemove = h.onRemove
				copy(handlers[i:], handlers[i+1:])
				handlers[last] = handlerEntry{}
				c.handlers[command] = handlers[:last]
				break
			}
		}
		c.handlersMu.Unlock()
		if onRemove != nil {
			onRemove()
		}
	}
}

// registerTyped adapts a typed command and callback to the common handler list.
func registerTyped[T any](c *Client, command Command[T], handler func(T, Message) error, opts ...HandlerOption) func() {
	e := handlerEntry{
		decode: func(msg Message) (any, error) {
			return command.decode(msg)
		},
		invokeTyped: func(value any, msg Message) error {
			return handler(value.(T), msg)
		},
	}
	for _, option := range opts {
		if option != nil {
			option(&e)
		}
	}

	return c.addHandler(command.id, e)
}

// ID returns the wire command ID associated with the specification.
func (command Command[T]) ID() uint16 { return command.id }

// Subscribe creates a buffered channel subscription for the command. A full
// channel drops the newest result rather than blocking message dispatch; use
// Dropped to monitor those losses. Pass zero for an unbuffered channel.
func (command Command[T]) Subscribe(client *Client, buffer int) (*Subscription[T], error) {
	if client == nil {
		return nil, fmt.Errorf("cannot subscribe to command %d on a nil client", command.id)
	}
	if command.decode == nil {
		return nil, fmt.Errorf("command %d has no decoder", command.id)
	}
	if buffer < 0 {
		return nil, fmt.Errorf("subscription buffer cannot be negative: %d", buffer)
	}

	state := &subscriptionState[T]{ch: make(chan MessageResult[T], buffer)}
	entry := handlerEntry{
		onRemove: state.close,
		decode: func(msg Message) (any, error) {
			return command.decode(msg)
		},
		invokeTyped: func(value any, msg Message) error {
			state.offer(MessageResult[T]{Value: value.(T), Message: msg})
			return nil
		},
		onDecodeErr: func(err error, msg Message) {
			state.offer(MessageResult[T]{Message: msg, Err: err})
		},
	}

	remove := client.addHandler(command.id, entry)
	subscription := &Subscription[T]{C: state.ch, state: state}
	subscription.cancel = func() {
		remove()
		state.close()
	}
	return subscription, nil
}

// OnAny registers a handler that is called for every received message,
// regardless of command ID. Use [WithErrorHandler] to route errors from
// this handler to a dedicated callback instead of the global error channel.
// Returns a function that, when called, removes the handler.
func (c *Client) OnAny(handler MessageHandler, opts ...HandlerOption) func() {
	c.handlersMu.Lock()
	defer c.handlersMu.Unlock()

	c.nextHandlerID++
	id := c.nextHandlerID

	e := handlerEntry{id: id, fn: handler}
	for _, o := range opts {
		if o != nil {
			o(&e)
		}
	}

	c.onAny = append(c.onAny, e)

	return func() {
		c.handlersMu.Lock()
		defer c.handlersMu.Unlock()
		for i, h := range c.onAny {
			if h.id == id {
				last := len(c.onAny) - 1
				copy(c.onAny[i:], c.onAny[i+1:])
				c.onAny[last] = handlerEntry{}
				c.onAny = c.onAny[:last]
				break
			}
		}
	}
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

	//conn, _, err := c.cfg.Dialer.DialContext(ctx, c.cfg.URL, c.cfg.RequestHeader)
	conn, _, err := websocket.DefaultDialer.Dial(c.cfg.URL, c.cfg.RequestHeader)
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
			c.publishError(fmt.Errorf("connect error: %w", err))
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

		c.mu.Lock()
		c.authenticated = false
		c.mu.Unlock()

		c.cfg.Crypto.ResetKey() // Reset the encryption key on disconnect to ensure a fresh key is requested on reconnect.

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

	c.closeSubscriptions()
	return c.closeConn()
}

func (c *Client) closeSubscriptions() {
	c.handlersMu.Lock()
	var closeFns []func()
	for command, handlers := range c.handlers {
		write := 0
		originalLen := len(handlers)
		for _, handler := range handlers {
			if handler.onRemove != nil {
				closeFns = append(closeFns, handler.onRemove)
				continue
			}
			handlers[write] = handler
			write++
		}
		for i := write; i < originalLen; i++ {
			handlers[i] = handlerEntry{}
		}
		c.handlers[command] = handlers[:write]
	}
	c.handlersMu.Unlock()

	for _, closeFn := range closeFns {
		closeFn()
	}
}

var i int
var tt time.Duration

func (c *Client) readLoop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
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

		conn.SetReadDeadline(time.Now().Add(c.cfg.PingTimeout))
		_, raw, err := conn.ReadMessage()
		if err != nil {
			c.publishError(fmt.Errorf("read error: %w", err))
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

		if msg.Command == CommandPing {
			if err := c.handlePingAck(msg); err != nil {
				c.publishError(fmt.Errorf("ping ack error: %w", err))
			}
		}

		st := time.Now()
		c.dispatch(msg)
		tt += time.Since(st)
		i++
		if i%100 == 0 {
			fmt.Printf("avg dispatch time: %v\n", tt/time.Duration(i))
		}
	}
}

func (c *Client) dispatch(msg Message) {
	c.handlersMu.RLock()
	handlers := append([]handlerEntry{}, c.handlers[msg.Command]...)
	onAny := append([]handlerEntry{}, c.onAny...)
	c.handlersMu.RUnlock()

	var decoded bool
	var decodedValue any
	var decodeErr error
	for _, h := range handlers {
		var err error
		if h.decode == nil {
			err = h.fn(msg)
		} else {
			if !decoded {
				decodedValue, decodeErr = h.decode(msg)
				decoded = true
			}
			if decodeErr != nil {
				if h.onDecodeErr != nil {
					h.onDecodeErr(decodeErr, msg)
				} else {
					err = decodeErr
				}
			} else {
				err = h.invokeTyped(decodedValue, msg)
			}
		}
		if err != nil {
			if h.onError != nil {
				h.onError(err)
			} else {
				c.publishError(err)
			}
		}
	}

	for _, h := range onAny {
		if err := h.fn(msg); err != nil {
			if h.onError != nil {
				h.onError(err)
			} else {
				c.publishError(err)
			}
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

	if c.cfg.OnError != nil {
		c.cfg.OnError(err)
	}

	//select {
	//case c.errCh <- err:
	//default:
	//}
}

// auth funcs
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

func (c *Client) handleSessionKeyMessage(msg Message) error {
	keyBytes, err := c.extractSessionKey(msg)
	if err != nil {
		return err
	}

	if err := c.SetSessionKey(keyBytes); err != nil {
		return fmt.Errorf("failed to set session key: %w", err)
	}

	if err := c.Login(); err != nil {
		return fmt.Errorf("failed to send login after key load: %w", err)
	}

	return nil
}

func (c *Client) extractSessionKey(msg Message) ([]byte, error) {
	if c.cfg.KeyExtractor != nil {
		return c.cfg.KeyExtractor(msg, c.cfg.SessionKeyOffset)
	}

	if c.cfg.SessionKeyOffset < 0 || c.cfg.SessionKeyOffset >= len(msg.Body) {
		return nil, fmt.Errorf(
			"%w: body length %d with offset %d",
			ErrSessionKeyNotLoaded,
			len(msg.Body),
			c.cfg.SessionKeyOffset,
		)
	}

	return append([]byte(nil), msg.Body[c.cfg.SessionKeyOffset:]...), nil
}

// Login sends the login command to the server. This is called
// automatically after the session key is loaded, but can be called
// manually if needed.
func (c *Client) Login() error {
	loginRequest := models.LoginRequest{
		ServerURL: c.cfg.ServerURL,
		Password:  c.cfg.Password,
		AccountID: c.cfg.AccountID,
		DeviceID:  c.cfg.DeviceID,
	}
	body, err := loginRequest.Serialize()
	if err != nil {
		return err
	}

	if err := c.SendCommand(CommandLogin, body); err != nil {
		return fmt.Errorf("failed to send login command: %w", err)
	}
	return nil
}

func (c *Client) handleLoginAck(msg Message) error {
	if msg.Code != 0 {
		return &ServerError{
			Command: msg.Command,
			Code:    msg.Code,
			Msg:     "login rejected",
		}
	}

	c.mu.Lock()
	c.authenticated = true
	c.mu.Unlock()
	return nil
}

func (c *Client) handleLoginSuccess(msg Message) error {
	if msg.Code != 0 {
		return &ServerError{
			Command: msg.Command,
			Code:    msg.Code,
			Msg:     "login success command failed",
		}
	}

	c.mu.Lock()
	c.authenticated = true
	c.mu.Unlock()

	return nil
}

func (c *Client) waitUntilAuthenticated() error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.RLock()
		auth := c.authenticated
		c.mu.RUnlock()

		if auth {
			return nil
		}
	}

	return fmt.Errorf("client not authenticated")
}

// IsAuthenticated reports whether the client has completed the login
// handshake.
func (c *Client) IsAuthenticated() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.authenticated
}

// heartbeat funcs
func (c *Client) handlePingAck(msg Message) error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.pingAckCh != nil {
		select {
		case c.pingAckCh <- struct{}{}:
		default:
		}
	}
	return nil
}

var (
	ErrPreparePingMessage = fmt.Errorf("failed to prepare ping message")
	ErrSendPing           = fmt.Errorf("failed to send ping message")
	ErrPingTimeout        = fmt.Errorf("ping timeout")
)

// heartbeatLoop sends periodic ping commands and checks for acknowledgments.
func (c *Client) heartbeatLoop(done chan struct{}) {
	if err := c.waitUntilAuthenticated(); err != nil {
		c.publishError(fmt.Errorf("heartbeat loop: %w", err))
		return
	}
	frame, err := c.prepareMessage(CommandPing, nil)
	if err != nil {
		c.publishError(ErrPreparePingMessage)
		return
	}

	timer := time.NewTimer(c.cfg.PingInterval)
	defer timer.Stop()

	for {
		select {
		case <-done:
			return
		case <-timer.C:
			if err := c.Send(frame); err != nil {
				c.publishError(fmt.Errorf("%w: %v", ErrSendPing, err))
				return
			}
		case <-c.pingAckCh:
			timer.Reset(c.cfg.PingInterval)
			continue
		}

	}
}

// SendCommand encrypts and sends a command with the given body.
func (c *Client) SendCommand(command uint16, body []byte) error {
	frame, err := c.prepareMessage(command, body)
	if err != nil {
		return err
	}

	return c.Send(frame)
}

func (c *Client) prepareMessage(command uint16, body []byte) ([]byte, error) {
	payload, err := c.cfg.Codec.EncodeCommand(command, body)
	if err != nil {
		return nil, err
	}

	cipherText, err := c.cfg.Crypto.Encrypt(payload)
	if err != nil {
		return nil, err
	}

	frame, err := c.cfg.Codec.Pack(cipherText)
	if err != nil {
		return nil, err
	}

	return frame, nil
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
