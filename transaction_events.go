package mt5

import "github.com/kahono0/mt5/models"

type EnvelopeTransaction = models.EnvelopeTransaction

type TransactionEventHandler func(event *EnvelopeTransaction)

func (c *AppClient) RegisterTransactionEventHandler(handler TransactionEventHandler, opts ...HandlerOption) func() {
	if handler == nil {
		return func() {}
	}

	return c.RegisterHandler(CommandTransactionEvent, func(m Message) error {
		event := &EnvelopeTransaction{}
		if err := event.ParseFromBytes(m.Body); err != nil {
			return err
		}
		handler(event)
		return nil
	}, opts...)
}
