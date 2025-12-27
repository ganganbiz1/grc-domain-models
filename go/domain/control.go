// Package domain provides GRC domain models.
package domain

import (
	"fmt"
	"time"

	"github.com/example/grc-domain-models/domain/shared"
)

// ControlStatus represents the status of a control.
// Go doesn't have union types, so we use an interface with unexported method
// to create a "sealed" interface pattern.
type ControlStatus interface {
	controlStatus() // unexported method prevents external implementations
	String() string
}

// NotImplemented represents a control that hasn't been implemented.
type NotImplemented struct{}

func (NotImplemented) controlStatus() {}
func (NotImplemented) String() string { return "Not Implemented" }

// InProgress represents a control that is being implemented.
type InProgress struct {
	Progress shared.Percentage
}

func (InProgress) controlStatus() {}
func (s InProgress) String() string {
	return fmt.Sprintf("In Progress (%d%%)", s.Progress.Value())
}

// Implemented represents a control that has been implemented.
type Implemented struct {
	ImplementedAt time.Time
}

func (Implemented) controlStatus() {}
func (s Implemented) String() string {
	return fmt.Sprintf("Implemented (%s)", s.ImplementedAt.Format(time.RFC3339))
}

// NotApplicable represents a control that is not applicable.
type NotApplicable struct {
	Reason string
}

func (NotApplicable) controlStatus() {}
func (s NotApplicable) String() string {
	return fmt.Sprintf("Not Applicable: %s", s.Reason)
}

// Failed represents a control that has failed.
type Failed struct {
	Reason     string
	DetectedAt time.Time
}

func (Failed) controlStatus() {}
func (s Failed) String() string {
	return fmt.Sprintf("Failed: %s (detected at %s)", s.Reason, s.DetectedAt.Format(time.RFC3339))
}

// MatchControlStatus provides exhaustive pattern matching for ControlStatus.
// Note: Go cannot guarantee compile-time exhaustiveness.
func MatchControlStatus[T any](
	status ControlStatus,
	onNotImplemented func() T,
	onInProgress func(shared.Percentage) T,
	onImplemented func(time.Time) T,
	onNotApplicable func(string) T,
	onFailed func(string, time.Time) T,
) T {
	switch s := status.(type) {
	case NotImplemented:
		return onNotImplemented()
	case InProgress:
		return onInProgress(s.Progress)
	case Implemented:
		return onImplemented(s.ImplementedAt)
	case NotApplicable:
		return onNotApplicable(s.Reason)
	case Failed:
		return onFailed(s.Reason, s.DetectedAt)
	default:
		panic(fmt.Sprintf("unknown ControlStatus type: %T", status))
	}
}

// Control represents a compliance control entity.
// Fields are unexported to ensure immutability.
type Control struct {
	id          shared.ControlID
	frameworkID shared.FrameworkID
	code        string
	title       string
	description string
	status      ControlStatus
	ownerID     shared.UserID
}

// Getter methods for Control
func (c *Control) ID() shared.ControlID          { return c.id }
func (c *Control) FrameworkID() shared.FrameworkID { return c.frameworkID }
func (c *Control) Code() string                  { return c.code }
func (c *Control) Title() string                 { return c.title }
func (c *Control) Description() string           { return c.description }
func (c *Control) Status() ControlStatus         { return c.status }
func (c *Control) OwnerID() shared.UserID        { return c.ownerID }

// CreateControlInput holds the input for creating a Control.
type CreateControlInput struct {
	ID          string
	FrameworkID shared.FrameworkID
	Code        string
	Title       string
	Description string
	OwnerID     shared.UserID
}

// NewControl creates a new Control with validation.
func NewControl(input CreateControlInput) (*Control, error) {
	var errors shared.ValidationErrors

	id, err := shared.NewControlID(input.ID)
	if err != nil {
		if ve, ok := err.(shared.ValidationError); ok {
			errors = append(errors, ve)
		}
	}

	if input.Code == "" {
		errors.Add("code", "Control code is required", "REQUIRED")
	}

	if input.Title == "" {
		errors.Add("title", "Control title is required", "REQUIRED")
	}

	if errors.HasErrors() {
		return nil, errors
	}

	return &Control{
		id:          id,
		frameworkID: input.FrameworkID,
		code:        input.Code,
		title:       input.Title,
		description: input.Description,
		status:      NotImplemented{},
		ownerID:     input.OwnerID,
	}, nil
}

// WithStatus returns a new Control with the updated status.
// This preserves immutability by creating a new instance.
func (c *Control) WithStatus(newStatus ControlStatus) (*Control, error) {
	// Business rule: Cannot transition directly from Failed to Implemented
	if _, isFailed := c.status.(Failed); isFailed {
		if _, isImplemented := newStatus.(Implemented); isImplemented {
			return nil, shared.NewValidationError(
				"status",
				"Cannot transition directly from Failed to Implemented",
				"INVALID_TRANSITION",
			)
		}
	}

	return &Control{
		id:          c.id,
		frameworkID: c.frameworkID,
		code:        c.code,
		title:       c.title,
		description: c.description,
		status:      newStatus,
		ownerID:     c.ownerID,
	}, nil
}

// GetControlStatusLabel returns a localized label for the control status.
func GetControlStatusLabel(status ControlStatus) string {
	return MatchControlStatus(
		status,
		func() string { return "未実装" },
		func(p shared.Percentage) string { return fmt.Sprintf("実装中 (%d%%)", p.Value()) },
		func(t time.Time) string { return fmt.Sprintf("実装済み (%s)", t.Format(time.RFC3339)) },
		func(reason string) string { return fmt.Sprintf("適用外: %s", reason) },
		func(reason string, _ time.Time) string { return fmt.Sprintf("失敗: %s", reason) },
	)
}
