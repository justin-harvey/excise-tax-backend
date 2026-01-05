package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"

	"github.com/maxfelker/excise-tax-backend/v2/internal/payment"
	"github.com/maxfelker/excise-tax-backend/v2/pkg/logger"
	"github.com/maxfelker/excise-tax-backend/v2/pkg/xrpl"
)

const (
	exitSuccess = 0
	exitError   = 1
	exitUsage   = 2
)

var (
	// Global flags
	jsonOutput = flag.Bool("json", false, "output in JSON format")
	verbose    = flag.Bool("v", false, "verbose logging")
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(exitError)
	}
}

func run() error {
	if len(os.Args) < 2 {
		printUsage()
		return fmt.Errorf("command required")
	}

	command := os.Args[1]
	args := os.Args[2:]

	// Initialize logger
	logOpts := logger.Options{
		Level: logger.LevelInfo,
		JSON:  false,
	}
	if *verbose {
		logOpts.Level = logger.LevelDebug
	}
	log := logger.New(logOpts)

	// Initialize XRPL client (for XRPL payments)
	xrplClient := xrpl.New("wss://s.altnet.rippletest.net:51233",
		xrpl.WithLogger(log),
	)
	defer xrplClient.Close()

	ctx := context.Background()
	if err := xrplClient.Connect(ctx); err != nil {
		log.Warn("XRPL client connection failed, continuing without network access", "error", err)
	}

	// Create payment processor
	processor := payment.NewProcessor(xrplClient)

	// Route to command handler
	switch command {
	case "create":
		return handleCreate(processor, args)
	case "get":
		return handleGet(processor, args)
	case "list":
		return handleList(processor, args)
	case "verify":
		return handleVerify(processor, args)
	case "retry":
		return handleRetry(processor, args)
	case "events":
		return handleEvents(processor, args)
	case "help", "-h", "--help":
		printUsage()
		return nil
	default:
		printUsage()
		return fmt.Errorf("unknown command: %s", command)
	}
}

func handleCreate(processor *payment.Processor, args []string) error {
	fs := flag.NewFlagSet("create", flag.ExitOnError)
	paymentType := fs.String("type", "xrpl", "payment type (xrpl, credit_card, ach)")
	amount := fs.String("amount", "", "payment amount")
	currency := fs.String("currency", "XRP", "payment currency")
	from := fs.String("from", "", "sender address")
	to := fs.String("to", "", "receiver address")
	fs.Parse(args)

	ctx := context.Background()
	var req payment.CreatePaymentRequest

	// If positional argument provided, read from file or stdin
	if fs.NArg() > 0 {
		source := fs.Arg(0)
		var data []byte
		var err error

		if source == "-" {
			data, err = io.ReadAll(os.Stdin)
		} else {
			data, err = os.ReadFile(source)
		}

		if err != nil {
			return fmt.Errorf("failed to read payment data: %w", err)
		}

		if err := json.Unmarshal(data, &req); err != nil {
			return fmt.Errorf("failed to parse payment JSON: %w", err)
		}
	} else {
		// Use flags to create request
		if *amount == "" {
			return fmt.Errorf("--amount is required")
		}
		if *from == "" {
			return fmt.Errorf("--from is required")
		}
		if *to == "" {
			return fmt.Errorf("--to is required")
		}

		req = payment.CreatePaymentRequest{
			Type:     payment.PaymentType(*paymentType),
			Amount:   *amount,
			Currency: *currency,
			From:     *from,
			To:       *to,
		}
	}

	// Create payment
	tx, err := processor.CreatePayment(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}

	// Output result
	if *jsonOutput {
		return outputJSON(tx)
	}

	fmt.Printf("Payment created: %s\n", tx.ID)
	fmt.Printf("Type: %s\n", tx.Type)
	fmt.Printf("Amount: %s %s\n", tx.Amount, tx.Currency)
	fmt.Printf("From: %s\n", tx.From)
	fmt.Printf("To: %s\n", tx.To)
	fmt.Printf("State: %s\n", tx.State)
	fmt.Printf("Created: %s\n", tx.CreatedAt.Format(time.RFC3339))

	return nil
}

func handleGet(processor *payment.Processor, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("payment ID required")
	}

	paymentID := args[0]
	ctx := context.Background()

	tx, err := processor.GetPayment(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	// Output result
	if *jsonOutput {
		return outputJSON(tx)
	}

	fmt.Printf("Payment ID: %s\n", tx.ID)
	fmt.Printf("Type: %s\n", tx.Type)
	fmt.Printf("Amount: %s %s\n", tx.Amount, tx.Currency)
	fmt.Printf("From: %s\n", tx.From)
	fmt.Printf("To: %s\n", tx.To)
	fmt.Printf("State: %s\n", tx.State)
	fmt.Printf("Events: %d\n", len(tx.Events))
	fmt.Printf("Created: %s\n", tx.CreatedAt.Format(time.RFC3339))
	fmt.Printf("Updated: %s\n", tx.UpdatedAt.Format(time.RFC3339))

	if len(tx.Metadata) > 0 {
		fmt.Printf("Metadata: %s\n", string(tx.Metadata))
	}

	return nil
}

func handleList(processor *payment.Processor, args []string) error {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	stateStr := fs.String("state", "", "filter by state (pending, completed, in_review, failed)")
	typeStr := fs.String("type", "", "filter by type (xrpl, credit_card, ach)")
	limit := fs.Int("limit", 10, "maximum number of results")
	offset := fs.Int("offset", 0, "offset for pagination")
	fs.Parse(args)

	ctx := context.Background()
	filters := payment.ListPaymentsFilters{
		Limit:  *limit,
		Offset: *offset,
	}

	if *stateStr != "" {
		state := payment.TransactionState(*stateStr)
		filters.State = &state
	}

	if *typeStr != "" {
		paymentType := payment.PaymentType(*typeStr)
		filters.Type = &paymentType
	}

	payments, err := processor.ListPayments(ctx, filters)
	if err != nil {
		return fmt.Errorf("failed to list payments: %w", err)
	}

	// Output result
	if *jsonOutput {
		return outputJSON(payments)
	}

	if len(payments) == 0 {
		fmt.Println("No payments found")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tType\tAmount\tState\tCreated")
	for _, tx := range payments {
		fmt.Fprintf(w, "%s\t%s\t%s %s\t%s\t%s\n",
			tx.ID, tx.Type, tx.Amount, tx.Currency, tx.State,
			tx.CreatedAt.Format("2006-01-02 15:04"))
	}
	w.Flush()

	return nil
}

func handleVerify(processor *payment.Processor, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("payment ID required")
	}

	paymentID := args[0]
	ctx := context.Background()

	tx, err := processor.VerifyPayment(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("failed to verify payment: %w", err)
	}

	// Output result
	if *jsonOutput {
		return outputJSON(tx)
	}

	fmt.Printf("Payment ID: %s\n", tx.ID)
	fmt.Printf("State: %s\n", tx.State)
	fmt.Printf("Events: %d\n", len(tx.Events))
	fmt.Printf("Last Updated: %s\n", tx.UpdatedAt.Format(time.RFC3339))

	return nil
}

func handleRetry(processor *payment.Processor, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("payment ID required")
	}

	paymentID := args[0]
	ctx := context.Background()

	if err := processor.RetryPayment(ctx, paymentID); err != nil {
		return fmt.Errorf("failed to retry payment: %w", err)
	}

	fmt.Printf("Payment retry initiated for %s\n", paymentID)
	return nil
}

func handleEvents(processor *payment.Processor, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("payment ID required")
	}

	paymentID := args[0]
	ctx := context.Background()

	tx, err := processor.GetPayment(ctx, paymentID)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}

	// Output result
	if *jsonOutput {
		return outputJSON(tx.Events)
	}

	if len(tx.Events) == 0 {
		fmt.Println("No events found")
		return nil
	}

	fmt.Printf("Events for payment %s:\n\n", paymentID)
	for i, event := range tx.Events {
		fmt.Printf("Event %d: %s\n", i+1, event.Type)
		fmt.Printf("  State: %s\n", event.State)
		fmt.Printf("  Time: %s\n", event.Timestamp.Format(time.RFC3339))
		if len(event.Details) > 0 {
			fmt.Printf("  Details: %s\n", string(event.Details))
		}
		fmt.Println()
	}

	return nil
}

func outputJSON(v interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(v)
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `payment-cli - Payment processing command-line tool

Usage:
  payment-cli <command> [options]

Commands:
  create [file|-]     Create a new payment from JSON file or stdin
    --type string     Payment type: xrpl, credit_card, ach (default: xrpl)
    --amount string   Payment amount (required if not using file)
    --currency string Currency code (default: XRP)
    --from string     Sender address (required if not using file)
    --to string       Receiver address (required if not using file)

  get <payment-id>    Get payment details

  list                List payments with optional filters
    --state string    Filter by state: pending, completed, in_review, failed
    --type string     Filter by type: xrpl, credit_card, ach
    --limit int       Maximum results (default: 10)
    --offset int      Pagination offset (default: 0)

  verify <payment-id> Verify payment status

  retry <payment-id>  Retry a failed payment

  events <payment-id> Show payment event history

  help                Show this help message

Global Options:
  --json              Output in JSON format
  -v                  Verbose logging

Examples:
  # Create payment from flags
  payment-cli create --type xrpl --amount 100 \
    --from rSender... --to rReceiver...

  # Create payment from JSON file
  payment-cli create payment.json

  # Create payment from stdin
  cat payment.json | payment-cli create -

  # Get payment details
  payment-cli get pay_abc123

  # List pending XRPL payments
  payment-cli list --state pending --type xrpl

  # Verify payment and get latest status
  payment-cli verify pay_abc123

  # Retry failed payment
  payment-cli retry pay_abc123

  # Show event history
  payment-cli events pay_abc123

  # Get JSON output for scripting
  payment-cli get pay_abc123 --json
`)
}
