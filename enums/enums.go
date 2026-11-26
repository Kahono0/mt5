package enums

type TradeReturnCode uint32

const (
	TRADE_RETCODE_REQUOTE        TradeReturnCode = 10004 // Requote
	TRADE_RETCODE_REJECT         TradeReturnCode = 10006 // Request rejected
	TRADE_RETCODE_CANCEL         TradeReturnCode = 10007 // Request canceled by trader
	TRADE_RETCODE_PLACED         TradeReturnCode = 10008 // Order placed
	TRADE_RETCODE_DONE           TradeReturnCode = 10009 // Request completed
	TRADE_RETCODE_DONE_PARTIAL   TradeReturnCode = 10010 // Only part of the request was completed
	TRADE_RETCODE_ERROR          TradeReturnCode = 10011 // Request processing error
	TRADE_RETCODE_TIMEOUT        TradeReturnCode = 10012 // Request canceled by timeout
	TRADE_RETCODE_INVALID        TradeReturnCode = 10013 // Invalid request
	TRADE_RETCODE_INVALID_VOLUME TradeReturnCode = 10014 // Invalid volume in the request
	TRADE_RETCODE_INVALID_PRICE  TradeReturnCode = 10015 // Invalid price in the request
	TRADE_RETCODE_INVALID_STOPS  TradeReturnCode = 10016 // Invalid stops in the request
)

type TradeAction uint32

const (
	TRADE_ACTION_DEAL     TradeAction = 3
	TRADE_ACTION_PENDING  TradeAction = 4
	TRADE_ACTION_SLTP     TradeAction = 5
	TRADE_ACTION_MODIFY   TradeAction = 6
	TRADE_ACTION_REMOVE   TradeAction = 7
	TRADE_ACTION_CLOSE_BY TradeAction = 8
)

type OrderType uint32

const (
	ORDER_TYPE_BUY             OrderType = 0
	ORDER_TYPE_SELL            OrderType = 1
	ORDER_TYPE_BUY_LIMIT       OrderType = 2
	ORDER_TYPE_SELL_LIMIT      OrderType = 3
	ORDER_TYPE_BUY_STOP        OrderType = 4
	ORDER_TYPE_SELL_STOP       OrderType = 5
	ORDER_TYPE_BUY_STOP_LIMIT  OrderType = 6
	ORDER_TYPE_SELL_STOP_LIMIT OrderType = 7
	ORDER_TYPE_CLOSE_BY        OrderType = 8
)
