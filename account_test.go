package mt5

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/kahono0/mt5/models"
)

func TestParseAccountInfo(t *testing.T) {
	data, err := os.ReadFile("testdata/account_info_response.bin")
	if err != nil {
		t.Fatalf("failed to read test data: %v", err)
	}

	st := time.Now()
	info, err := parseAccountInfo(data)
	if err != nil {
		t.Fatalf("failed to parse account info: %v", err)
	}

	t.Logf("parsed account info in %v", time.Since(st))

	infoParsedCorrectly(t, info)

}

func TestParseAccountInfoFromBytes(t *testing.T) {
	data, err := os.ReadFile("testdata/account_info_response.bin")
	if err != nil {
		t.Fatalf("failed to read test data: %v", err)
	}

	st := time.Now()
	info := &models.AccountInfo{}
	if err := info.ParseFromBytes(data); err != nil {
		t.Fatalf("failed to parse account info: %v", err)
	}

	t.Logf("parsed account info in %v", time.Since(st))

	infoParsedCorrectly(t, info)
}

// benchmark schema compiled
func BenchmarkParseAccountInfo(b *testing.B) {
	data, err := os.ReadFile("testdata/account_info_response.bin")
	if err != nil {
		b.Fatalf("failed to read test data: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := parseAccountInfo(data); err != nil {
			b.Fatalf("failed to parse account info: %v", err)
		}
	}
}

// benchmark from bytes
func BenchmarkParseAccountInfoFromBytes(b *testing.B) {
	data, err := os.ReadFile("testdata/account_info_response.bin")
	if err != nil {
		b.Fatalf("failed to read test data: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		info := &models.AccountInfo{}
		if err := info.ParseFromBytes(data); err != nil {
			b.Fatalf("failed to parse account info: %v", err)
		}
	}
}

func infoParsedCorrectly(t *testing.T, info *models.AccountInfo) {
	if info.Core.ServerName != "MetaQuotes-Demo" {
		t.Fatalf("expected account name 'TestAccount', got '%s'", info.Core.AccountName)
	}

	if len(info.TradeSettings) != 4 {
		t.Fatalf("expected 4 trade settings, got %d", len(info.TradeSettings))
	}

	if info.TradeSettings[0].SymbolPath != "Forex\\*" {
		t.Fatalf("expected first trade setting symbol path 'Forex\\*', got '%s'", info.TradeSettings[0].SymbolPath)
	}

	if info.TradeSettings[3].SymbolPath != "Nasdaq\\*" {
		t.Fatalf("expected fourth trade setting symbol path 'Nasdaq\\*', got '%s'", info.TradeSettings[3].SymbolPath)
	}
}

func asPrettyJSON(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("failed to marshal to JSON: %v", err)
	}
	return string(b)
}
