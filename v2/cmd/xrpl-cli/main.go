package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/maxfelker/excise-tax-backend/v2/pkg/xrpl"
)

const (
	// Exit codes
	exitOK    = 0
	exitError = 1
	exitUsage = 2

	// XRPL network URLs
	testnetURL = "wss://s.altnet.rippletest.net:51233"
	mainnetURL = "wss://xrplcluster.com"
)

var (
	// Global flags
	jsonOutput bool
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitError)
	}
}

func run() error {
	// Define flags
	network := flag.String("network", "testnet", "XRPL network (testnet or mainnet)")
	timeout := flag.Duration("timeout", 10*time.Second, "Request timeout")
	flag.BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	flag.Usage = usage
	flag.Parse()

	// Get subcommand
	args := flag.Args()
	if len(args) < 1 {
		usage()
		os.Exit(exitUsage)
	}

	command := args[0]

	// Determine network URL
	var url string
	switch *network {
	case "testnet":
		url = testnetURL
	case "mainnet":
		url = mainnetURL
	default:
		return fmt.Errorf("invalid network: %s (must be 'testnet' or 'mainnet')", *network)
	}

	// Create client with optional logger
	var logger *slog.Logger
	if os.Getenv("DEBUG") != "" {
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}

	client := xrpl.New(url,
		xrpl.WithTimeout(*timeout),
		xrpl.WithLogger(logger),
	)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), *timeout+5*time.Second)
	defer cancel()

	// Connect to XRPL
	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	fmt.Fprintf(os.Stderr, "Connected to %s\n", *network)

	// Route to command handler
	switch command {
	case "balance":
		return handleBalance(ctx, client, args[1:])
	case "info":
		return handleInfo(ctx, client, args[1:])
	case "tx":
		return handleTransaction(ctx, client, args[1:])
	case "history":
		return handleHistory(ctx, client, args[1:])
	case "subscribe":
		return handleSubscribe(ctx, client, args[1:])
	case "help", "--help", "-h":
		usage()
		return nil
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func handleBalance(ctx context.Context, client *xrpl.Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: xrpl-cli balance <address>")
	}

	address := args[0]

	info, err := client.GetAccountInfo(ctx, address)
	if err != nil {
		return fmt.Errorf("failed to get account info: %w", err)
	}

	// Convert drops to XRP
	xrp, err := xrpl.DropsToXRP(fmt.Sprintf("%d", info.Balance))
	if err != nil {
		return fmt.Errorf("failed to convert balance: %w", err)
	}

	if jsonOutput {
		output := map[string]interface{}{
			"account":       info.Account,
			"balance":       xrp,
			"balance_drops": info.Balance,
		}
		return printJSON(output)
	}

	// Output to stdout (for piping)
	fmt.Printf("%s XRP\n", xrp)

	return nil
}

func handleInfo(ctx context.Context, client *xrpl.Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: xrpl-cli info <address>")
	}

	address := args[0]

	info, err := client.GetAccountInfo(ctx, address)
	if err != nil {
		return fmt.Errorf("failed to get account info: %w", err)
	}

	// Convert drops to XRP
	xrp, err := xrpl.DropsToXRP(fmt.Sprintf("%d", info.Balance))
	if err != nil {
		return fmt.Errorf("failed to convert balance: %w", err)
	}

	if jsonOutput {
		output := map[string]interface{}{
			"account":       info.Account,
			"balance":       xrp,
			"balance_drops": info.Balance,
			"sequence":      info.Sequence,
			"owner_count":   info.OwnerCount,
		}
		if info.PreviousTxnID != "" {
			output["previous_txn"] = info.PreviousTxnID
		}
		return printJSON(output)
	}

	// Output structured information to stdout
	fmt.Printf("Account:  %s\n", info.Account)
	fmt.Printf("Balance:  %s XRP (%d drops)\n", xrp, info.Balance)
	fmt.Printf("Sequence: %d\n", info.Sequence)
	fmt.Printf("Objects:  %d\n", info.OwnerCount)
	if info.PreviousTxnID != "" {
		fmt.Printf("Last Tx:  %s\n", info.PreviousTxnID)
	}

	return nil
}

func handleTransaction(ctx context.Context, client *xrpl.Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: xrpl-cli tx <hash>")
	}

	hash := args[0]

	result, err := client.GetTransaction(ctx, hash)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}

	if jsonOutput {
		output := map[string]interface{}{
			"hash":      result.Hash,
			"status":    result.Status,
			"validated": result.Validated,
			"type":      result.Tx.TransactionType,
			"account":   result.Tx.Account,
		}
		if result.Tx.Destination != "" {
			output["destination"] = result.Tx.Destination
		}
		if result.Tx.Amount != nil {
			output["amount"] = result.Tx.Amount
		}
		if result.Tx.Fee != "" {
			output["fee"] = result.Tx.Fee
		}
		if result.Tx.Date > 0 {
			unixTime := result.Tx.Date + 946684800
			output["date"] = time.Unix(unixTime, 0).Format(time.RFC3339)
			output["date_unix"] = unixTime
		}
		return printJSON(output)
	}

	// Output transaction details
	fmt.Printf("Hash:       %s\n", result.Hash)
	fmt.Printf("Status:     %s\n", result.Status)
	fmt.Printf("Validated:  %v\n", result.Validated)
	fmt.Printf("Type:       %s\n", result.Tx.TransactionType)
	fmt.Printf("From:       %s\n", result.Tx.Account)

	if result.Tx.Destination != "" {
		fmt.Printf("To:         %s\n", result.Tx.Destination)
	}

	if result.Tx.Amount != nil {
		// Amount can be string (XRP) or object (IOU)
		switch amount := result.Tx.Amount.(type) {
		case string:
			xrp, err := xrpl.DropsToXRP(amount)
			if err == nil {
				fmt.Printf("Amount:     %s XRP (%s drops)\n", xrp, amount)
			}
		default:
			fmt.Printf("Amount:     %v\n", amount)
		}
	}

	if result.Tx.Fee != "" {
		feeXRP, err := xrpl.DropsToXRP(result.Tx.Fee)
		if err == nil {
			fmt.Printf("Fee:        %s XRP\n", feeXRP)
		}
	}

	if result.Tx.Date > 0 {
		// Convert Ripple epoch (2000-01-01) to Unix epoch
		unixTime := result.Tx.Date + 946684800
		t := time.Unix(unixTime, 0)
		fmt.Printf("Date:       %s\n", t.Format(time.RFC3339))
	}

	return nil
}

func handleHistory(ctx context.Context, client *xrpl.Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: xrpl-cli history <address> [limit]")
	}

	address := args[0]
	limit := 10 // default

	if len(args) > 1 {
		_, err := fmt.Sscanf(args[1], "%d", &limit)
		if err != nil {
			return fmt.Errorf("invalid limit: %s", args[1])
		}
	}

	transactions, err := client.GetAccountTransactions(ctx, address, limit)
	if err != nil {
		return fmt.Errorf("failed to get transaction history: %w", err)
	}

	if jsonOutput {
		output := map[string]interface{}{
			"address":      address,
			"count":        len(transactions),
			"transactions": transactions,
		}
		return printJSON(output)
	}

	if len(transactions) == 0 {
		fmt.Println("No transactions found")
		return nil
	}

	fmt.Printf("Found %d transaction(s):\n\n", len(transactions))

	for i, tx := range transactions {
		fmt.Printf("%d. %s\n", i+1, tx.Hash)
		fmt.Printf("   Type:      %s\n", tx.TransactionType)
		fmt.Printf("   From:      %s\n", tx.Account)

		if tx.Destination != "" {
			fmt.Printf("   To:        %s\n", tx.Destination)
		}

		if tx.Amount != nil {
			switch amount := tx.Amount.(type) {
			case string:
				xrp, err := xrpl.DropsToXRP(amount)
				if err == nil {
					fmt.Printf("   Amount:    %s XRP\n", xrp)
				}
			default:
				fmt.Printf("   Amount:    %v\n", amount)
			}
		}

		if tx.Date > 0 {
			unixTime := tx.Date + 946684800
			t := time.Unix(unixTime, 0)
			fmt.Printf("   Date:      %s\n", t.Format("2006-01-02 15:04:05"))
		}

		fmt.Printf("   Validated: %v\n", tx.Validated)
		fmt.Println()
	}

	return nil
}

func handleSubscribe(ctx context.Context, client *xrpl.Client, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: xrpl-cli subscribe <address>")
	}

	address := args[0]

	fmt.Fprintf(os.Stderr, "Subscribing to transactions for %s...\n", address)

	streamChan, err := client.SubscribeToAccount(ctx, address)
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Subscribed! Listening for transactions (Press Ctrl+C to stop)...\n\n")

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Listen for transactions
	for {
		select {
		case msg, ok := <-streamChan:
			if !ok {
				return fmt.Errorf("subscription channel closed")
			}

			if jsonOutput {
				if err := printJSON(msg); err != nil {
					fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
				}
			} else {
				printStreamMessage(msg)
			}

		case <-sigChan:
			fmt.Fprintf(os.Stderr, "\nUnsubscribing...\n")
			if err := client.Unsubscribe(ctx, address); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to unsubscribe: %v\n", err)
			}
			return nil

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func printStreamMessage(msg xrpl.StreamMessage) {
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("New Transaction: %s\n", msg.Type)

	if msg.Transaction != nil {
		tx := msg.Transaction
		fmt.Printf("  Hash:      %s\n", tx.Hash)
		fmt.Printf("  Type:      %s\n", tx.TransactionType)
		fmt.Printf("  From:      %s\n", tx.Account)

		if tx.Destination != "" {
			fmt.Printf("  To:        %s\n", tx.Destination)
		}

		if tx.Amount != nil {
			switch amount := tx.Amount.(type) {
			case string:
				xrp, err := xrpl.DropsToXRP(amount)
				if err == nil {
					fmt.Printf("  Amount:    %s XRP\n", xrp)
				}
			default:
				fmt.Printf("  Amount:    %v\n", amount)
			}
		}

		if tx.Fee != "" {
			feeXRP, err := xrpl.DropsToXRP(tx.Fee)
			if err == nil {
				fmt.Printf("  Fee:       %s XRP\n", feeXRP)
			}
		}

		if tx.Date > 0 {
			unixTime := tx.Date + 946684800
			t := time.Unix(unixTime, 0)
			fmt.Printf("  Date:      %s\n", t.Format(time.RFC3339))
		}
	}

	fmt.Printf("  Validated: %v\n", msg.Validated)

	if msg.Status != "" {
		fmt.Printf("  Status:    %s\n", msg.Status)
	}

	if msg.EngineResult != "" {
		fmt.Printf("  Result:    %s\n", msg.EngineResult)
		if msg.EngineResultMessage != "" {
			fmt.Printf("  Message:   %s\n", msg.EngineResultMessage)
		}
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
}

// printJSON outputs data as formatted JSON to stdout
func printJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func usage() {
	fmt.Fprintf(os.Stderr, `xrpl-cli - XRPL Command Line Tool

Usage:
  xrpl-cli [options] <command> [arguments]

Commands:
  balance <address>         Get XRP balance for an address
  info <address>            Get detailed account information
  tx <hash>                 Get transaction details by hash
  history <address> [limit] Get transaction history (default limit: 10)
  subscribe <address>       Subscribe to real-time transactions for an account
  help                      Show this help message

Options:
  -network string      XRPL network: testnet or mainnet (default "testnet")
  -timeout duration    Request timeout (default 10s)
  -json                Output in JSON format for machine parsing

Examples:
  # Get balance on testnet
  xrpl-cli balance rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY

  # Get account info on mainnet in JSON format
  xrpl-cli -network mainnet -json info rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY

  # Get transaction details
  xrpl-cli tx C71F385124008A436842B56DEF8196B0621762FBD9464F2510EE9C3D1A3322DA

  # Get last 20 transactions in JSON
  xrpl-cli -json history rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY 20

  # Subscribe to real-time transactions
  xrpl-cli subscribe rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY

  # Subscribe with JSON output for processing
  xrpl-cli -json subscribe rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY

  # Use as Unix filter (extract just the number)
  xrpl-cli balance rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY | cut -d' ' -f1

  # Parse JSON with jq
  xrpl-cli -json balance rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY | jq .balance

Exit Codes:
  0  Success
  1  Error
  2  Usage error

`)
}
