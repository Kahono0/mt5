package mt5

import (
	"os"
	"testing"
	"time"

	"github.com/kahono0/mt5/internal/ec"
)

func TestParseTradeHistory(t *testing.T) {
	data, err := os.ReadFile("testdata/trade_history.bin")
	if err != nil {
		t.Fatalf("failed to read test data: %v", err)
	}

	t.Logf("data length: %d bytes", len(data))

	st := time.Now()
	history, err := ec.ParseGen(data, &TradeHistory{})
	if err != nil {
		t.Fatalf("failed to parse trade history: %v", err)
	}

	//val, err := ec.ParseToValue(&TradeHistory{}, data)
	//if err != nil {
	//	t.Fatalf("failed to parse trade history: %v", err)
	//}

	//history := val.(*TradeHistory)

	t.Logf("parsed trade history in %v", time.Since(st))

	if len(history.Deals) != 52 {
		t.Fatalf("expected 52 deals, got %d", len(history.Deals))
	}

	if len(history.Orders) != 52 {
		t.Fatalf("expected 52 orders, got %d", len(history.Orders))
	}
}

// test using compiled schema
func TestParseTradeHistoryCompiled(t *testing.T) {
	data, err := os.ReadFile("testdata/trade_history.bin")
	if err != nil {
		t.Fatalf("failed to read test data: %v", err)
	}

	t.Logf("data length: %d bytes", len(data))

	st := time.Now()
	fields := (&TradeHistory{}).Fields()

	schema := ec.CompileSchema(fields)

	parsed, err := schema.Parse(data, 0)
	if err != nil {
		t.Fatalf("failed to parse trade history: %v", err)
	}

	history := (&TradeHistory{}).FromValues(parsed).(*TradeHistory)

	t.Logf("parsed trade history in %v", time.Since(st))

	if len(history.Deals) != 52 {
		t.Fatalf("expected 52 deals, got %d", len(history.Deals))
	}

	if len(history.Orders) != 52 {
		t.Fatalf("expected 52 orders, got %d", len(history.Orders))
	}
}

// test parsing trade history using handwritten ParseFromBytes method
func TestParseTradeHistoryFromBytes(t *testing.T) {
	data, err := os.ReadFile("testdata/trade_history.bin")
	if err != nil {
		t.Fatalf("failed to read test data: %v", err)
	}

	t.Logf("data length: %d bytes", len(data))

	st := time.Now()
	history := &TradeHistory{}
	err = history.ParseFromBytes(data)
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

// benchmark parsing trade history using compiled schema
func BenchmarkParseTradeHistoryCompiled(b *testing.B) {
	data, err := os.ReadFile("testdata/trade_history.bin")
	if err != nil {
		b.Fatalf("failed to read test data: %v", err)
	}

	fields := (&TradeHistory{}).Fields()
	schema := ec.CompileSchema(fields)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := schema.Parse(data, 0)
		if err != nil {
			b.Fatalf("failed to parse trade history: %v", err)
		}
	}
}

// benchmark parsing trade history
func BenchmarkParseTradeHistoryGeneric(b *testing.B) {
	data, err := os.ReadFile("testdata/trade_history.bin")
	if err != nil {
		b.Fatalf("failed to read test data: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ec.ParseGen(data, &TradeHistory{})
		if err != nil {
			b.Fatalf("failed to parse trade history: %v", err)
		}
	}
}

func BenchmarkParseTradeHistoryToValue(b *testing.B) {
	data, err := os.ReadFile("testdata/trade_history.bin")
	if err != nil {
		b.Fatalf("failed to read test data: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ec.ParseToValue(&TradeHistory{}, data)
		if err != nil {
			b.Fatalf("failed to parse trade history: %v", err)
		}
	}
}
