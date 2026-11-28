package mt5

import "github.com/kahono0/mt5/models"

type EnvelopeTransaction = models.EnvelopeTransaction

type TransactionEventHandler func(event *EnvelopeTransaction)

// RegisterTransactionEventHandler registers a callback for decoded
// transaction events. Typed handlers for this command share one event per
// message; treat it and its nested fields as read-only.
func (c *AppClient) RegisterTransactionEventHandler(handler TransactionEventHandler, opts ...HandlerOption) func() {
	if handler == nil {
		return func() {}
	}

	return registerTyped(c.core, transactionCommand, func(event *EnvelopeTransaction, _ Message) error {
		handler(event)
		return nil
	}, opts...)
}

func decodeTransactionEventMessage(msg Message) (*EnvelopeTransaction, error) {
	event := &EnvelopeTransaction{}
	if err := event.ParseFromBytes(msg.Body); err != nil {
		return nil, err
	}
	return event, nil
}
