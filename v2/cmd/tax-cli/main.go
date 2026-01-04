// Tax CLI - Command-line tool for excise tax calculations
//
// This tool follows Unix philosophy:
// - Small, focused tool that does one thing well
// - Accepts input from files or stdin
// - Produces output to stdout
// - Proper exit codes for scripting
//
// Usage:
//
//	tax-cli calculate <file>     Calculate tax from production data JSON file
//	tax-cli rates                List all available tax rates
//	tax-cli rate <args>          Get specific tax rate
//	tax-cli validate <file>      Validate production data JSON file
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/maxfelker/excise-tax-backend/v2/internal/tax"
	"github.com/maxfelker/excise-tax-backend/v2/pkg/logger"
)

const (
	exitSuccess         = 0
	exitError           = 1
	exitUsageError      = 2
	exitValidationError = 3
)

var (
	// Global flags
	jsonOutput    = flag.Bool("json", false, "Output in JSON format")
	ratesFile     = flag.String("rates", "", "Path to tax rates JSON file (uses federal defaults if not specified)")
	jurisdiction  = flag.String("jurisdiction", "federal", "Tax jurisdiction")
	effectiveDate = flag.String("date", "", "Effective date for tax calculation (default: today, format: 2006-01-02)")
	verbose       = flag.Bool("v", false, "Verbose output (debug logging)")
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(exitUsageError)
	}

	command := os.Args[1]

	// Special case for help command - don't parse flags
	if command == "help" || command == "-h" || command == "--help" {
		printUsage()
		os.Exit(exitSuccess)
	}

	// Parse flags after the command
	// Set up a new FlagSet to parse after command
	fs := flag.NewFlagSet(command, flag.ContinueOnError) // Changed from ExitOnError to ContinueOnError
	jsonOutput := fs.Bool("json", false, "Output in JSON format")
	ratesFile := fs.String("rates", "", "Path to tax rates JSON file")
	jurisdiction := fs.String("jurisdiction", "federal", "Tax jurisdiction")
	effectiveDate := fs.String("date", "", "Effective date (YYYY-MM-DD)")
	verbose := fs.Bool("v", false, "Verbose output")

	// Parse remaining args
	parseErr := fs.Parse(os.Args[2:])
	if parseErr != nil {
		fmt.Fprintf(os.Stderr, "Flag parse error: %v\n", parseErr)
		os.Exit(exitUsageError)
	}

	// Initialize logger
	logLevel := logger.LevelInfo
	if *verbose {
		logLevel = logger.LevelDebug
	}
	log := logger.New(logger.Options{
		Level:  logLevel,
		JSON:   false,
		Writer: os.Stderr,
	})

	ctx := context.Background()

	// Load tax rates
	var rates []tax.TaxRate
	var err error

	if *ratesFile != "" {
		log.Debug("loading rates from file", "file", *ratesFile)
		rates, err = tax.LoadRatesFromJSON(*ratesFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to load rates: %v\n", err)
			os.Exit(exitError)
		}
	} else {
		log.Debug("using default federal rates")
		rates = tax.DefaultFederalRates()
	}

	// Create calculator
	calc := tax.NewCalculator(rates, tax.WithLogger(log))

	// Parse effective date
	var date time.Time
	if *effectiveDate != "" {
		date, err = time.Parse("2006-01-02", *effectiveDate)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid date format (use YYYY-MM-DD): %v\n", err)
			os.Exit(exitUsageError)
		}
	} else {
		date = time.Now()
	}

	// Execute subcommand
	switch command {
	case "calculate", "calc":
		exitCode := calculateCommand(ctx, calc, date, fs.Args(), *jsonOutput, *jurisdiction, log)
		os.Exit(exitCode)
	case "rates":
		exitCode := ratesCommand(calc, *jsonOutput, log)
		os.Exit(exitCode)
	case "rate":
		exitCode := rateCommand(calc, date, fs.Args(), *jsonOutput, *jurisdiction, log)
		os.Exit(exitCode)
	case "validate":
		exitCode := validateCommand(fs.Args(), *jsonOutput, log)
		os.Exit(exitCode)
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command: %s\n\n", command)
		printUsage()
		os.Exit(exitUsageError)
	}
}

func calculateCommand(ctx context.Context, calc *tax.Calculator, date time.Time, args []string, jsonOutput bool, jurisdiction string, log *slog.Logger) int {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: calculate command requires a file argument or '-' for stdin\n")
		fmt.Fprintf(os.Stderr, "Usage: tax-cli calculate <file>\n")
		return exitUsageError
	}

	filename := args[0]

	// Read production data
	var reader io.Reader
	if filename == "-" {
		reader = os.Stdin
	} else {
		file, err := os.Open(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open file: %v\n", err)
			return exitError
		}
		defer file.Close()
		reader = file
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to read input: %v\n", err)
		return exitError
	}

	// Parse production data
	var productionData tax.ProductionData
	if err := json.Unmarshal(data, &productionData); err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid JSON: %v\n", err)
		return exitError
	}

	log.Debug("calculating tax",
		"items", len(productionData.Items),
		"jurisdiction", jurisdiction,
		"date", date.Format("2006-01-02"))

	// Calculate tax
	result, err := calc.Calculate(ctx, &productionData, jurisdiction, date)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: calculation failed: %v\n", err)
		return exitError
	}

	// Output results
	if jsonOutput {
		output, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(output))
	} else {
		printCalculationResult(result)
	}

	return exitSuccess
}

func ratesCommand(calc *tax.Calculator, jsonOutput bool, log *slog.Logger) int {
	rates := calc.GetAllRates()

	if jsonOutput {
		output, _ := json.MarshalIndent(rates, "", "  ")
		fmt.Println(string(output))
	} else {
		fmt.Printf("Available Tax Rates (%d total):\n\n", len(rates))
		for _, rate := range rates {
			fmt.Printf("Product: %-10s  Rate: $%.2f per %s  Jurisdiction: %s\n",
				rate.ProductType, rate.RatePerUnit, rate.UnitType, rate.Jurisdiction)
			fmt.Printf("  Effective: %s", rate.EffectiveDate.Format("2006-01-02"))
			if rate.ExpirationDate != nil {
				fmt.Printf("  Expires: %s", rate.ExpirationDate.Format("2006-01-02"))
			}
			fmt.Println()
			if rate.Description != "" {
				fmt.Printf("  %s\n", rate.Description)
			}
			fmt.Println()
		}
	}

	return exitSuccess
}

func rateCommand(calc *tax.Calculator, date time.Time, args []string, jsonOutput bool, jurisdiction string, log *slog.Logger) int {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Error: rate command requires product-type and unit-type\n")
		fmt.Fprintf(os.Stderr, "Usage: tax-cli rate <product-type> <unit-type>\n")
		fmt.Fprintf(os.Stderr, "Example: tax-cli rate beer barrel\n")
		return exitUsageError
	}

	productType := tax.ProductType(args[0])
	unitType := tax.UnitType(args[1])

	rate, err := calc.GetRate(productType, unitType, jurisdiction, date)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return exitError
	}

	if jsonOutput {
		output, _ := json.MarshalIndent(rate, "", "  ")
		fmt.Println(string(output))
	} else {
		fmt.Printf("Tax Rate for %s (%s)\n\n", productType, unitType)
		fmt.Printf("Rate:         $%.2f per %s\n", rate.RatePerUnit, rate.UnitType)
		fmt.Printf("Jurisdiction: %s\n", rate.Jurisdiction)
		fmt.Printf("Effective:    %s\n", rate.EffectiveDate.Format("2006-01-02"))
		if rate.ExpirationDate != nil {
			fmt.Printf("Expires:      %s\n", rate.ExpirationDate.Format("2006-01-02"))
		}
		if rate.Description != "" {
			fmt.Printf("\n%s\n", rate.Description)
		}
	}

	return exitSuccess
}

func validateCommand(args []string, jsonOutput bool, log *slog.Logger) int {
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Error: validate command requires a file argument or '-' for stdin\n")
		fmt.Fprintf(os.Stderr, "Usage: tax-cli validate <file>\n")
		return exitUsageError
	}

	filename := args[0]

	// Read production data
	var reader io.Reader
	if filename == "-" {
		reader = os.Stdin
	} else {
		file, err := os.Open(filename)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to open file: %v\n", err)
			return exitError
		}
		defer file.Close()
		reader = file
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to read input: %v\n", err)
		return exitError
	}

	// Parse production data
	var productionData tax.ProductionData
	if err := json.Unmarshal(data, &productionData); err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid JSON: %v\n", err)
		return exitError
	}

	// Validate data
	errors := validateProductionData(&productionData)

	if len(errors) == 0 {
		if jsonOutput {
			fmt.Println(`{"valid":true}`)
		} else {
			fmt.Println("✓ Production data is valid")
		}
		return exitSuccess
	}

	// Has errors
	if jsonOutput {
		result := map[string]interface{}{
			"valid":  false,
			"errors": errors,
		}
		output, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(output))
	} else {
		fmt.Fprintf(os.Stderr, "✗ Validation failed with %d error(s):\n\n", len(errors))
		for _, err := range errors {
			fmt.Fprintf(os.Stderr, "  - %s: %s\n", err.Field, err.Message)
		}
	}

	return exitValidationError
}

func validateProductionData(data *tax.ProductionData) []tax.ValidationError {
	var errors []tax.ValidationError

	if len(data.Items) == 0 {
		errors = append(errors, tax.ValidationError{
			Field:   "items",
			Message: "at least one production item is required",
		})
		return errors
	}

	for i, item := range data.Items {
		prefix := fmt.Sprintf("items[%d]", i)

		if item.ProductName == "" {
			errors = append(errors, tax.ValidationError{
				Field:   prefix + ".product_name",
				Message: "product name is required",
			})
		}

		if item.ProductType == "" {
			errors = append(errors, tax.ValidationError{
				Field:   prefix + ".product_type",
				Message: "product type is required",
			})
		} else if !isValidProductType(item.ProductType) {
			errors = append(errors, tax.ValidationError{
				Field:   prefix + ".product_type",
				Message: fmt.Sprintf("invalid product type: %s", item.ProductType),
			})
		}

		if item.UnitType == "" {
			errors = append(errors, tax.ValidationError{
				Field:   prefix + ".unit_type",
				Message: "unit type is required",
			})
		} else if !isValidUnitType(item.UnitType) {
			errors = append(errors, tax.ValidationError{
				Field:   prefix + ".unit_type",
				Message: fmt.Sprintf("invalid unit type: %s", item.UnitType),
			})
		}

		if item.Quantity <= 0 {
			errors = append(errors, tax.ValidationError{
				Field:   prefix + ".quantity",
				Message: "quantity must be greater than zero",
			})
		}

		if item.ABV < 0 || item.ABV > 100 {
			errors = append(errors, tax.ValidationError{
				Field:   prefix + ".abv",
				Message: "ABV must be between 0 and 100",
			})
		}
	}

	return errors
}

func isValidProductType(pt tax.ProductType) bool {
	switch pt {
	case tax.ProductTypeBeer, tax.ProductTypeWine, tax.ProductTypeSpirits, tax.ProductTypeOther:
		return true
	default:
		return false
	}
}

func isValidUnitType(ut tax.UnitType) bool {
	switch ut {
	case tax.UnitTypeGallon, tax.UnitTypeBarrel, tax.UnitTypeCase, tax.UnitTypeLiter:
		return true
	default:
		return false
	}
}

func printCalculationResult(result *tax.TaxCalculation) {
	fmt.Printf("Tax Calculation Results\n")
	fmt.Printf("=======================\n\n")
	fmt.Printf("Jurisdiction:     %s\n", result.Jurisdiction)
	fmt.Printf("Calculated:       %s\n\n", result.CalculatedAt.Format("2006-01-02 15:04:05"))

	if len(result.BreakdownByProduct) > 0 {
		fmt.Printf("Breakdown by Product:\n")
		fmt.Printf("---------------------\n")
		for _, item := range result.BreakdownByProduct {
			fmt.Printf("Product:  %s (%s)\n", item.ProductName, item.ProductType)
			fmt.Printf("Quantity: %.2f %s @ $%.2f per %s\n",
				item.Quantity, item.UnitType, item.RatePerUnit, item.UnitType)
			fmt.Printf("Tax:      $%.2f\n\n", item.TaxAmount)
		}
	}

	fmt.Printf("---------------------\n")
	fmt.Printf("Total Tax:        $%.2f\n", result.TotalTaxAmount)
}

func printUsage() {
	fmt.Fprintf(os.Stderr, `tax-cli - Excise Tax Calculator

USAGE:
    tax-cli <command> [options] [arguments]

COMMANDS:
    calculate <file>    Calculate excise tax from production data JSON file
                        Use '-' to read from stdin
    
    rates              List all available tax rates
    
    rate <type> <unit> Get specific tax rate
                        Example: tax-cli rate beer barrel
    
    validate <file>    Validate production data JSON file
                        Use '-' to read from stdin
    
    help               Show this help message

GLOBAL OPTIONS:
    --json              Output in JSON format (for scripting)
    --rates <file>      Path to custom tax rates JSON file
    --jurisdiction <j>  Tax jurisdiction (default: federal)
    --date <date>       Effective date for calculation (YYYY-MM-DD, default: today)
    -v                  Verbose output (debug logging)

EXAMPLES:
    # Calculate tax from a JSON file
    tax-cli calculate production.json

    # Calculate with JSON output (for scripting)
    tax-cli calculate production.json --json

    # Calculate from stdin
    cat production.json | tax-cli calculate -

    # List all tax rates
    tax-cli rates

    # Get specific rate
    tax-cli rate beer barrel

    # Validate production data
    tax-cli validate production.json

    # Use custom jurisdiction and date
    tax-cli calculate production.json --jurisdiction california --date 2024-01-01

PRODUCTION DATA FORMAT:
    {
      "items": [
        {
          "product_type": "beer",
          "product_name": "IPA",
          "unit_type": "barrel",
          "quantity": 100,
          "abv": 6.5,
          "notes": "Optional notes"
        }
      ]
    }

    Product types: beer, wine, spirits, other
    Unit types:    gallon, barrel, case, liter

EXIT CODES:
    0 - Success
    1 - General error
    2 - Usage error
    3 - Validation error

`)
}
