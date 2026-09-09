package mt5

import (
	"errors"
	"fmt"
)

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

// ServerError is returned when the MT5 server rejects a request.
// Use [errors.As] to extract the command and response code programmatically.
type ServerError struct {
	// Command is the protocol command that was rejected.
	Command uint16
	// Code is the server's response/status code.
	Code uint8
	// Msg is a human-readable description of the failure.
	Msg string
}

func (e *ServerError) Error() string {
	return fmt.Sprintf("server error (cmd=%d, code=%d): %s", e.Command, e.Code, e.Msg)
}

// CommandMismatchError is returned when a received message has an
// unexpected command ID. Use [errors.As] to inspect the expected and
// actual command values.
type CommandMismatchError struct {
	// Expected is the command ID that was expected.
	Expected uint16
	// Got is the command ID that was received.
	Got uint16
}

func (e *CommandMismatchError) Error() string {
	return fmt.Sprintf("unexpected command: expected %d, got %d", e.Expected, e.Got)
}
