package mt5

import (
	"os"
	"testing"
	"time"
)

func TestParseTradeHistory(t *testing.T) {
	data, err := os.ReadFile("testdata/trade_history.bin")
	if err != nil {
		t.Fatalf("failed to read test data: %v", err)
	}

	t.Logf("data length: %d bytes", len(data))

	st := time.Now()
	history, err := parseTradeHistory(data)
	if err != nil {
		t.Fatalf("failed to parse trade history: %v", err)
	}

	t.Logf("parsed trade history in %v", time.Since(st))

	if len(history.Deals) != 52 {
		t.Fatalf("expected 52 deals, got %d", len(history.Deals))
	}

	if len(history.Orders) != 52 {
		t.Fatalf("expected 52 orders, got %d", len(history.Orders))
	}
}
