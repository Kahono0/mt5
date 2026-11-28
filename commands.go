package mt5

// Protocol command identifiers sent over the MT5 binary protocol.

// Request commands
const (
	CommandRequestSessionKey  uint16 = 0
	CommandRequestAccountInfo uint16 = 3
	CommandRequestPositions   uint16 = 4
	CommandLoadHistory        uint16 = 5
	CommandQuerySymbols       uint16 = 6
	CommandSubscribeTicks     uint16 = 7
	CommandLoadSymbols        uint16 = 18
	CommandLogin              uint16 = 28
)

const (
	CommandTickUpdate       uint16 = 8
	CommandTransactionEvent uint16 = 10

	CommandSubmitTrade  uint16 = 12
	CommandLoginSuccess uint16 = 15
	CommandTradeEvent   uint16 = 19
	CommandPing         uint16 = 51
)

var decoders = [256]func(msg Message) (any, error){
	CommandRequestSessionKey:  func(msg Message) (any, error) { return msg, nil },
	CommandLogin:              func(msg Message) (any, error) { return msg, nil },
	CommandLoginSuccess:       func(msg Message) (any, error) { return msg, nil },
	CommandRequestAccountInfo: func(msg Message) (any, error) { return decodeAccountInfoMessage(msg) },
	CommandRequestPositions:   func(msg Message) (any, error) { return decodePositionsMessage(msg) },
	CommandLoadHistory:        func(msg Message) (any, error) { return decodeTradeHistoryMessage(msg) },
	CommandQuerySymbols:       func(msg Message) (any, error) { return DecodeSymbols(msg) },
	CommandTickUpdate:         func(msg Message) (any, error) { return DecodeTicks(msg) },
	CommandTradeEvent:         func(msg Message) (any, error) { return decodeTradeEventMessage(msg) },
	CommandTransactionEvent:   func(msg Message) (any, error) { return decodeTransactionEventMessage(msg) },
}
