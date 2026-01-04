package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// envelope is a generic wrapper for JSON responses
type envelope map[string]interface{}

// writeJSON writes JSON responses with proper headers
func (app *Application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	// Add any custom headers
	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(js)

	return err
}

// readJSON reads and decodes JSON from request body
func (app *Application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// Limit request body size to 1MB
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	// Check for additional JSON values
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return fmt.Errorf("body must only contain a single JSON value")
	}

	return nil
}

// readPathParam extracts a path parameter from the request
func (app *Application) readPathParam(r *http.Request, name string) string {
	return r.PathValue(name)
}

// readQueryParam extracts a query parameter with a default value
func (app *Application) readQueryParam(r *http.Request, name, defaultValue string) string {
	value := r.URL.Query().Get(name)
	if value == "" {
		return defaultValue
	}
	return value
}

// readQueryInt extracts an integer query parameter
func (app *Application) readQueryInt(r *http.Request, name string, defaultValue int) int {
	value := r.URL.Query().Get(name)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}

// readQueryBool extracts a boolean query parameter
func (app *Application) readQueryBool(r *http.Request, name string, defaultValue bool) bool {
	value := r.URL.Query().Get(name)
	if value == "" {
		return defaultValue
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return boolValue
}

// background runs a function in a background goroutine with panic recovery
func (app *Application) background(fn func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				app.logger.Error("panic in background goroutine", "error", err)
			}
		}()

		fn()
	}()
}

// getRequestID retrieves the request ID from the context
func (app *Application) getRequestID(r *http.Request) string {
	if id := r.Context().Value(requestIDKey); id != nil {
		if reqID, ok := id.(string); ok {
			return reqID
		}
	}
	return ""
}

// isValidXRPLAddress checks if a string is a valid XRPL address format
func isValidXRPLAddress(address string) bool {
	// XRPL addresses start with 'r' and are 25-35 characters long
	if !strings.HasPrefix(address, "r") {
		return false
	}
	if len(address) < 25 || len(address) > 35 {
		return false
	}
	return true
}

// isValidTxHash checks if a string is a valid transaction hash format
func isValidTxHash(hash string) bool {
	// Transaction hashes are 64 character hex strings
	if len(hash) != 64 {
		return false
	}
	for _, c := range hash {
		if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f')) {
			return false
		}
	}
	return true
}
