package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"time"

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
