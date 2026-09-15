package mt5

// Protocol command identifiers sent over the MT5 binary protocol.

// Command identifies a message command and the function that decodes its
// payload. Its type parameter associates a command with the value delivered
// to handlers registered through [On].
//
// Use the predefined commands below instead of constructing one directly.
type Command[T any] struct {
	ID     uint16
	Decode func(Message) (T, error)
}

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
	CommandTickUpdate    uint16 = 8
	CommandOrderPosition uint16 = 10
	CommandSubmitTrade   uint16 = 12
	CommandLoginSuccess  uint16 = 15
	CommandTradeEvent    uint16 = 19
	CommandPing          uint16 = 51
)

// Predefined decoded message commands.
//
// These values retain the output type of their respective decoders, allowing
// [On] to check a handler's input type at compile time.
var (
	TickUpdate = Command[[]Tick]{
		ID:     CommandTickUpdate,
		Decode: DecodeTicks,
	}
	Positions = Command[[]Position]{
		ID:     CommandRequestPositions,
		Decode: DecodePositions,
	}
	AccountInfoResponse = Command[*AccountInfo]{
		ID:     CommandRequestAccountInfo,
		Decode: DecodeAccountInfo,
	}
	SymbolsResponse = Command[*SymbolList]{
		ID:     CommandQuerySymbols,
		Decode: DecodeSymbols,
	}
)
