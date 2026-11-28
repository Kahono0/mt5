package mt5

import (
	"context"
	"testing"
	"time"

	"github.com/kahono0/mt5/models"
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
	loginReq := &models.LoginRequest{
		ServerURL: "server.example.com",
		Password:  "secret",
		AccountID: 12345,
		DeviceID:  "device-1",
	}
	body, err := loginReq.Serialize()
	if err != nil {
		t.Fatalf("unexpected serialize error: %v", err)
	}
	if len(body) == 0 {
		t.Fatalf("expected non-empty login payload")
	}
}

func TestRegisterTickHandlerNilDoesNotPanic(t *testing.T) {
	client, err := NewApp(validAppConfig())
	if err != nil {
		t.Fatalf("unexpected new error: %v", err)
	}

	client.RegisterTickHandler(nil)
}
