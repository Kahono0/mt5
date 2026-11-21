package mt5

import (
	"context"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/kahono0/mt5/models"
)

type TradeHistory = models.TradeHistory

type TradeHistoryHandler func(history *TradeHistory)

func (c *AppClient) RegisterTradeHistoryHandler(handler TradeHistoryHandler, opts ...HandlerOption) func() {
	if handler == nil {
		return func() {}
	}

	return c.RegisterParsableHandler(CommandLoadHistory, &models.TradeHistory{}, func(parsed any) {
		history := parsed.(*models.TradeHistory)
		handler(history)
	}, opts...)
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

	type result struct {
		history *TradeHistory
		err     error
	}

	ch := make(chan result, 1)

	cancel := c.RegisterTradeHistoryHandler(func(history *TradeHistory) {
		ch <- result{history: history, err: nil}
	}, WithErrorHandler(func(err error) {
		ch <- result{history: nil, err: err}
	}))
	defer cancel()

	if err := c.RequestTradeHistory(from, to); err != nil {
		return nil, fmt.Errorf("failed to request trade history: %w", err)
	}

	select {
	case res := <-ch:
		return res.history, res.err
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
