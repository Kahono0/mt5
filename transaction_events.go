package mt5

import "github.com/kahono0/mt5/models"

type TransactionEventHandler func(event *models.EnvelopeTransaction)

func (c *AppClient) RegisterTransactionEventHandler(handler TransactionEventHandler, opts ...HandlerOption) func() {
	if handler == nil {
		return func() {}
	}

	return c.RegisterHandler(CommandTransactionEvent, func(m Message) error {
		event := &models.EnvelopeTransaction{}
		if err := event.ParseFromBytes(m.Body); err != nil {
			return err
		}
		handler(event)
		return nil
	}, opts...)
}
