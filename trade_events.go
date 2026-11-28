package mt5

import "github.com/kahono0/mt5/models"

type TradeEvent = models.TradeEvent

type TradeEventResult struct {
	Value *TradeEvent
	Err   error
}

type TradeEventHandler func(TradeEventResult)

// RegisterTradeEventHandler registers a callback for decoded trade events.
// Typed handlers for this command share one event per message; treat it and
// its nested fields as read-only.
func (c *AppClient) RegisterTradeEventHandler(handler TradeEventHandler) func() {
	if handler == nil {
		return func() {}
	}

	return c.On(CommandTradeEvent, func(result Result) {
		handler(TradeEventResult{Value: result.Value.(*TradeEvent), Err: result.Err})
	})
}

func decodeTradeEventMessage(msg Message) (*TradeEvent, error) {
	event := &TradeEvent{}
	if err := event.ParseFromBytes(msg.Body); err != nil {
		return nil, err
	}
	return event, nil
}
