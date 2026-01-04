# Tax CLI - Excise Tax Calculator

A command-line tool for calculating federal excise taxes on alcoholic beverages. Follows Unix philosophy: small, focused, composable.

## Features

- **Calculate excise taxes** from production data (JSON format)
- **List tax rates** for different product types
- **Query specific rates** by product and unit type
- **Validate production data** before calculation
- **JSON output** for scripting and automation
- **Custom tax rates** via JSON configuration
- **Multiple jurisdictions** support
- **Date-specific rates** for historical or future calculations

## Installation

From the v2 directory:

```bash
# Build the CLI
go build -o tax-cli ./cmd/tax-cli

# Or install globally
go install ./cmd/tax-cli

# Verify installation
./tax-cli help
```

## Quick Start

### Calculate Tax

Create a production data file (`production.json`):

```json
{
  "items": [
    {
      "product_type": "beer",
      "product_name": "West Coast IPA",
      "unit_type": "barrel",
      "quantity": 50,
      "abv": 6.8,
      "notes": "Q1 2024 production"
    },
    {
      "product_type": "wine",
      "product_name": "Cabernet Sauvignon",
      "unit_type": "gallon",
      "quantity": 500,
      "abv": 13.5
    }
  ]
}
```

Calculate the tax:

```bash
./tax-cli calculate production.json
```

Output:
```
Tax Calculation Results
=======================

Jurisdiction:     federal
Calculated:       2024-01-15 10:30:00

Breakdown by Product:
---------------------
Product:  West Coast IPA (beer)
Quantity: 50.00 barrel @ $18.00 per barrel
Tax:      $900.00

Product:  Cabernet Sauvignon (wine)
Quantity: 500.00 gallon @ $1.07 per gallon
Tax:      $535.00

---------------------
Total Tax:        $1435.00
```

### List Available Rates

```bash
./tax-cli rates
```

### Get Specific Rate

```bash
./tax-cli rate beer barrel
./tax-cli rate wine gallon
./tax-cli rate spirits gallon
```

### Validate Production Data

```bash
./tax-cli validate production.json
```

## Usage

### Commands

- `calculate <file>` - Calculate tax from production data JSON file
- `rates` - List all available tax rates
- `rate <type> <unit>` - Get specific tax rate
- `validate <file>` - Validate production data JSON file
- `help` - Show help message

### Global Options

- `--json` - Output in JSON format (for scripting)
- `--rates <file>` - Path to custom tax rates JSON file
- `--jurisdiction <name>` - Tax jurisdiction (default: federal)
- `--date <YYYY-MM-DD>` - Effective date for calculation (default: today)
- `-v` - Verbose output (debug logging)

### Production Data Format

```json
{
  "items": [
    {
      "product_type": "beer|wine|spirits|other",
      "product_name": "Product Name",
      "unit_type": "gallon|barrel|case|liter",
      "quantity": 100,
      "abv": 6.5,
      "notes": "Optional notes"
    }
  ]
}
```

**Product Types:**
- `beer` - Beer and malt beverages
- `wine` - Wine and wine products
- `spirits` - Distilled spirits
- `other` - Other alcoholic beverages

**Unit Types:**
- `gallon` - US gallons
- `barrel` - Barrels (31 gallons for beer)
- `case` - Cases
- `liter` - Liters

## Examples

### Calculate with JSON Output

```bash
./tax-cli calculate production.json --json
```

Output:
```json
{
  "total_tax_amount": 1435.00,
  "jurisdiction": "federal",
  "calculated_at": "2024-01-15T10:30:00Z",
  "breakdown_by_product": [
    {
      "product_type": "beer",
      "product_name": "West Coast IPA",
      "quantity": 50,
      "unit_type": "barrel",
      "rate_per_unit": 18.00,
      "tax_amount": 900.00
    },
    {
      "product_type": "wine",
      "product_name": "Cabernet Sauvignon",
      "quantity": 500,
      "unit_type": "gallon",
      "rate_per_unit": 1.07,
      "tax_amount": 535.00
    }
  ]
}
```

### Unix Filter Pattern (stdin)

```bash
cat production.json | ./tax-cli calculate -
```

### Historical Calculation

```bash
./tax-cli calculate production.json --date 2020-01-01
```

### Custom Jurisdiction and Rates

Create `california_rates.json`:
```json
[
  {
    "product_type": "beer",
    "rate_per_unit": 0.20,
    "unit_type": "gallon",
    "effective_date": "2018-01-01T00:00:00Z",
    "expiration_date": null,
    "description": "California state excise tax on beer",
    "jurisdiction": "california"
  }
]
```

Calculate with custom rates:
```bash
./tax-cli calculate production.json \
  --rates california_rates.json \
  --jurisdiction california
```

### Scripting Example

```bash
#!/bin/bash
# Calculate taxes for all production files

for file in production/*.json; do
  echo "Calculating tax for $file..."
  
  # Calculate and extract total
  total=$(./tax-cli calculate "$file" --json | jq -r '.total_tax_amount')
  
  echo "Total tax: \$$total"
  echo "---"
done
```

## Default Federal Rates

The tool includes default federal excise tax rates (as of 2024):

| Product | Unit   | Rate per Unit | Description |
|---------|--------|---------------|-------------|
| Beer    | Barrel | $18.00        | Standard rate (>2M barrels/year) |
| Beer    | Gallon | $0.58         | Per gallon rate |
| Wine    | Gallon | $1.07         | Still wine (≤14% ABV) |
| Spirits | Gallon | $13.50        | Per proof gallon |

**Note:** These are example rates. Always verify current rates with the TTB (Alcohol and Tobacco Tax and Trade Bureau).

## Exit Codes

- `0` - Success
- `1` - General error (file not found, calculation failed, etc.)
- `2` - Usage error (invalid arguments, unknown command)
- `3` - Validation error (invalid production data)

## Testing

Run the test suite:

```bash
# From v2 directory
go test ./internal/tax/...

# With coverage
go test -cover ./internal/tax/...

# With race detector
go test -race ./internal/tax/...
```

## Design Philosophy

This tool follows the **Unix Philosophy**:

1. **Small is beautiful** - Does one thing well: tax calculation
2. **Make each program a filter** - Accepts JSON input, produces structured output
3. **Build a prototype quickly** - Simple, focused implementation
4. **Store data in text files** - JSON for input/output, easy to read and process
5. **Use software leverage** - Composable with other CLI tools (jq, grep, etc.)
6. **Portability over efficiency** - Standard library first, minimal dependencies

## Architecture

```
cmd/tax-cli/
  main.go              # CLI entry point, subcommands
  example_production.json

internal/tax/
  types.go             # Domain types (ProductType, TaxRate, etc.)
  calculator.go        # Pure tax calculation functions
  rates.go             # Tax rate loading and defaults
  errors.go            # Sentinel errors
  calculator_test.go   # Unit tests
  rates_test.go        # Rate loading tests
```

### Pure Functions

The core calculator uses pure functions with no side effects:

```go
func (c *Calculator) Calculate(
    ctx context.Context,
    data *ProductionData,
    jurisdiction string,
    effectiveDate time.Time,
) (*TaxCalculation, error)
```

This makes testing easy and behavior predictable.

## Contributing

When adding features:

1. Keep commands focused and simple
2. Write tests for all new functions
3. Update this README with examples
4. Follow Unix philosophy principles
5. Use standard library when possible

## Future Enhancements

- [ ] Unit conversions (gallons ↔ barrels ↔ liters)
- [ ] State tax rates database
- [ ] Tax credit calculations (small producer exemptions)
- [ ] CSV output format
- [ ] Interactive mode for manual entry
- [ ] Tax filing report generation
- [ ] Historical rate lookup by date range

## License

See LICENSE file in repository root.

## Related Tools

- `xrpl-cli` - XRPL blockchain operations
- `jq` - JSON processor for parsing output

## Support

For issues or questions, please file an issue on GitHub.
