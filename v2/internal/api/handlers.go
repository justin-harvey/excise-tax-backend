package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/maxfelker/excise-tax-backend/v2/internal/tax"
	"github.com/maxfelker/excise-tax-backend/v2/pkg/xrpl"
)

// healthCheckHandler returns a simple health check
func (app *Application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	data := envelope{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
	}

	err := app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// readinessHandler checks if the service is ready to accept traffic
func (app *Application) readinessHandler(w http.ResponseWriter, r *http.Request) {
	// Check XRPL connection
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Try a simple ping by getting server info
	_, err := app.xrpl.GetServerInfo(ctx)
	if err != nil {
		app.errorResponseJSON(w, r, http.StatusServiceUnavailable,
			"Service Unavailable",
			"XRPL connection is not ready",
		)
		return
	}

	data := envelope{
		"status": "ready",
		"checks": envelope{
			"xrpl": "connected",
		},
	}

	err = app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// livenessHandler indicates if the service is alive
func (app *Application) livenessHandler(w http.ResponseWriter, r *http.Request) {
	data := envelope{
		"status":    "alive",
		"timestamp": time.Now().UTC(),
	}

	err := app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// infoHandler returns API version and capabilities
func (app *Application) infoHandler(w http.ResponseWriter, r *http.Request) {
	data := envelope{
		"version":        "1.0.0",
		"environment":    app.config.Env,
		"golang_version": runtime.Version(),
		"capabilities": []string{
			"xrpl_account_balance",
			"xrpl_account_info",
			"xrpl_account_transactions",
			"xrpl_transaction_details",
			"xrpl_transaction_status",
			"tax_calculation",
			"tax_rates",
			"tax_validation",
			"tax_product_types",
		},
	}

	err := app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// getAccountBalanceHandler retrieves the balance for an XRPL account
func (app *Application) getAccountBalanceHandler(w http.ResponseWriter, r *http.Request) {
	address := app.readPathParam(r, "address")

	// Validate address format
	if !isValidXRPLAddress(address) {
		app.badRequestResponse(w, r, errors.New("invalid XRPL address format"))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Get account info which includes balance
	accountInfo, err := app.xrpl.GetAccountInfo(ctx, address)
	if err != nil {
		if errors.Is(err, xrpl.ErrAccountNotFound) {
			app.notFoundResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"account":       address,
		"balance_xrp":   fmt.Sprintf("%.6f", float64(accountInfo.Balance)/1_000_000),
		"balance_drops": accountInfo.Balance,
	}

	err = app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// getAccountInfoHandler retrieves detailed info for an XRPL account
func (app *Application) getAccountInfoHandler(w http.ResponseWriter, r *http.Request) {
	address := app.readPathParam(r, "address")

	// Validate address format
	if !isValidXRPLAddress(address) {
		app.badRequestResponse(w, r, errors.New("invalid XRPL address format"))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	accountInfo, err := app.xrpl.GetAccountInfo(ctx, address)
	if err != nil {
		if errors.Is(err, xrpl.ErrAccountNotFound) {
			app.notFoundResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"account":              address,
		"balance_xrp":          fmt.Sprintf("%.6f", float64(accountInfo.Balance)/1_000_000),
		"balance_drops":        accountInfo.Balance,
		"sequence":             accountInfo.Sequence,
		"flags":                accountInfo.Flags,
		"owner_count":          accountInfo.OwnerCount,
		"previous_txn_id":      accountInfo.PreviousTxnID,
		"previous_txn_lgr_seq": accountInfo.PreviousTxnLgrSeq,
	}

	err = app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// getAccountTransactionsHandler retrieves transaction history for an account
func (app *Application) getAccountTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	address := app.readPathParam(r, "address")

	// Validate address format
	if !isValidXRPLAddress(address) {
		app.badRequestResponse(w, r, errors.New("invalid XRPL address format"))
		return
	}

	// Parse query parameters
	limit := app.readQueryInt(r, "limit", 10)
	if limit < 1 {
		limit = 1
	}
	if limit > 100 {
		limit = 100
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	transactions, err := app.xrpl.GetAccountTransactions(ctx, address, limit)
	if err != nil {
		if errors.Is(err, xrpl.ErrAccountNotFound) {
			app.notFoundResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"account":      address,
		"limit":        limit,
		"count":        len(transactions),
		"transactions": transactions,
	}

	err = app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// getTransactionHandler retrieves details of a specific transaction
func (app *Application) getTransactionHandler(w http.ResponseWriter, r *http.Request) {
	hash := app.readPathParam(r, "hash")

	// Validate hash format
	if !isValidTxHash(hash) {
		app.badRequestResponse(w, r, errors.New("invalid transaction hash format"))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	tx, err := app.xrpl.GetTransaction(ctx, hash)
	if err != nil {
		if errors.Is(err, xrpl.ErrTransactionNotFound) {
			app.notFoundResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"hash":        hash,
		"transaction": tx,
	}

	err = app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// getTransactionStatusHandler verifies the status of a transaction
func (app *Application) getTransactionStatusHandler(w http.ResponseWriter, r *http.Request) {
	hash := app.readPathParam(r, "hash")

	// Validate hash format
	if !isValidTxHash(hash) {
		app.badRequestResponse(w, r, errors.New("invalid transaction hash format"))
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	status, err := app.xrpl.VerifyTransaction(ctx, hash)
	if err != nil {
		if errors.Is(err, xrpl.ErrTransactionNotFound) {
			app.notFoundResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"hash":         hash,
		"validated":    status.Validated,
		"status":       status.Status,
		"ledger_index": status.LedgerIndex,
	}

	err = app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// getTaxRatesHandler returns current tax rates
func (app *Application) getTaxRatesHandler(w http.ResponseWriter, r *http.Request) {
	jurisdiction := r.URL.Query().Get("jurisdiction")

	rates, err := app.taxCalculator.GetRates(jurisdiction)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"rates": rates,
	}

	err = app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// calculateTaxHandler calculates tax for production data
func (app *Application) calculateTaxHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProductionData tax.ProductionData `json:"production_data"`
		Jurisdiction   string             `json:"jurisdiction"`
		EffectiveDate  *time.Time         `json:"effective_date"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Set defaults
	if input.Jurisdiction == "" {
		input.Jurisdiction = "federal"
	}

	effectiveDate := time.Now()
	if input.EffectiveDate != nil {
		effectiveDate = *input.EffectiveDate
	}

	// Perform calculation
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	calculation, err := app.taxCalculator.Calculate(ctx, &input.ProductionData, input.Jurisdiction, effectiveDate)
	if err != nil {
		if errors.Is(err, tax.ErrNoProductionItems) || errors.Is(err, tax.ErrInvalidProductType) {
			app.badRequestResponse(w, r, err)
			return
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	data := envelope{
		"calculation": calculation,
	}

	err = app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// validateProductionDataHandler validates production data
func (app *Application) validateProductionDataHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		ProductionData tax.ProductionData `json:"production_data"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Validate the production data
	validationResult := validateProductionData(&input.ProductionData)

	data := envelope{
		"valid":  validationResult.Valid,
		"errors": validationResult.Errors,
	}

	err = app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// getProductTypesHandler returns supported product types
func (app *Application) getProductTypesHandler(w http.ResponseWriter, r *http.Request) {
	productTypes := []envelope{
		{
			"type":        "beer",
			"description": "Beer and malt beverages",
			"unit_types":  []string{"gallon", "barrel"},
		},
		{
			"type":        "wine",
			"description": "Wine and wine products",
			"unit_types":  []string{"gallon", "liter"},
		},
		{
			"type":        "spirits",
			"description": "Distilled spirits",
			"unit_types":  []string{"gallon", "liter"},
		},
		{
			"type":        "other",
			"description": "Other alcoholic beverages",
			"unit_types":  []string{"gallon", "liter", "case"},
		},
	}

	data := envelope{
		"product_types": productTypes,
	}

	err := app.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// validateProductionData validates production data items
func validateProductionData(data *tax.ProductionData) tax.ValidationResult {
	var errors []tax.ValidationError

	if data == nil {
		return tax.ValidationResult{
			Valid:  false,
			Errors: []tax.ValidationError{{Field: "production_data", Message: "production data is required"}},
		}
	}

	if len(data.Items) == 0 {
		return tax.ValidationResult{
			Valid:  false,
			Errors: []tax.ValidationError{{Field: "items", Message: "at least one production item is required"}},
		}
	}

	for i, item := range data.Items {
		prefix := fmt.Sprintf("items[%d]", i)

		// Validate product type
		if item.ProductType == "" {
			errors = append(errors, tax.ValidationError{
				Field:   prefix + ".product_type",
				Message: "product type is required",
			})
		} else if !isValidProductType(item.ProductType) {
			errors = append(errors, tax.ValidationError{
				Field:   prefix + ".product_type",
				Message: "invalid product type. Must be one of: beer, wine, spirits, other",
			})
		}

		// Validate product name
		if item.ProductName == "" {
			errors = append(errors, tax.ValidationError{
				Field:   prefix + ".product_name",
				Message: "product name is required",
			})
		}

		// Validate unit type
		if item.UnitType == "" {
			errors = append(errors, tax.ValidationError{
				Field:   prefix + ".unit_type",
				Message: "unit type is required",
			})
		} else if !isValidUnitType(item.UnitType) {
			errors = append(errors, tax.ValidationError{
				Field:   prefix + ".unit_type",
				Message: "invalid unit type. Must be one of: gallon, barrel, case, liter",
			})
		}

		// Validate quantity
		if item.Quantity <= 0 {
			errors = append(errors, tax.ValidationError{
				Field:   prefix + ".quantity",
				Message: "quantity must be greater than 0",
			})
		}

		// Validate ABV if provided
		if item.ABV < 0 || item.ABV > 100 {
			errors = append(errors, tax.ValidationError{
				Field:   prefix + ".abv",
				Message: "ABV must be between 0 and 100",
			})
		}
	}

	return tax.ValidationResult{
		Valid:  len(errors) == 0,
		Errors: errors,
	}
}

// isValidProductType checks if the product type is valid
func isValidProductType(productType tax.ProductType) bool {
	switch productType {
	case tax.ProductTypeBeer, tax.ProductTypeWine, tax.ProductTypeSpirits, tax.ProductTypeOther:
		return true
	default:
		return false
	}
}

// isValidUnitType checks if the unit type is valid
func isValidUnitType(unitType tax.UnitType) bool {
	switch unitType {
	case tax.UnitTypeGallon, tax.UnitTypeBarrel, tax.UnitTypeCase, tax.UnitTypeLiter:
		return true
	default:
		return false
	}
}
