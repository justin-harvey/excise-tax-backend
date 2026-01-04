package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/maxfelker/excise-tax-backend/v2/internal/xrpl"
)

const (
	// Exit codes
	exitOK       = 0
	exitError    = 1
	exitUsage    = 2
	
	// XRPL network URLs
	testnetURL = "wss://s.altnet.rippletest.net:51233"
	mainnetURL = "wss://xrplcluster.com"
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

	// Create client
	client := xrpl.New(url)
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
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
	xrp, err := xrpl.DropsToXRP(info.Balance)
	if err != nil {
		return fmt.Errorf("failed to convert balance: %w", err)
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
	xrp, err := xrpl.DropsToXRP(info.Balance)
	if err != nil {
		return fmt.Errorf("failed to convert balance: %w", err)
	}

	// Output structured information to stdout
	fmt.Printf("Account:  %s\n", info.Account)
	fmt.Printf("Balance:  %s XRP (%s drops)\n", xrp, info.Balance)
	fmt.Printf("Sequence: %d\n", info.Sequence)
	fmt.Printf("Objects:  %d\n", info.OwnerCount)
	if info.PreviousTxn != "" {
		fmt.Printf("Last Tx:  %s\n", info.PreviousTxn)
	}
	
	return nil
}

func usage() {
	fmt.Fprintf(os.Stderr, `xrpl-cli - XRPL Command Line Tool

Usage:
  xrpl-cli [options] <command> [arguments]

Commands:
  balance <address>    Get XRP balance for an address
  info <address>       Get detailed account information
  help                 Show this help message

Options:
  -network string      XRPL network: testnet or mainnet (default "testnet")
  -timeout duration    Request timeout (default 10s)

Examples:
  # Get balance on testnet
  xrpl-cli balance rN7n7otQDd6FczFgLdlqtyMVrn3NnrcVc3

  # Get account info on mainnet
  xrpl-cli -network mainnet info rN7n7otQDd6FczFgLdlqtyMVrn3NnrcVc3

  # Use as Unix filter (extract just the number)
  xrpl-cli balance rN7n7otQDd6FczFgLdlqtyMVrn3NnrcVc3 | cut -d' ' -f1

Exit Codes:
  0  Success
  1  Error
  2  Usage error

`)
}
