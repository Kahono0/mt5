package mt5

// Protocol command identifiers sent over the MT5 binary protocol.
const (
	CommandRequestSessionKey uint16 = 0
	CommandRequestAccount    uint16 = 3
	CommandRequestPositions  uint16 = 4
	CommandLoadHistory       uint16 = 5
	CommandSubscribeTicks    uint16 = 7
	CommandTickUpdate        uint16 = 8
	CommandOrderPosition     uint16 = 10
	CommandSubmitTrade       uint16 = 12
	CommandLoginSuccess      uint16 = 15
	CommandLoadSymbols       uint16 = 18
	CommandTradeEvent        uint16 = 19
	CommandLogin             uint16 = 28
	CommandPing              uint16 = 51
)
