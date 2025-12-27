package domain

import (
	"fmt"
	"time"

	"github.com/example/grc-domain-models/domain/shared"
)

// RiskLevel represents the severity level of a risk.
type RiskLevel int

const (
	RiskLevelLow RiskLevel = iota + 1
	RiskLevelMedium
	RiskLevelHigh
	RiskLevelCritical
)

func (l RiskLevel) String() string {
	switch l {
	case RiskLevelLow:
		return "Low"
	case RiskLevelMedium:
		return "Medium"
	case RiskLevelHigh:
		return "High"
	case RiskLevelCritical:
		return "Critical"
	default:
		return "Unknown"
	}
}

// RiskScore is an immutable value object representing a risk score.
type RiskScore struct {
	likelihood RiskLevel
	impact     RiskLevel
	value      int
	label      string
}

// CalculateRiskScore creates a new RiskScore from likelihood and impact.
func CalculateRiskScore(likelihood, impact RiskLevel) RiskScore {
	value := int(likelihood) * int(impact)
	var label string
	switch {
	case value <= 2:
		label = "Low"
	case value <= 6:
		label = "Medium"
	case value <= 12:
		label = "High"
	default:
		label = "Critical"
	}
	return RiskScore{
		likelihood: likelihood,
		impact:     impact,
		value:      value,
		label:      label,
	}
}

// Getter methods for RiskScore
func (r RiskScore) Likelihood() RiskLevel { return r.likelihood }
func (r RiskScore) Impact() RiskLevel     { return r.impact }
func (r RiskScore) Value() int            { return r.value }
func (r RiskScore) Label() string         { return r.label }

// RiskCategory represents the category of a risk.
type RiskCategory string

const (
	RiskCategoryOperational RiskCategory = "Operational"
	RiskCategoryTechnical   RiskCategory = "Technical"
	RiskCategoryCompliance  RiskCategory = "Compliance"
	RiskCategoryFinancial   RiskCategory = "Financial"
)

// RiskStatus represents the status of a risk.
// Uses the sealed interface pattern.
type RiskStatus interface {
	riskStatus()
	String() string
}

type Identified struct {
	IdentifiedAt time.Time
}

func (Identified) riskStatus() {}
func (s Identified) String() string {
	return fmt.Sprintf("Identified (%s)", s.IdentifiedAt.Format(time.RFC3339))
}

type Assessed struct {
	AssessedAt time.Time
	AssessorID shared.UserID
}

func (Assessed) riskStatus() {}
func (s Assessed) String() string {
	return fmt.Sprintf("Assessed (%s)", s.AssessedAt.Format(time.RFC3339))
}

type Mitigated struct {
	MitigatedAt time.Time
	ControlIDs  []shared.ControlID
}

func (Mitigated) riskStatus() {}
func (s Mitigated) String() string {
	return fmt.Sprintf("Mitigated (%d controls)", len(s.ControlIDs))
}

type Accepted struct {
	AcceptedByID shared.UserID
	Reason       string
	ExpiresAt    time.Time
}

func (Accepted) riskStatus() {}
func (s Accepted) String() string {
	return fmt.Sprintf("Accepted: %s (expires %s)", s.Reason, s.ExpiresAt.Format(time.RFC3339))
}

type Closed struct {
	ClosedAt   time.Time
	Resolution string
}

func (Closed) riskStatus() {}
func (s Closed) String() string {
	return fmt.Sprintf("Closed: %s", s.Resolution)
}

// MatchRiskStatus provides pattern matching for RiskStatus.
func MatchRiskStatus[T any](
	status RiskStatus,
	onIdentified func(time.Time) T,
	onAssessed func(time.Time, shared.UserID) T,
	onMitigated func(time.Time, []shared.ControlID) T,
	onAccepted func(shared.UserID, string, time.Time) T,
	onClosed func(time.Time, string) T,
) T {
	switch s := status.(type) {
	case Identified:
		return onIdentified(s.IdentifiedAt)
	case Assessed:
		return onAssessed(s.AssessedAt, s.AssessorID)
	case Mitigated:
		return onMitigated(s.MitigatedAt, s.ControlIDs)
	case Accepted:
		return onAccepted(s.AcceptedByID, s.Reason, s.ExpiresAt)
	case Closed:
		return onClosed(s.ClosedAt, s.Resolution)
	default:
		panic(fmt.Sprintf("unknown RiskStatus: %T", status))
	}
}

// Risk represents a compliance risk entity.
type Risk struct {
	id            shared.RiskID
	title         string
	description   string
	category      RiskCategory
	inherentScore RiskScore
	residualScore RiskScore
	status        RiskStatus
	ownerID       shared.UserID
}

// Getter methods
func (r *Risk) ID() shared.RiskID        { return r.id }
func (r *Risk) Title() string            { return r.title }
func (r *Risk) Description() string      { return r.description }
func (r *Risk) Category() RiskCategory   { return r.category }
func (r *Risk) InherentScore() RiskScore { return r.inherentScore }
func (r *Risk) ResidualScore() RiskScore { return r.residualScore }
func (r *Risk) Status() RiskStatus       { return r.status }
func (r *Risk) OwnerID() shared.UserID   { return r.ownerID }

// CreateRiskInput holds the input for creating a Risk.
type CreateRiskInput struct {
	ID          string
	Title       string
	Description string
	Category    RiskCategory
	Likelihood  RiskLevel
	Impact      RiskLevel
	OwnerID     shared.UserID
}

// NewRisk creates a new Risk with validation.
func NewRisk(input CreateRiskInput) (*Risk, error) {
	var errors shared.ValidationErrors

	id, err := shared.NewRiskID(input.ID)
	if err != nil {
		if ve, ok := err.(shared.ValidationError); ok {
			errors = append(errors, ve)
		}
	}

	if input.Title == "" {
		errors.Add("title", "Risk title is required", "REQUIRED")
	}

	if errors.HasErrors() {
		return nil, errors
	}

	inherentScore := CalculateRiskScore(input.Likelihood, input.Impact)

	return &Risk{
		id:            id,
		title:         input.Title,
		description:   input.Description,
		category:      input.Category,
		inherentScore: inherentScore,
		residualScore: inherentScore, // Initially the same
		status:        Identified{IdentifiedAt: time.Now()},
		ownerID:       input.OwnerID,
	}, nil
}

// WithStatus returns a new Risk with the updated status.
func (r *Risk) WithStatus(newStatus RiskStatus) (*Risk, error) {
	// Business rule: Cannot transition from Closed status
	if _, isClosed := r.status.(Closed); isClosed {
		return nil, shared.NewValidationError(
			"status",
			"Cannot transition from Closed status",
			"INVALID_TRANSITION",
		)
	}

	// Business rule: Accepted expiration must be in the future
	if accepted, ok := newStatus.(Accepted); ok {
		if accepted.ExpiresAt.Before(time.Now()) {
			return nil, shared.NewValidationError(
				"expiresAt",
				"Acceptance expiration date must be in the future",
				"INVALID_EXPIRATION",
			)
		}
	}

	return &Risk{
		id:            r.id,
		title:         r.title,
		description:   r.description,
		category:      r.category,
		inherentScore: r.inherentScore,
		residualScore: r.residualScore,
		status:        newStatus,
		ownerID:       r.ownerID,
	}, nil
}

// WithResidualScore returns a new Risk with the updated residual score.
func (r *Risk) WithResidualScore(likelihood, impact RiskLevel) *Risk {
	return &Risk{
		id:            r.id,
		title:         r.title,
		description:   r.description,
		category:      r.category,
		inherentScore: r.inherentScore,
		residualScore: CalculateRiskScore(likelihood, impact),
		status:        r.status,
		ownerID:       r.ownerID,
	}
}

// GetRiskStatusLabel returns a localized label for the risk status.
func GetRiskStatusLabel(status RiskStatus) string {
	return MatchRiskStatus(
		status,
		func(t time.Time) string { return fmt.Sprintf("特定済み (%s)", t.Format(time.RFC3339)) },
		func(t time.Time, _ shared.UserID) string { return fmt.Sprintf("評価済み (%s)", t.Format(time.RFC3339)) },
		func(_ time.Time, controlIDs []shared.ControlID) string {
			return fmt.Sprintf("軽減済み (%d件の統制)", len(controlIDs))
		},
		func(_ shared.UserID, reason string, expiresAt time.Time) string {
			return fmt.Sprintf("受容 (%s, 期限: %s)", reason, expiresAt.Format(time.RFC3339))
		},
		func(_ time.Time, resolution string) string { return fmt.Sprintf("クローズ (%s)", resolution) },
	)
}
