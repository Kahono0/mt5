package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/kahono0/mt5"
	"github.com/kahono0/mt5/models"
)

var masterConfig = mt5.AppConfig{
	URL:                  "wss://web.metatrader.app/terminal",
	ServerURL:            "web.metatrader.app",
	AccountID:            5055676794,
	Password:             "@cGdZe1a",
	DeviceID:             "Linux x86_64;1;en-US;1920x1080;374418",
	PingInterval:         5 * time.Second,
	ReconnectMaxAttempts: -1,
	OnError: func(err error) {
		log.Printf("client error: %v", err)
	},
	OnDisconnect: func(err error) {
		log.Printf("client disconnected: %v", err)
	},
}

var slaveConfig = mt5.AppConfig{
	URL:                  "wss://web.metatrader.app/terminal",
	ServerURL:            "web.metatrader.app",
	AccountID:            5059096177,
	Password:             "GrYk-c8v",
	DeviceID:             "Linux x86_64;1;en-US;1920x1080;376718",
	PingInterval:         5 * time.Second,
	ReconnectMaxAttempts: -1,
	OnError: func(err error) {
		log.Printf("client error: %v", err)
	},
	OnDisconnect: func(err error) {
		log.Printf("client disconnected: %v", err)
	},
}

func main() {
	client, err := mt5.NewApp(masterConfig)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := client.Start(context.Background()); err != nil {
			log.Printf("client stopped: %v", err)
		}
	}()

	client.RegisterTransactionEventHandler(func(event *models.EnvelopeTransaction) {
		fmt.Println("Transaction Event: ", asPrettyJSON(event))
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.WaitUntilAuthenticated(ctx); err != nil {
		log.Fatal(err)
	}

	log.Printf("Client authenticated successfully")

	accountInfo, err := client.GetAccountInfo(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Account info: %s", accountInfo.Core.AccountName)

	openPositions, err := client.GetOpenPositions(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Open positions: %v", openPositions)
	fmt.Println("Open positions: ", asPrettyJSON(openPositions))

	//now := time.Now()
	//from := now.Add(-72 * time.Hour)
	//to := now

	//tradeHistory, err := client.GetTradeHistory(context.Background(), from, to)
	//if err != nil {
	//	log.Fatal(err)
	//}

	//log.Printf("Trade history: DealsCount=%d, OrdersCount=%d", len(tradeHistory.Deals), len(tradeHistory.Orders))

	select {}
}

func asPrettyJSON(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("failed to marshal to JSON: %v", err)
	}
	return string(b)
}
