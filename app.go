package mt5

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const defaultSessionKeyOffset = 66

// KeyExtractor is a function that extracts the session key bytes from a
// server message. It receives the message and the configured body offset.
type KeyExtractor func(msg Message, bodyOffset int) ([]byte, error)

// DisconnectFunc is called when the [AppClient] loses its connection.
type DisconnectFunc func(err error)

// AppConfig holds the parameters for creating a new [AppClient].
type AppConfig struct {
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
	// SkipCommands is a set of command IDs to silently discard.
	SkipCommands map[uint16]bool
	// PingInterval controls how often heartbeat pings are sent.
	PingInterval time.Duration
	// PingCommand is the command ID used for heartbeat pings.
	PingCommand uint16
	// ReconnectBackoff is the delay between reconnection attempts.
	ReconnectBackoff time.Duration
	// ReconnectMaxAttempts limits reconnection attempts.
	ReconnectMaxAttempts int
	// MessageBuffer is the size of the message channel buffer.
	MessageBuffer int
	// ErrorBuffer is the size of the error channel buffer.
	ErrorBuffer int
	// OnError is called for every error encountered during operation.
	// Unlike the Errors channel, this callback is never lossy — it is
	// always invoked regardless of channel buffer pressure.
	OnError func(error)
	// OnDisconnect is called when the connection drops.
	OnDisconnect DisconnectFunc
}

// AppClient is a high-level MT5 client that wraps [Client] and adds
// automatic session-key negotiation, login, and convenient action methods.
//
// Use [NewApp] to create an AppClient.
type AppClient struct {
	cfg  AppConfig
	core *Client

	mu               sync.RWMutex
	authenticated    bool
	sessionKeyLoaded bool
	awaitingKey      bool

	initHandlersOnce sync.Once
}

// NewApp creates a new [AppClient] with the given configuration. It
// validates that all required fields are present and initializes the
// underlying [Client].
func NewApp(cfg AppConfig) (*AppClient, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("%w: URL is required", ErrInvalidConfig)
	}
	if cfg.ServerURL == "" {
		return nil, fmt.Errorf("%w: ServerURL is required", ErrInvalidConfig)
	}
	if cfg.AccountID == 0 {
		return nil, fmt.Errorf("%w: AccountID is required", ErrInvalidConfig)
	}
	if cfg.Password == "" {
		return nil, fmt.Errorf("%w: Password is required", ErrInvalidConfig)
	}
	if cfg.DeviceID == "" {
		return nil, fmt.Errorf("%w: DeviceID is required", ErrInvalidConfig)
	}

	if cfg.SessionKeyOffset <= 0 {
		cfg.SessionKeyOffset = defaultSessionKeyOffset
	}
	if cfg.PingCommand == 0 {
		cfg.PingCommand = CommandPing
	}

	core, err := New(Config{
		URL:                  cfg.URL,
		Dialer:               cfg.Dialer,
		RequestHeader:        cfg.RequestHeader,
		SkipCommands:         cfg.SkipCommands,
		PingInterval:         cfg.PingInterval,
		PingCommand:          cfg.PingCommand,
		ReconnectBackoff:     cfg.ReconnectBackoff,
		ReconnectMaxAttempts: cfg.ReconnectMaxAttempts,
		AutoRequestKey:       true,
		RequestKeyBody:       cfg.SessionKeyBody,
		MessageBuffer:        cfg.MessageBuffer,
		ErrorBuffer:          cfg.ErrorBuffer,
		OnError:              cfg.OnError,
	})
	if err != nil {
		return nil, err
	}

	c := &AppClient{
		cfg:         cfg,
		core:        core,
		awaitingKey: true,
	}
	c.initInternalHandlers()
	return c, nil
}

// Core returns the underlying low-level [Client].
func (c *AppClient) Core() *Client {
	return c.core
}

// Cfg returns a copy of the [AppConfig] used to create this client.
func (c *AppClient) Cfg() AppConfig {
	return c.cfg
}

// Start connects to the server and runs the read loop. It blocks until
// the context is cancelled or the client is closed.
func (c *AppClient) Start(ctx context.Context) error {
	c.mu.Lock()
	c.awaitingKey = true
	c.sessionKeyLoaded = false
	c.authenticated = false
	c.mu.Unlock()

	return c.core.Run(ctx)
}

// Close gracefully shuts down the client.
func (c *AppClient) Close() error {
	return c.core.Close()
}

// Messages returns a read-only channel of dispatched messages.
func (c *AppClient) Messages() <-chan Message {
	return c.core.Messages()
}

// Errors returns a read-only channel of non-fatal errors.
func (c *AppClient) Errors() <-chan error {
	return c.core.Errors()
}

// RegisterHandler registers a handler for a specific command ID on the
// underlying [Client]. Use [WithErrorHandler] to route errors from this
// handler to a dedicated callback.
func (c *AppClient) RegisterHandler(command uint16, handler MessageHandler, opts ...HandlerOption) {
	c.core.On(command, handler, opts...)
}

// RegisterAnyHandler registers a handler that is called for every
// received message. Use [WithErrorHandler] to route errors from this
// handler to a dedicated callback.
func (c *AppClient) RegisterAnyHandler(handler MessageHandler, opts ...HandlerOption) {
	c.core.OnAny(handler, opts...)
}

// IsAuthenticated reports whether the client has completed the login
// handshake.
func (c *AppClient) IsAuthenticated() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.authenticated
}

// SetUnauthenticated forces the authentication state to false.
func (c *AppClient) SetUnauthenticated() {
	c.mu.Lock()
	c.authenticated = false
	c.mu.Unlock()
}

// IsSessionKeyLoaded reports whether a session key has been extracted
// from the server handshake.
func (c *AppClient) IsSessionKeyLoaded() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sessionKeyLoaded
}

// WaitUntilAuthenticated blocks until the client is authenticated or
// the context is cancelled.
func (c *AppClient) WaitUntilAuthenticated(ctx context.Context) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		if c.IsAuthenticated() {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func (c *AppClient) initInternalHandlers() {
	c.initHandlersOnce.Do(func() {
		c.core.OnAny(c.handleSessionKeyMessage)
		c.core.On(CommandLogin, c.handleLoginAck)
		c.core.On(CommandLoginSuccess, c.handleLoginSuccess)
	})
}

func (c *AppClient) handleSessionKeyMessage(msg Message) error {
	c.mu.RLock()
	awaiting := c.awaitingKey
	loaded := c.sessionKeyLoaded
	c.mu.RUnlock()

	if !awaiting && msg.Command != CommandRequestSessionKey {
		return nil
	}

	if loaded && msg.Command != CommandRequestSessionKey {
		return nil
	}

	keyBytes, err := c.extractSessionKey(msg)
	if err != nil {
		if msg.Command == CommandRequestSessionKey {
			return err
		}
		return nil
	}

	if err := c.core.SetSessionKey(keyBytes); err != nil {
		return fmt.Errorf("failed to set session key: %w", err)
	}

	c.mu.Lock()
	c.sessionKeyLoaded = true
	c.awaitingKey = false
	c.authenticated = false
	c.mu.Unlock()

	if err := c.Login(); err != nil {
		return fmt.Errorf("failed to send login after key load: %w", err)
	}

	return nil
}

func (c *AppClient) extractSessionKey(msg Message) ([]byte, error) {
	if c.cfg.KeyExtractor != nil {
		return c.cfg.KeyExtractor(msg, c.cfg.SessionKeyOffset)
	}

	if c.cfg.SessionKeyOffset < 0 || c.cfg.SessionKeyOffset >= len(msg.Body) {
		return nil, fmt.Errorf("%w: body length %d with offset %d", ErrSessionKeyNotLoaded, len(msg.Body), c.cfg.SessionKeyOffset)
	}

	return append([]byte(nil), msg.Body[c.cfg.SessionKeyOffset:]...), nil
}

func (c *AppClient) handleLoginAck(msg Message) error {
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

func (c *AppClient) handleLoginSuccess(msg Message) error {
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

func (c *AppClient) handleDisconnect(err error) {
	c.mu.Lock()
	c.authenticated = false
	c.sessionKeyLoaded = false
	c.awaitingKey = true
	c.mu.Unlock()

	if c.cfg.OnDisconnect != nil {
		c.cfg.OnDisconnect(err)
	}
}
