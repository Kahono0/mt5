package mt5

import (
	"context"
	"errors"
	"testing"
	"time"
)

func validAppConfig() AppConfig {
	return AppConfig{
		URL:       "ws://localhost:8080/ws",
		ServerURL: "server.example.com",
		AccountID: 12345,
		Password:  "secret",
		DeviceID:  "device-1",
	}
}

func TestNewAppRequiresMandatoryFields(t *testing.T) {
	cfg := validAppConfig()
	cfg.URL = ""
	if _, err := NewApp(cfg); err == nil {
		t.Fatalf("expected error when URL is missing")
	}

	cfg = validAppConfig()
	cfg.ServerURL = ""
	if _, err := NewApp(cfg); err == nil {
		t.Fatalf("expected error when ServerURL is missing")
	}

	cfg = validAppConfig()
	cfg.AccountID = 0
	if _, err := NewApp(cfg); err == nil {
		t.Fatalf("expected error when AccountID is missing")
	}

	cfg = validAppConfig()
	cfg.Password = ""
	if _, err := NewApp(cfg); err == nil {
		t.Fatalf("expected error when Password is missing")
	}

	cfg = validAppConfig()
	cfg.DeviceID = ""
	if _, err := NewApp(cfg); err == nil {
		t.Fatalf("expected error when DeviceID is missing")
	}
}

func TestWaitUntilAuthenticatedHonorsContext(t *testing.T) {
	client, err := NewApp(validAppConfig())
	if err != nil {
		t.Fatalf("unexpected new error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()

	if err := client.WaitUntilAuthenticated(ctx); err == nil {
		t.Fatalf("expected context timeout error")
	}
}

func TestLoginRequestSerializes(t *testing.T) {
	body, err := buildLoginRequest("server.example.com", "secret", 12345, "device-1")
	if err != nil {
		t.Fatalf("unexpected serialize error: %v", err)
	}
	if len(body) == 0 {
		t.Fatalf("expected non-empty login payload")
	}
}

func TestHandleDisconnectResetsAuthState(t *testing.T) {
	client, err := NewApp(validAppConfig())
	if err != nil {
		t.Fatalf("unexpected new error: %v", err)
	}

	client.mu.Lock()
	client.authenticated = true
	client.sessionKeyLoaded = true
	client.awaitingKey = false
	client.mu.Unlock()

	client.handleDisconnect(errors.New("read error"))

	client.mu.RLock()
	defer client.mu.RUnlock()
	if client.authenticated {
		t.Fatalf("expected authenticated=false after disconnect")
	}
	if client.sessionKeyLoaded {
		t.Fatalf("expected sessionKeyLoaded=false after disconnect")
	}
	if !client.awaitingKey {
		t.Fatalf("expected awaitingKey=true after disconnect")
	}
}

func TestHandleDisconnectInvokesCallback(t *testing.T) {
	called := false
	var gotErr error

	client, err := NewApp(AppConfig{
		URL:       "ws://localhost:8080/ws",
		ServerURL: "server.example.com",
		AccountID: 12345,
		Password:  "secret",
		DeviceID:  "device-1",
		OnDisconnect: func(err error) {
			called = true
			gotErr = err
		},
	})
	if err != nil {
		t.Fatalf("unexpected new error: %v", err)
	}

	expected := errors.New("socket read failed")
	client.handleDisconnect(expected)

	if !called {
		t.Fatalf("expected OnDisconnect callback to be called")
	}
	if !errors.Is(gotErr, expected) {
		t.Fatalf("expected callback error to match disconnect error")
	}
}

func TestRegisterTickHandlerNilDoesNotPanic(t *testing.T) {
	client, err := NewApp(validAppConfig())
	if err != nil {
		t.Fatalf("unexpected new error: %v", err)
	}

	client.RegisterTickHandler(nil)
}
