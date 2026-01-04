package xrpl_test

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/maxfelker/excise-tax-backend/v2/pkg/xrpl"
)

// ExampleNew demonstrates creating a basic XRPL client.
func ExampleNew() {
	client := xrpl.New("wss://s.altnet.rippletest.net:51233")
	fmt.Printf("Client created: %v\n", client != nil)
	// Output: Client created: true
}

// ExampleNew_withOptions demonstrates creating a client with custom options.
func ExampleNew_withOptions() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	client := xrpl.New("wss://s.altnet.rippletest.net:51233",
		xrpl.WithTimeout(15*time.Second),
		xrpl.WithLogger(logger),
		xrpl.WithMaxRetries(5),
	)

	fmt.Printf("Client created with options: %v\n", client != nil)
	// Output: Client created with options: true
}

// ExampleClient_GetAccountInfo demonstrates querying account information.
func ExampleClient_GetAccountInfo() {
	client := xrpl.New("wss://s.altnet.rippletest.net:51233")
	ctx := context.Background()

	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	info, err := client.GetAccountInfo(ctx, "rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY")
	if err != nil {
		log.Fatal(err)
	}

	// Convert drops to XRP
	xrp, _ := xrpl.DropsToXRP(info.Balance)
	fmt.Printf("Account has %s XRP\n", xrp)
}

// ExampleClient_GetAccountTransactions demonstrates querying transaction history.
func ExampleClient_GetAccountTransactions() {
	client := xrpl.New("wss://s.altnet.rippletest.net:51233")
	ctx := context.Background()

	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Get last 5 transactions
	txs, err := client.GetAccountTransactions(ctx, "rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY", 5)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d transactions\n", len(txs))
	for _, tx := range txs {
		fmt.Printf("- %s (%s)\n", tx.TransactionType, tx.Hash[:8])
	}
}

// ExampleClient_SubscribeToAccount demonstrates subscribing to real-time transactions.
func ExampleClient_SubscribeToAccount() {
	client := xrpl.New("wss://s.altnet.rippletest.net:51233")
	ctx := context.Background()

	if err := client.Connect(ctx); err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	address := "rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY"

	stream, err := client.SubscribeToAccount(ctx, address)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Unsubscribe(ctx, address)

	// Listen for 10 seconds
	timeout := time.After(10 * time.Second)

	for {
		select {
		case msg := <-stream:
			if msg.Transaction != nil {
				fmt.Printf("New transaction: %s\n", msg.Transaction.Hash)
			}
		case <-timeout:
			fmt.Println("Monitoring complete")
			return
		}
	}
}

// ExampleDropsToXRP demonstrates converting drops to XRP.
func ExampleDropsToXRP() {
	drops := "1000000"
	xrp, err := xrpl.DropsToXRP(drops)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s drops = %s XRP\n", drops, xrp)
	// Output: 1000000 drops = 1.000000 XRP
}

// ExampleXRPToDrops demonstrates converting XRP to drops.
func ExampleXRPToDrops() {
	xrp := 1.5
	drops, err := xrpl.XRPToDrops(xrp)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%f XRP = %s drops\n", xrp, drops)
	// Output: 1.500000 XRP = 1500000 drops
}

// ExampleIsValidAddress demonstrates address validation.
func ExampleIsValidAddress() {
	valid := "rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY"
	invalid := "invalid_address"

	fmt.Printf("%s is valid: %v\n", valid, xrpl.IsValidAddress(valid))
	fmt.Printf("%s is valid: %v\n", invalid, xrpl.IsValidAddress(invalid))
	// Output:
	// rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY is valid: true
	// invalid_address is valid: false
}
