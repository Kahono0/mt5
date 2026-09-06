package login

import (
	"github.com/kahono0/mt5/internal/ec"
	"github.com/kahono0/mt5/internal/mthelpers"
)

func emptyBuffer(length int) []byte {
	return make([]byte, length)
}

// GenerateLoginFields creates the same field list as in JS
func GenerateLoginFields(server, password string, login uint64, deviceID string) []ec.Field {
	urlLength := uint32(len(server))
	devID := mthelpers.DeviceIDFromString(deviceID)
	fields := []ec.Field{
		{PropType: 6, PropValue: uint32(0)},                          // 1. Int | value: 0
		{PropType: 11, PropValue: password, PropLength: 64},          // 2. String | password
		{PropType: 11, PropValue: "", PropLength: 128},               // 3. String | ""
		{PropType: 12, PropValue: devID, PropLength: 16},             // 4. Bytes | len 16
		{PropType: 11, PropValue: "", PropLength: 64},                // 5. String | ""
		{PropType: 11, PropValue: "", PropLength: 64},                // 6. String | ""
		{PropType: 18, PropValue: uint64(0)},                         // 7. BigInt | 0
		{PropType: 11, PropValue: "", PropLength: 128},               // 8. String | ""
		{PropType: 6, PropValue: urlLength},                          // 9. Int | 18
		{PropType: 11, PropValue: server, PropLength: 256},           // 10. String
		{PropType: 18, PropValue: login},                             // 11. BigInt | login id
		{PropType: 12, PropValue: emptyBuffer(160), PropLength: 160}, // 12. Bytes | len 160
		{PropType: 18, PropValue: uint64(0)},                         // 13. BigInt | session = 0
	}
	return fields
}
