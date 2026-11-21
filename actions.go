package mt5

import (
	"encoding/binary"
	"fmt"

	"github.com/kahono0/mt5/internal/ec"
	"github.com/kahono0/mt5/internal/login"
)

// SendCommand sends a raw command through the underlying [Client].
func (c *AppClient) SendCommand(command uint16, body []byte) error {
	return c.core.SendCommand(command, body)
}

// Login sends the login command to the server. This is called
// automatically after the session key is loaded, but can be called
// manually if needed.
func (c *AppClient) Login() error {
	body, err := buildLoginRequest(c.cfg.ServerURL, c.cfg.Password, c.cfg.AccountID, c.cfg.DeviceID)
	if err != nil {
		return err
	}

	if err := c.core.SendCommand(CommandLogin, body); err != nil {
		return fmt.Errorf("failed to send login command: %w", err)
	}
	return nil
}

// SubscribeTicks subscribes to real-time tick updates for the given
// symbol IDs. The client must be authenticated.
func (c *AppClient) SubscribeTicks(symbolIDs []uint32) error {
	if !c.IsAuthenticated() {
		return ErrNotAuthenticated
	}
	body := BuildTickSubscriptionPayload(symbolIDs)
	if err := c.core.SendCommand(CommandSubscribeTicks, body); err != nil {
		return fmt.Errorf("failed to subscribe ticks: %w", err)
	}
	return nil
}

// BuildTickSubscriptionPayload encodes a list of symbol IDs into the
// binary payload format expected by the tick subscription command.
func BuildTickSubscriptionPayload(symbolIDs []uint32) []byte {
	buf := make([]byte, 4+len(symbolIDs)*4)
	binary.LittleEndian.PutUint32(buf[0:4], uint32(len(symbolIDs)))
	offset := 4
	for _, id := range symbolIDs {
		binary.LittleEndian.PutUint32(buf[offset:offset+4], id)
		offset += 4
	}
	return buf
}

func buildLoginRequest(
	serverURL, password string,
	accountID uint64,
	deviceID string,
) ([]byte, error) {
	fields := login.GenerateLoginFields(serverURL, password, accountID, deviceID)
	body, err := ec.Serialize(fields)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize login fields: %w", err)
	}
	return body, nil
}
