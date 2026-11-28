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
	// PingTimeout is the maximum time to wait for a ping acknowledgment.
	// Defaults to 30 seconds.
	// If a ping is not acknowledged within this time, the client will consider the connection lost and attempt to reconnect.
	PingTimeout time.Duration
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

	mu sync.RWMutex
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

	core, err := New(Config{
		URL:       cfg.URL,
		AccountID: cfg.AccountID,
		ServerURL: cfg.ServerURL,
		Password:  cfg.Password,
		DeviceID:  cfg.DeviceID,

		SessionKeyOffset: cfg.SessionKeyOffset,
		KeyExtractor:     cfg.KeyExtractor,
		SessionKeyBody:   cfg.SessionKeyBody,

		Dialer:               cfg.Dialer,
		RequestHeader:        cfg.RequestHeader,
		SkipCommands:         cfg.SkipCommands,
		PingInterval:         cfg.PingInterval,
		PingTimeout:          cfg.PingTimeout,
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
		cfg:  cfg,
		core: core,
	}
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
	return c.core.Run(ctx)
}

// Close gracefully shuts down the client.
func (c *AppClient) Close() error {
	return c.core.Close()
}

// RegisterHandler registers a handler for a specific command ID on the
// underlying [Client]. Use [WithErrorHandler] to route errors from this
// handler to a dedicated callback. Returns a function that, when called,
// removes the handler.
func (c *AppClient) RegisterHandler(command uint16, handler MessageHandler, opts ...HandlerOption) func() {
	return c.core.On(command, handler, opts...)
}

// RegisterAnyHandler registers a handler that is called for every
// received message. Use [WithErrorHandler] to route errors from this
// handler to a dedicated callback. Returns a function that, when called,
// removes the handler.
func (c *AppClient) RegisterAnyHandler(handler MessageHandler, opts ...HandlerOption) func() {
	return c.core.OnAny(handler, opts...)
}

// IsAuthenticated reports whether the client has completed the login
// handshake.
func (c *AppClient) IsAuthenticated() bool {
	return c.core.IsAuthenticated()
}

// WaitUntilAuthenticated blocks until the client is authenticated or
// the context is cancelled.
func (c *AppClient) WaitUntilAuthenticated(ctx context.Context) error {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		if c.core.IsAuthenticated() {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
