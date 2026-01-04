package api

import (
	"fmt"
	"log/slog"
	"net/http"
)

// errorResponse is the structure for error responses
type errorResponse struct {
	Error errorDetail `json:"error"`
}

// errorDetail contains the error details following RFC 7807
type errorDetail struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail,omitempty"`
	Instance string `json:"instance,omitempty"`
}

// logError logs an error message
func (app *Application) logError(r *http.Request, err error) {
	app.logger.Error("request error",
		slog.String("request_id", app.getRequestID(r)),
		slog.String("method", r.Method),
		slog.String("path", r.URL.Path),
		slog.String("error", err.Error()),
	)
}

// errorResponseJSON sends a JSON error response
func (app *Application) errorResponseJSON(w http.ResponseWriter, r *http.Request, status int, title, detail string) {
	errResp := errorResponse{
		Error: errorDetail{
			Type:     fmt.Sprintf("https://api.example.com/errors/%d", status),
			Title:    title,
			Status:   status,
			Detail:   detail,
			Instance: r.URL.Path,
		},
	}

	// Marshal to JSON
	if err := app.writeJSON(w, status, envelope{"error": errResp.Error}, nil); err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// serverErrorResponse sends a 500 Internal Server Error response
func (app *Application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)

	app.errorResponseJSON(w, r, http.StatusInternalServerError,
		"Internal Server Error",
		"The server encountered an unexpected error while processing your request.",
	)
}

// notFoundResponse sends a 404 Not Found response
func (app *Application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponseJSON(w, r, http.StatusNotFound,
		"Not Found",
		"The requested resource could not be found.",
	)
}

// methodNotAllowedResponse sends a 405 Method Not Allowed response
func (app *Application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponseJSON(w, r, http.StatusMethodNotAllowed,
		"Method Not Allowed",
		fmt.Sprintf("The %s method is not supported for this resource.", r.Method),
	)
}

// badRequestResponse sends a 400 Bad Request response
func (app *Application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponseJSON(w, r, http.StatusBadRequest,
		"Bad Request",
		err.Error(),
	)
}

// validationErrorResponse sends a 422 Unprocessable Entity response
func (app *Application) validationErrorResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponseJSON(w, r, http.StatusUnprocessableEntity,
		"Validation Failed",
		"One or more fields failed validation.",
	)
}

// rateLimitExceededResponse sends a 429 Too Many Requests response
func (app *Application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponseJSON(w, r, http.StatusTooManyRequests,
		"Rate Limit Exceeded",
		"You have exceeded the rate limit. Please try again later.",
	)
}

// invalidAuthenticationResponse sends a 401 Unauthorized response
func (app *Application) invalidAuthenticationResponse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("WWW-Authenticate", "Bearer")

	app.errorResponseJSON(w, r, http.StatusUnauthorized,
		"Unauthorized",
		"You must be authenticated to access this resource.",
	)
}

// invalidCredentialsResponse sends a 401 Unauthorized response for invalid credentials
func (app *Application) invalidCredentialsResponse(w http.ResponseWriter, r *http.Request) {
	app.errorResponseJSON(w, r, http.StatusUnauthorized,
		"Unauthorized",
		"Invalid authentication credentials.",
	)
}
