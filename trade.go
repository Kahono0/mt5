package mt5

import (
	"fmt"

	"github.com/kahono0/mt5/models"
)

func (c *AppClient) SendOrder(req *models.TradeRequest) error {
	body, err := req.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize trade request: %w", err)
	}

	if err := c.core.SendCommand(CommandSubmitTrade, body); err != nil {
		return fmt.Errorf("failed to send trade request: %w", err)
	}

	return nil
}
