package api

import (
	"net/http"
)

// routes sets up all the HTTP routes for the API
func (app *Application) routes() http.Handler {
	mux := http.NewServeMux()

	// Health check endpoints
	mux.HandleFunc("GET /health", app.healthCheckHandler)
	mux.HandleFunc("GET /health/ready", app.readinessHandler)
	mux.HandleFunc("GET /health/live", app.livenessHandler)

	// API info endpoint
	mux.HandleFunc("GET /info", app.infoHandler)

	// XRPL Account endpoints
	mux.HandleFunc("GET /xrpl/accounts/{address}/balance", app.getAccountBalanceHandler)
	mux.HandleFunc("GET /xrpl/accounts/{address}/info", app.getAccountInfoHandler)
	mux.HandleFunc("GET /xrpl/accounts/{address}/transactions", app.getAccountTransactionsHandler)

	// XRPL Transaction endpoints
	mux.HandleFunc("GET /xrpl/transactions/{hash}", app.getTransactionHandler)
	mux.HandleFunc("GET /xrpl/transactions/{hash}/status", app.getTransactionStatusHandler)

	// Tax endpoints
	mux.HandleFunc("GET /tax/rates", app.getTaxRatesHandler)
	mux.HandleFunc("POST /tax/calculate", app.calculateTaxHandler)
	mux.HandleFunc("POST /tax/validate", app.validateProductionDataHandler)
	mux.HandleFunc("GET /tax/product-types", app.getProductTypesHandler)

	// Payment endpoints
	mux.HandleFunc("POST /payments", app.handleCreatePayment)
	mux.HandleFunc("GET /payments/{id}", app.handleGetPayment)
	mux.HandleFunc("GET /payments", app.handleListPayments)
	mux.HandleFunc("GET /payments/{id}/verify", app.handleVerifyPayment)
	mux.HandleFunc("POST /payments/{id}/retry", app.handleRetryPayment)
	mux.HandleFunc("GET /payments/{id}/events", app.handleGetPaymentEvents)

	// Wrap with middleware
	return app.recoverPanic(app.rateLimit(app.enableCORS(app.logRequest(mux))))
}
