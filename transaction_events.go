package mt5

import "github.com/kahono0/mt5/models"

type EnvelopeTransaction = models.EnvelopeTransaction

type TransactionEventResult struct {
	Value *EnvelopeTransaction
	Err   error
}

type TransactionEventHandler func(TransactionEventResult)

// RegisterTransactionEventHandler registers a callback for decoded
// transaction events. Typed handlers for this command share one event per
// message; treat it and its nested fields as read-only.
func (c *AppClient) RegisterTransactionEventHandler(handler TransactionEventHandler) func() {
	if handler == nil {
		return func() {}
	}

	return c.On(CommandTransactionEvent, func(result Result) {
		handler(TransactionEventResult{Value: result.Value.(*EnvelopeTransaction), Err: result.Err})
	})
}

func decodeTransactionEventMessage(msg Message) (*EnvelopeTransaction, error) {
	event := &EnvelopeTransaction{}
	if err := event.ParseFromBytes(msg.Body); err != nil {
		return nil, err
	}
	return event, nil
}
