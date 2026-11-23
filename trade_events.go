package mt5

import "github.com/kahono0/mt5/models"

type TradeEventHandler func(event *models.TradeEvent)

func (c *AppClient) RegisterTradeEventHandler(handler TradeEventHandler, opts ...HandlerOption) func() {
	if handler == nil {
		return func() {}
	}

	return c.RegisterHandler(CommandTradeEvent, func(msg Message) error {
		event := &models.TradeEvent{}
		if err := event.ParseFromBytes(msg.Body); err != nil {
			return err
		}
		handler(event)

		return nil
	}, opts...)
}
