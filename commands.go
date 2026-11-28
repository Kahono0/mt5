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
type Command[T any] struct {
	id     uint16
	decode func(Message) (T, error)
}

func newTypedCommand[T any](id uint16, decode func(Message) (T, error)) Command[T] {
	return Command[T]{id: id, decode: decode}
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

// AccountInfoCommand returns the typed command specification for account
// information messages.
func AccountInfoCommand() Command[*AccountInfo] { return accountInfoCommand }

// PositionsCommand returns the typed command specification for position
// messages.
func PositionsCommand() Command[[]*Position] { return positionsCommand }

// TradeHistoryCommand returns the typed command specification for trade
// history messages.
func TradeHistoryCommand() Command[*TradeHistory] { return tradeHistoryCommand }

// QuerySymbolsCommand returns the typed command specification for symbol
// list messages.
func QuerySymbolsCommand() Command[*SymbolList] { return querySymbolsCommand }

// TickUpdateCommand returns the typed command specification for tick updates.
func TickUpdateCommand() Command[[]Tick] { return tickUpdateCommand }

// TradeEventCommand returns the typed command specification for trade events.
func TradeEventCommand() Command[*TradeEvent] { return tradeEventCommand }

// TransactionEventCommand returns the typed command specification for
// transaction events.
func TransactionEventCommand() Command[*EnvelopeTransaction] { return transactionCommand }
