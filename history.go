package mt5

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/kahono0/mt5/models"
)

type TradeHistory = models.TradeHistory

type TradeHistoryResult struct {
	Value *TradeHistory
	Err   error
}

type TradeHistoryHandler func(TradeHistoryResult)

// RegisterTradeHistoryHandler registers a callback for decoded trade history.
// Typed handlers for this command share one history value per message; treat
// it and its nested slices as read-only.
func (c *AppClient) RegisterTradeHistoryHandler(handler TradeHistoryHandler) func() {
	if handler == nil {
		return func() {}
	}

	return c.On(CommandLoadHistory, func(result Result) {
		handler(TradeHistoryResult{Value: result.Value.(*TradeHistory), Err: result.Err})
	})
}

func decodeTradeHistoryMessage(msg Message) (*TradeHistory, error) {
	history := &TradeHistory{}
	if err := history.ParseFromBytes(msg.Body); err != nil {
		return nil, fmt.Errorf("failed to parse trade history: %w", err)
	}
	return history, nil
}

// RequestTradeHistory sends a request for trade history between the
// given timestamps. The client must be authenticated.
func (c *AppClient) RequestTradeHistory(from, to time.Time) error {
	if !c.IsAuthenticated() {
		return ErrNotAuthenticated
	}
	body := buildTradeHistoryPayload(from, to)
	if err := c.core.SendCommand(CommandLoadHistory, body); err != nil {
		return fmt.Errorf("failed to request trade history: %w", err)
	}
	return nil
}

func (c *AppClient) GetTradeHistory(ctx context.Context, from, to time.Time) (*TradeHistory, error) {
	if !c.IsAuthenticated() {
		return nil, ErrNotAuthenticated
	}

	ch := make(chan TradeHistoryResult, 1)

	cancel := c.RegisterTradeHistoryHandler(func(result TradeHistoryResult) {
		ch <- result
	})
	defer cancel()

	if err := c.RequestTradeHistory(from, to); err != nil {
		return nil, fmt.Errorf("failed to request trade history: %w", err)
	}

	select {
	case res := <-ch:
		return res.Value, res.Err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// BuildTradeHistoryPayload encodes a from/to timestamp pair into the
// binary payload format expected by the trade history command.
func buildTradeHistoryPayload(from, to time.Time) []byte {
	body := make([]byte, 8)
	binary.LittleEndian.PutUint32(body[0:4], uint32(from.Unix()))
	binary.LittleEndian.PutUint32(body[4:8], uint32(to.Unix()))
	return body
}
