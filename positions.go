package mt5

import (
	"context"
	"fmt"

	"github.com/kahono0/mt5/models"
)

type Position = models.Position

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

type PositionsHandler func(positions []*Position)

func (c *AppClient) RegisterPositionsHandler(handler PositionsHandler, opts ...HandlerOption) func() {
	if handler == nil {
		return func() {}
	}

	return c.RegisterHandler(CommandRequestPositions, func(msg Message) error {
		positions, err := models.ParsePositionsFromBytes(msg.Body)
		if err != nil {
			return fmt.Errorf("failed to decode positions: %w", err)
		}

		handler(positions)
		return nil
	}, opts...)
}

func (c *AppClient) GetOpenPositions(ctx context.Context) ([]*Position, error) {
	if !c.IsAuthenticated() {
		return nil, ErrNotAuthenticated
	}

	type result struct {
		positions []*Position
		err       error
	}

	ch := make(chan result, 1)

	cancel := c.RegisterPositionsHandler(func(positions []*Position) {
		ch <- result{positions: positions, err: nil}
	}, WithErrorHandler(func(err error) {
		ch <- result{positions: nil, err: err}
	}))
	defer cancel()

	if err := c.RequestOpenPositions(); err != nil {
		return nil, fmt.Errorf("failed to request open positions: %w", err)
	}

	select {
	case res := <-ch:
		return res.positions, res.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
