package mt5

import (
	"context"
	"fmt"

	"github.com/kahono0/mt5/models"
)

type Position = models.Position

type PositionsResult struct {
	Value []*Position
	Err   error
}

// RequestOpenPositions sends a request for open positions.
// The client must be authenticated.
func (c *AppClient) RequestOpenPositions() error {
	if !c.IsAuthenticated() {
		return ErrNotAuthenticated
	}
	if err := c.core.SendCommand(CommandRequestPositions, nil); err != nil {
		return fmt.Errorf("failed to request open positions: %w", err)
	}
	return nil
}

type PositionsHandler func(PositionsResult)

// RegisterPositionsHandler registers a callback for decoded positions.
// Typed handlers for this command share one positions slice per message;
// treat the slice and its entries as read-only.
func (c *AppClient) RegisterPositionsHandler(handler PositionsHandler) func() {
	if handler == nil {
		return func() {}
	}

	return c.On(CommandRequestPositions, func(result Result) {
		handler(PositionsResult{Value: result.Value.([]*Position), Err: result.Err})
	})
}

func decodePositionsMessage(msg Message) ([]*Position, error) {
	positions, err := models.ParsePositionsFromBytes(msg.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode positions: %w", err)
	}
	return positions, nil
}

func (c *AppClient) GetOpenPositions(ctx context.Context) ([]*Position, error) {
	if !c.IsAuthenticated() {
		return nil, ErrNotAuthenticated
	}

	ch := make(chan PositionsResult, 1)

	cancel := c.RegisterPositionsHandler(func(pr PositionsResult) {
		ch <- pr
	})

	defer cancel()

	if err := c.RequestOpenPositions(); err != nil {
		return nil, fmt.Errorf("failed to request open positions: %w", err)
	}

	select {
	case res := <-ch:
		return res.Value, res.Err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
