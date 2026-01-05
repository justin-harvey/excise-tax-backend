package payment

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// NewEvent creates a new immutable payment event
func NewEvent(transactionID string, eventType EventType, state TransactionState, details interface{}) (PaymentEvent, error) {
	if transactionID == "" {
		return PaymentEvent{}, &ValidationError{Field: "transaction_id", Message: "transaction ID is required"}
	}

	if eventType == "" {
		return PaymentEvent{}, &ValidationError{Field: "type", Message: "event type is required"}
	}

	if !state.IsValid() {
		return PaymentEvent{}, ErrInvalidState
	}

	var detailsJSON json.RawMessage
	if details != nil {
		data, err := json.Marshal(details)
		if err != nil {
			return PaymentEvent{}, fmt.Errorf("failed to marshal event details: %w", err)
		}
		detailsJSON = data
	}

	return PaymentEvent{
		ID:            uuid.New().String(),
		TransactionID: transactionID,
		Type:          eventType,
		State:         state,
		Details:       detailsJSON,
		Timestamp:     time.Now().UTC(),
	}, nil
}

// AppendEvent adds a new event to a payment transaction
// Events are immutable - this creates a new event and appends it
func AppendEvent(tx *PaymentTransaction, eventType EventType, newState TransactionState, details interface{}) error {
	if tx == nil {
		return ErrPaymentNotFound
	}

	// Validate state transition
	if err := validateStateTransition(tx.State, newState); err != nil {
		return err
	}

	// Create new event
	event, err := NewEvent(tx.ID, eventType, newState, details)
	if err != nil {
		return err
	}

	// Append event to transaction
	tx.Events = append(tx.Events, event)
	tx.State = newState
	tx.UpdatedAt = time.Now().UTC()

	return nil
}

// validateStateTransition checks if a state transition is valid
func validateStateTransition(from, to TransactionState) error {
	// Allow same state (for events that don't change state)
	if from == to {
		return nil
	}

	validTransitions := map[TransactionState][]TransactionState{
		StatePending: {
			StateCompleted,
			StateInReview,
			StateFailed,
		},
		StateInReview: {
			StateCompleted,
			StateFailed,
			StatePending, // Can go back to pending after review
		},
		StateFailed: {
			StatePending, // Can retry
		},
		StateCompleted: {
			// Terminal state - no transitions
		},
	}

	validNextStates, ok := validTransitions[from]
	if !ok {
		return ErrInvalidTransition
	}

	for _, validState := range validNextStates {
		if to == validState {
			return nil
		}
	}

	return fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidTransition, from, to)
}

// GetEventsByType returns all events of a specific type
func GetEventsByType(tx *PaymentTransaction, eventType EventType) []PaymentEvent {
	if tx == nil {
		return nil
	}

	var filtered []PaymentEvent
	for _, event := range tx.Events {
		if event.Type == eventType {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// GetLatestEvent returns the most recent event
func GetLatestEvent(tx *PaymentTransaction) *PaymentEvent {
	if tx == nil || len(tx.Events) == 0 {
		return nil
	}
	return &tx.Events[len(tx.Events)-1]
}

// CountEventsByType returns the count of events of a specific type
func CountEventsByType(tx *PaymentTransaction, eventType EventType) int {
	if tx == nil {
		return 0
	}

	count := 0
	for _, event := range tx.Events {
		if event.Type == eventType {
			count++
		}
	}
	return count
}
