package mt5

import (
	"context"
	"fmt"

	"github.com/kahono0/mt5/models"
)

type AccountInfo = models.AccountInfo

// AccountInfoHandler is a function that processes decoded account data.
// Handle errors within the callback; only framework-level decode
// errors are routed to [Config.OnError] and the error channel.
type AccountInfoHandler func(info *AccountInfo, msg Message)

// RequestAccountInfo sends a request for account information.
// The client must be authenticated.
func (c *AppClient) RequestAccountInfo() error {
	if !c.IsAuthenticated() {
		return ErrNotAuthenticated
	}
	if err := c.core.SendCommand(CommandRequestAccountInfo, nil); err != nil {
		return fmt.Errorf("failed to request account info: %w", err)
	}
	return nil
}

// GetAccountInfo sends a request for account information and blocks until
// the response is received or the context is cancelled. This is a synchronous
// wrapper around [RequestAccountInfo] and [RegisterAccountInfoHandler].
func (c *AppClient) GetAccountInfo(ctx context.Context) (*AccountInfo, error) {
	if !c.IsAuthenticated() {
		return nil, ErrNotAuthenticated
	}

	type result struct {
		info *AccountInfo
		err  error
	}
	ch := make(chan result, 1)

	cancel := c.RegisterAccountInfoHandler(func(info *AccountInfo, msg Message) {
		ch <- result{info, nil}
	}, WithErrorHandler(func(err error) {
		ch <- result{nil, err}
	}))
	defer cancel()

	if err := c.RequestAccountInfo(); err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r := <-ch:
		return r.info, r.err
	}
}

// RegisterAccountInfoHandler registers a handler that receives decoded
// account information. Use [WithErrorHandler] to route errors from this
// handler to a dedicated callback. Returns a function that, when called,
// removes the handler.
func (c *AppClient) RegisterAccountInfoHandler(handler AccountInfoHandler, opts ...HandlerOption) func() {
	if handler == nil {
		return func() {}
	}

	return c.RegisterHandler(CommandRequestAccountInfo, func(msg Message) error {
		info := &models.AccountInfo{}
		if err := info.ParseFromBytes(msg.Body); err != nil {
			return fmt.Errorf("failed to parse account info: %w", err)
		}
		handler(info, msg)
		return nil
	}, opts...)
}

func parseAccountInfo(data []byte) (*AccountInfo, error) {
	info := &AccountInfo{}
	if err := info.ParseFromBytes(data); err != nil {
		return nil, fmt.Errorf("failed to parse account info: %w", err)
	}
	return info, nil
}
