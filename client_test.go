package mt5

import "testing"

func TestNewAppliesDefaults(t *testing.T) {
	c, err := New(Config{URL: "ws://localhost:8080/ws"})
	if err != nil {
		t.Fatalf("unexpected new error: %v", err)
	}

	if c.cfg.PingInterval <= 0 {
		t.Fatalf("expected positive ping interval")
	}

	if c.cfg.PingCommand == 0 {
		t.Fatalf("expected non-zero ping command")
	}

	if c.cfg.Codec == nil || c.cfg.Crypto == nil {
		t.Fatalf("expected default codec and crypto")
	}
}

func TestNewRequiresURL(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatalf("expected url required error")
	}
}

func TestLoadSessionKeyFromMessageOffsetValidation(t *testing.T) {
	c, err := New(Config{URL: "ws://localhost:8080/ws"})
	if err != nil {
		t.Fatalf("unexpected new error: %v", err)
	}

	msg := Message{Body: []byte{1, 2, 3}}
	if err := c.LoadSessionKeyFromMessage(msg, 3); err == nil {
		t.Fatalf("expected invalid offset error")
	}
}

func TestNewRetainsLifecycleHooks(t *testing.T) {
	onConnectCalled := false
	onDisconnectCalled := false

	c, err := New(Config{
		URL: "ws://localhost:8080/ws",
		OnConnect: func(client *Client, reconnect bool) error {
			onConnectCalled = true
			_ = reconnect
			_ = client
			return nil
		},
		OnDisconnect: func(client *Client, err error) {
			onDisconnectCalled = true
			_ = client
			_ = err
		},
	})
	if err != nil {
		t.Fatalf("unexpected new error: %v", err)
	}

	if c.cfg.OnConnect == nil || c.cfg.OnDisconnect == nil {
		t.Fatalf("expected lifecycle hooks to be retained")
	}

	if onConnectCalled || onDisconnectCalled {
		t.Fatalf("hooks should not be called during New")
	}
}
