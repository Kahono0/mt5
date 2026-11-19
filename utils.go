package mt5

import (
	"fmt"
	"math/rand"
)

func RandomDeviceID() string {
	r := rand.Intn(1000000)
	return fmt.Sprintf("Linux x86_64;1;en-US;1920x1080;%d", r)
}
