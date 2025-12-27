package shared

import (
	"net/url"
)

// ID types - using distinct types for type safety
// Go doesn't have branded types, so we use type aliases with validation

// FrameworkID represents a framework identifier.
type FrameworkID string

// ControlID represents a control identifier.
type ControlID string

// EvidenceID represents an evidence identifier.
type EvidenceID string

// RiskID represents a risk identifier.
type RiskID string

// UserID represents a user identifier.
type UserID string

// IntegrationID represents an integration identifier.
type IntegrationID string

// NewFrameworkID creates a validated FrameworkID.
func NewFrameworkID(value string) (FrameworkID, error) {
	if value == "" {
		return "", NewValidationError("id", "FrameworkID cannot be empty", "EMPTY_ID")
	}
	return FrameworkID(value), nil
}

// NewControlID creates a validated ControlID.
func NewControlID(value string) (ControlID, error) {
	if value == "" {
		return "", NewValidationError("id", "ControlID cannot be empty", "EMPTY_ID")
	}
	return ControlID(value), nil
}

// NewEvidenceID creates a validated EvidenceID.
func NewEvidenceID(value string) (EvidenceID, error) {
	if value == "" {
		return "", NewValidationError("id", "EvidenceID cannot be empty", "EMPTY_ID")
	}
	return EvidenceID(value), nil
}

// NewRiskID creates a validated RiskID.
func NewRiskID(value string) (RiskID, error) {
	if value == "" {
		return "", NewValidationError("id", "RiskID cannot be empty", "EMPTY_ID")
	}
	return RiskID(value), nil
}

// NewUserID creates a validated UserID.
func NewUserID(value string) (UserID, error) {
	if value == "" {
		return "", NewValidationError("id", "UserID cannot be empty", "EMPTY_ID")
	}
	return UserID(value), nil
}

// NewIntegrationID creates a validated IntegrationID.
func NewIntegrationID(value string) (IntegrationID, error) {
	if value == "" {
		return "", NewValidationError("id", "IntegrationID cannot be empty", "EMPTY_ID")
	}
	return IntegrationID(value), nil
}

// Percentage represents a value between 0 and 100.
// The struct is immutable - fields are unexported.
type Percentage struct {
	value int
}

// NewPercentage creates a validated Percentage.
func NewPercentage(value int) (Percentage, error) {
	if value < 0 || value > 100 {
		return Percentage{}, NewValidationError(
			"percentage",
			"Percentage must be between 0 and 100",
			"INVALID_PERCENTAGE",
		)
	}
	return Percentage{value: value}, nil
}

// Value returns the percentage value.
func (p Percentage) Value() int {
	return p.value
}

// URL represents a validated URL.
type URL struct {
	value string
}

// NewURL creates a validated URL.
func NewURL(value string) (URL, error) {
	_, err := url.ParseRequestURI(value)
	if err != nil {
		return URL{}, NewValidationError("url", "Invalid URL format", "INVALID_URL")
	}
	return URL{value: value}, nil
}

// String returns the URL string.
func (u URL) String() string {
	return u.value
}
