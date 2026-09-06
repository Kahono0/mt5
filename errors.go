package mt5

import "errors"

// Low-level client errors.
var (
	// ErrClientClosed is returned when an operation is attempted on a closed client.
	ErrClientClosed = errors.New("client is closed")
	// ErrNotConnected is returned when the client has no active connection.
	ErrNotConnected = errors.New("client is not connected")
	// ErrMessageTooShort is returned when a received message is too short to parse.
	ErrMessageTooShort = errors.New("message too short")
	// ErrReadError is returned when a WebSocket read fails.
	ErrReadError = errors.New("read error")
)

// High-level AppClient errors.
var (
	// ErrInvalidConfig is returned when the AppClient configuration is incomplete.
	ErrInvalidConfig = errors.New("invalid mt5 config")
	// ErrNotAuthenticated is returned when an action requires authentication
	// but the client has not yet authenticated.
	ErrNotAuthenticated = errors.New("client is not authenticated")
	// ErrSessionKeyNotLoaded is returned when the session key has not been
	// extracted from the server's handshake message.
	ErrSessionKeyNotLoaded = errors.New("session key not loaded")
)
