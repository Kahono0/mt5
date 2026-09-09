package mt5_test

import (
	"context"
	"log"
	"time"

	"github.com/kahono0/mt5"
)

func ExampleNewApp() {
	client, err := mt5.NewApp(mt5.AppConfig{
		URL:                  "wss://broker.example/ws",
		ServerURL:            "broker.example",
		AccountID:            123456,
		Password:             "secret",
		DeviceID:             "device-id",
		PingInterval:         5 * time.Second,
		ReconnectMaxAttempts: -1,
		OnError: func(err error) {
			switch err := err.(type) {
			case *mt5.ServerError:
				log.Printf("server error: cmd=%d code=%d msg=%s", err.Command, err.Code, err.Msg)
			case *mt5.CommandMismatchError:
				log.Printf("command mismatch: expected=%d got=%d", err.Expected, err.Got)
			default:
				log.Printf("error: %v", err)
			}
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	client.RegisterTickHandler(func(ticks []mt5.Tick, msg mt5.Message) {
		for _, tick := range ticks {
			log.Printf("tick: symbol=%d bid=%.5f ask=%.5f", tick.SymbolID, tick.Bid, tick.Ask)
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go func() {
		if err := client.Start(ctx); err != nil {
			log.Printf("client stopped: %v", err)
		}
	}()

	_ = client.WaitUntilAuthenticated(ctx)
	_ = client.SubscribeTicks([]uint32{1001, 1002})
	_ = client.RequestAccountInfo()
	_ = client.Close()
}
