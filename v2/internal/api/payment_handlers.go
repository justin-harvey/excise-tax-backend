package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/maxfelker/excise-tax-backend/v2/internal/payment"
)

// handleCreatePayment creates a new payment
func (app *Application) handleCreatePayment(w http.ResponseWriter, r *http.Request) {
	var req payment.CreatePaymentRequest

	if err := app.readJSON(w, r, &req); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Create payment
	tx, err := app.paymentProcessor.CreatePayment(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, payment.ErrInvalidPaymentType):
			app.errorResponseJSON(w, r, http.StatusBadRequest, "Invalid Payment Type", err.Error())
		case errors.Is(err, payment.ErrInvalidAmount):
			app.errorResponseJSON(w, r, http.StatusBadRequest, "Invalid Amount", err.Error())
		default:
			var valErr *payment.ValidationError
			if errors.As(err, &valErr) {
				app.errorResponseJSON(w, r, http.StatusBadRequest, "Validation Error", valErr.Error())
			} else {
				app.serverErrorResponse(w, r, err)
			}
		}
		return
	}

	// Return created payment
	err = app.writeJSON(w, http.StatusCreated, envelope{"payment": tx}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// handleGetPayment retrieves a payment by ID
func (app *Application) handleGetPayment(w http.ResponseWriter, r *http.Request) {
	paymentID := r.PathValue("id")
	if paymentID == "" {
		app.errorResponseJSON(w, r, http.StatusBadRequest, "Invalid Request", "payment ID is required")
		return
	}

	tx, err := app.paymentProcessor.GetPayment(r.Context(), paymentID)
	if err != nil {
		if errors.Is(err, payment.ErrPaymentNotFound) {
			app.notFoundResponse(w, r)
		} else {
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"payment": tx}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// handleListPayments lists payments with optional filters
func (app *Application) handleListPayments(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	qs := r.URL.Query()

	filters := payment.ListPaymentsFilters{
		Limit:  app.readInt(qs, "limit", 10),
		Offset: app.readInt(qs, "offset", 0),
	}

	// Parse state filter
	if stateStr := qs.Get("state"); stateStr != "" {
		state := payment.TransactionState(stateStr)
		if !state.IsValid() {
			app.errorResponseJSON(w, r, http.StatusBadRequest, "Invalid State", "invalid transaction state")
			return
		}
		filters.State = &state
	}

	// Parse type filter
	if typeStr := qs.Get("type"); typeStr != "" {
		paymentType := payment.PaymentType(typeStr)
		if !paymentType.IsValid() {
			app.errorResponseJSON(w, r, http.StatusBadRequest, "Invalid Type", "invalid payment type")
			return
		}
		filters.Type = &paymentType
	}

	payments, err := app.paymentProcessor.ListPayments(r.Context(), filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{
		"payments": payments,
		"total":    len(payments),
	}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// handleVerifyPayment verifies a payment's current status
func (app *Application) handleVerifyPayment(w http.ResponseWriter, r *http.Request) {
	paymentID := r.PathValue("id")
	if paymentID == "" {
		app.errorResponseJSON(w, r, http.StatusBadRequest, "Invalid Request", "payment ID is required")
		return
	}

	tx, err := app.paymentProcessor.VerifyPayment(r.Context(), paymentID)
	if err != nil {
		if errors.Is(err, payment.ErrPaymentNotFound) {
			app.notFoundResponse(w, r)
		} else {
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"payment": tx}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// handleRetryPayment retries a failed payment
func (app *Application) handleRetryPayment(w http.ResponseWriter, r *http.Request) {
	paymentID := r.PathValue("id")
	if paymentID == "" {
		app.errorResponseJSON(w, r, http.StatusBadRequest, "Invalid Request", "payment ID is required")
		return
	}

	err := app.paymentProcessor.RetryPayment(r.Context(), paymentID)
	if err != nil {
		if errors.Is(err, payment.ErrPaymentNotFound) {
			app.notFoundResponse(w, r)
		} else {
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{
		"message": "payment retry initiated",
		"id":      paymentID,
	}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// handleGetPaymentEvents retrieves the event history for a payment
func (app *Application) handleGetPaymentEvents(w http.ResponseWriter, r *http.Request) {
	paymentID := r.PathValue("id")
	if paymentID == "" {
		app.errorResponseJSON(w, r, http.StatusBadRequest, "Invalid Request", "payment ID is required")
		return
	}

	tx, err := app.paymentProcessor.GetPayment(r.Context(), paymentID)
	if err != nil {
		if errors.Is(err, payment.ErrPaymentNotFound) {
			app.notFoundResponse(w, r)
		} else {
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{
		"payment_id": paymentID,
		"events":     tx.Events,
		"total":      len(tx.Events),
	}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// readInt reads an integer from query string with a default value
func (app *Application) readInt(qs map[string][]string, key string, defaultValue int) int {
	values, ok := qs[key]
	if !ok || len(values) == 0 {
		return defaultValue
	}

	var i int
	if err := json.Unmarshal([]byte(values[0]), &i); err != nil {
		return defaultValue
	}

	return i
}
