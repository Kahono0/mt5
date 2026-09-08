package mt5

// Protocol command identifiers sent over the MT5 binary protocol.

// Request commands
const (
	CommandRequestSessionKey uint16 = 0
	CommandRequestAccount    uint16 = 3
	CommandRequestPositions  uint16 = 4
	CommandLoadHistory       uint16 = 5
	CommandQuerySymbols      uint16 = 6
	CommandSubscribeTicks    uint16 = 7
	CommandLoadSymbols       uint16 = 18
	CommandLogin             uint16 = 28
)

const (
	CommandTickUpdate    uint16 = 8
	CommandOrderPosition uint16 = 10
	CommandSubmitTrade   uint16 = 12
	CommandLoginSuccess  uint16 = 15
	CommandTradeEvent    uint16 = 19
	CommandPing          uint16 = 51
)
