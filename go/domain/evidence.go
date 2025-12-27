package domain

import (
	"fmt"
	"time"

	"github.com/example/grc-domain-models/domain/shared"
)

// FileType represents document file types.
type FileType string

const (
	FileTypePDF  FileType = "PDF"
	FileTypeDOCX FileType = "DOCX"
	FileTypeXLSX FileType = "XLSX"
	FileTypePNG  FileType = "PNG"
	FileTypeJPG  FileType = "JPG"
)

// CheckResult represents the result of an automated check.
type CheckResult interface {
	checkResult()
	String() string
}

type CheckPassed struct{}

func (CheckPassed) checkResult()    {}
func (CheckPassed) String() string { return "Passed" }

type CheckFailed struct {
	Reason string
}

func (CheckFailed) checkResult()     {}
func (c CheckFailed) String() string { return fmt.Sprintf("Failed: %s", c.Reason) }

type CheckSkipped struct {
	Reason string
}

func (CheckSkipped) checkResult()     {}
func (c CheckSkipped) String() string { return fmt.Sprintf("Skipped: %s", c.Reason) }

// EvidenceType represents the type of evidence.
// Uses the sealed interface pattern.
type EvidenceType interface {
	evidenceType()
	String() string
}

type Document struct {
	FileURL  shared.URL
	FileType FileType
}

func (Document) evidenceType() {}
func (d Document) String() string {
	return fmt.Sprintf("Document (%s)", d.FileType)
}

type Screenshot struct {
	ImageURL   shared.URL
	CapturedAt time.Time
}

func (Screenshot) evidenceType() {}
func (s Screenshot) String() string {
	return fmt.Sprintf("Screenshot (captured at %s)", s.CapturedAt.Format(time.RFC3339))
}

type AutomatedCheck struct {
	IntegrationID shared.IntegrationID
	CheckName     string
	LastRunAt     time.Time
	Result        CheckResult
}

func (AutomatedCheck) evidenceType() {}
func (a AutomatedCheck) String() string {
	return fmt.Sprintf("Automated Check: %s (%s)", a.CheckName, a.Result.String())
}

type ManualReview struct {
	ReviewerID shared.UserID
	ReviewedAt time.Time
	Notes      string
}

func (ManualReview) evidenceType() {}
func (m ManualReview) String() string {
	return fmt.Sprintf("Manual Review (reviewed at %s)", m.ReviewedAt.Format(time.RFC3339))
}

// MatchEvidenceType provides pattern matching for EvidenceType.
func MatchEvidenceType[T any](
	et EvidenceType,
	onDocument func(shared.URL, FileType) T,
	onScreenshot func(shared.URL, time.Time) T,
	onAutomatedCheck func(shared.IntegrationID, string, time.Time, CheckResult) T,
	onManualReview func(shared.UserID, time.Time, string) T,
) T {
	switch e := et.(type) {
	case Document:
		return onDocument(e.FileURL, e.FileType)
	case Screenshot:
		return onScreenshot(e.ImageURL, e.CapturedAt)
	case AutomatedCheck:
		return onAutomatedCheck(e.IntegrationID, e.CheckName, e.LastRunAt, e.Result)
	case ManualReview:
		return onManualReview(e.ReviewerID, e.ReviewedAt, e.Notes)
	default:
		panic(fmt.Sprintf("unknown EvidenceType: %T", et))
	}
}

// EvidenceStatus represents the status of evidence.
type EvidenceStatus string

const (
	EvidenceStatusValid    EvidenceStatus = "Valid"
	EvidenceStatusExpired  EvidenceStatus = "Expired"
	EvidenceStatusPending  EvidenceStatus = "Pending"
	EvidenceStatusRejected EvidenceStatus = "Rejected"
)

// Evidence represents a piece of compliance evidence.
type Evidence struct {
	id           shared.EvidenceID
	controlID    shared.ControlID
	evidenceType EvidenceType
	collectedAt  time.Time
	expiresAt    *time.Time // nil means no expiration
	description  string
}

// Getter methods
func (e *Evidence) ID() shared.EvidenceID      { return e.id }
func (e *Evidence) ControlID() shared.ControlID { return e.controlID }
func (e *Evidence) EvidenceType() EvidenceType  { return e.evidenceType }
func (e *Evidence) CollectedAt() time.Time      { return e.collectedAt }
func (e *Evidence) ExpiresAt() *time.Time       { return e.expiresAt }
func (e *Evidence) Description() string         { return e.description }

// CreateEvidenceInput holds the input for creating Evidence.
type CreateEvidenceInput struct {
	ID           string
	ControlID    shared.ControlID
	EvidenceType EvidenceType
	CollectedAt  time.Time
	ExpiresAt    *time.Time
	Description  string
}

// NewEvidence creates a new Evidence with validation.
func NewEvidence(input CreateEvidenceInput) (*Evidence, error) {
	var errors shared.ValidationErrors

	id, err := shared.NewEvidenceID(input.ID)
	if err != nil {
		if ve, ok := err.(shared.ValidationError); ok {
			errors = append(errors, ve)
		}
	}

	now := time.Now()

	// Validate expiration date
	if input.ExpiresAt != nil && input.ExpiresAt.Before(now) {
		errors.Add("expiresAt", "Expiration date must be in the future", "INVALID_EXPIRATION")
	}

	// Validate collection date
	if input.CollectedAt.After(now) {
		errors.Add("collectedAt", "Collection date cannot be in the future", "INVALID_COLLECTION_DATE")
	}

	if errors.HasErrors() {
		return nil, errors
	}

	return &Evidence{
		id:           id,
		controlID:    input.ControlID,
		evidenceType: input.EvidenceType,
		collectedAt:  input.CollectedAt,
		expiresAt:    input.ExpiresAt,
		description:  input.Description,
	}, nil
}

// Status calculates the current status of the evidence.
func (e *Evidence) Status() EvidenceStatus {
	now := time.Now()

	// Check expiration
	if e.expiresAt != nil && e.expiresAt.Before(now) {
		return EvidenceStatusExpired
	}

	// Check automated check result
	if ac, ok := e.evidenceType.(AutomatedCheck); ok {
		switch ac.Result.(type) {
		case CheckFailed:
			return EvidenceStatusRejected
		case CheckSkipped:
			return EvidenceStatusPending
		}
	}

	return EvidenceStatusValid
}

// GetEvidenceTypeLabel returns a localized label for the evidence type.
func GetEvidenceTypeLabel(et EvidenceType) string {
	return MatchEvidenceType(
		et,
		func(_ shared.URL, ft FileType) string { return fmt.Sprintf("ドキュメント (%s)", ft) },
		func(_ shared.URL, capturedAt time.Time) string {
			return fmt.Sprintf("スクリーンショット (%s)", capturedAt.Format(time.RFC3339))
		},
		func(_ shared.IntegrationID, checkName string, _ time.Time, result CheckResult) string {
			return fmt.Sprintf("自動チェック: %s (%s)", checkName, result.String())
		},
		func(_ shared.UserID, reviewedAt time.Time, _ string) string {
			return fmt.Sprintf("手動レビュー (%s)", reviewedAt.Format(time.RFC3339))
		},
	)
}
