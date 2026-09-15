package mt5

import (
	"errors"
	"testing"
)

func TestOnDecodesMessageBeforeCallingHandler(t *testing.T) {
	client, err := New(Config{URL: "ws://localhost:8080/ws"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	called := false
	cancel := client.OnT(TickUpdate, func(ticks []Tick, msg Message) error {
		called = true
		if msg.Command != CommandTickUpdate {
			t.Errorf("message command = %d, want %d", msg.Command, CommandTickUpdate)
		}
		if len(ticks) != 0 {
			t.Errorf("tick count = %d, want 0", len(ticks))
		}
		return nil
	})

	client.dispatch(Message{Command: CommandTickUpdate})
	if !called {
		t.Fatal("decoded handler was not called")
	}

	called = false
	cancel()
	client.dispatch(Message{Command: CommandTickUpdate})
	if called {
		t.Fatal("decoded handler was called after cancellation")
	}
}

func TestOnRoutesDecodeErrorsWithoutCallingHandler(t *testing.T) {
	client, err := New(Config{URL: "ws://localhost:8080/ws"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	called := false
	decodeErr := errors.New("invalid payload")
	badCommand := Command[struct{}]{
		ID: CommandRequestAccountInfo,
		Decode: func(Message) (struct{}, error) {
			return struct{}{}, decodeErr
		},
	}
	client.OnT(badCommand, func(info struct{}, msg Message) error {
		called = true
		return nil
	})

	client.dispatch(Message{Command: CommandRequestAccountInfo})
	if called {
		t.Fatal("handler was called after a decoding error")
	}

	select {
	case err := <-client.Errors():
		if !errors.Is(err, decodeErr) {
			t.Fatalf("published error = %v, want %v", err, decodeErr)
		}
	default:
		t.Fatal("expected decode error to be published")
	}
}
