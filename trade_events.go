package mt5

import "github.com/kahono0/mt5/models"

type TradeEvent = models.TradeEvent

type TradeEventHandler func(event *TradeEvent)

// RegisterTradeEventHandler registers a callback for decoded trade events.
// Typed handlers for this command share one event per message; treat it and
// its nested fields as read-only.
func (c *AppClient) RegisterTradeEventHandler(handler TradeEventHandler, opts ...HandlerOption) func() {
	if handler == nil {
		return func() {}
	}

	return registerTyped(c.core, tradeEventCommand, func(event *TradeEvent, _ Message) error {
		handler(event)

		return nil
	}, opts...)
}

func decodeTradeEventMessage(msg Message) (*TradeEvent, error) {
	event := &TradeEvent{}
	if err := event.ParseFromBytes(msg.Body); err != nil {
		return nil, err
	}
	return event, nil
}
