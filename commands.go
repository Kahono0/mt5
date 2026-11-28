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

// typedCommand binds a wire command to the decoded value type and its decoder.
// One canonical instance is kept here for each typed command so all of its
// handlers use the same decoder.
type typedCommand[T any] struct {
	id     uint16
	decode func(Message) (T, error)
}

func newTypedCommand[T any](id uint16, decode func(Message) (T, error)) typedCommand[T] {
	return typedCommand[T]{id: id, decode: decode}
}

var (
	accountInfoCommand  = newTypedCommand(CommandRequestAccountInfo, decodeAccountInfoMessage)
	positionsCommand    = newTypedCommand(CommandRequestPositions, decodePositionsMessage)
	tradeHistoryCommand = newTypedCommand(CommandLoadHistory, decodeTradeHistoryMessage)
	querySymbolsCommand = newTypedCommand(CommandQuerySymbols, DecodeSymbols)
	tickUpdateCommand   = newTypedCommand(CommandTickUpdate, DecodeTicks)
	tradeEventCommand   = newTypedCommand(CommandTradeEvent, decodeTradeEventMessage)
	transactionCommand  = newTypedCommand(CommandTransactionEvent, decodeTransactionEventMessage)
)
