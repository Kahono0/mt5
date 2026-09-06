// Package mt5 provides a client library for connecting to MetaTrader 5
// servers over WebSocket. It handles the MT5 binary protocol, including
// encryption, command framing, authentication, heartbeats, and reconnection.
//
// The package offers two levels of abstraction:
//
//   - [Client] is a low-level WebSocket client that manages the connection
//     lifecycle, encryption, and message dispatch.
//
//   - [AppClient] is a high-level wrapper that adds automatic session-key
//     negotiation, login, and convenient action methods.
//
// # Quick Start
//
// Use [NewApp] to create an [AppClient] that handles the full connection
// and authentication flow:
//
//	client, err := mt5.NewApp(mt5.AppConfig{
//	    URL:       "wss://broker.example/ws",
//	    ServerURL: "broker.example",
//	    AccountID: 123456,
//	    Password:  "secret",
//	    DeviceID:  "device-id",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	ctx := context.Background()
//	go client.Start(ctx)
//	client.WaitUntilAuthenticated(ctx)
package mt5
