package domain

import (
	"regexp"

	"github.com/example/grc-domain-models/domain/shared"
)

// FrameworkType represents the type of compliance framework.
type FrameworkType string

const (
	FrameworkTypeSOC2    FrameworkType = "SOC2"
	FrameworkTypeISO27001 FrameworkType = "ISO27001"
	FrameworkTypeHIPAA   FrameworkType = "HIPAA"
	FrameworkTypePCIDSS  FrameworkType = "PCI_DSS"
	FrameworkTypeGDPR    FrameworkType = "GDPR"
)

func (t FrameworkType) String() string {
	switch t {
	case FrameworkTypeSOC2:
		return "SOC 2"
	case FrameworkTypeISO27001:
		return "ISO 27001"
	case FrameworkTypeHIPAA:
		return "HIPAA"
	case FrameworkTypePCIDSS:
		return "PCI DSS"
	case FrameworkTypeGDPR:
		return "GDPR"
	default:
		return string(t)
	}
}

// FrameworkStatus represents the status of a framework.
type FrameworkStatus string

const (
	FrameworkStatusDraft      FrameworkStatus = "Draft"
	FrameworkStatusActive     FrameworkStatus = "Active"
	FrameworkStatusDeprecated FrameworkStatus = "Deprecated"
)

// Framework represents a compliance framework entity.
type Framework struct {
	id          shared.FrameworkID
	fwType      FrameworkType
	name        string
	version     string
	description string
	status      FrameworkStatus
	controlIDs  []shared.ControlID
}

// Getter methods
func (f *Framework) ID() shared.FrameworkID        { return f.id }
func (f *Framework) Type() FrameworkType           { return f.fwType }
func (f *Framework) Name() string                  { return f.name }
func (f *Framework) Version() string               { return f.version }
func (f *Framework) Description() string           { return f.description }
func (f *Framework) Status() FrameworkStatus       { return f.status }
func (f *Framework) ControlIDs() []shared.ControlID {
	// Return a copy to maintain immutability
	result := make([]shared.ControlID, len(f.controlIDs))
	copy(result, f.controlIDs)
	return result
}

// CreateFrameworkInput holds the input for creating a Framework.
type CreateFrameworkInput struct {
	ID          string
	Type        FrameworkType
	Name        string
	Version     string
	Description string
}

var semverPattern = regexp.MustCompile(`^\d+\.\d+(\.\d+)?$`)

// NewFramework creates a new Framework with validation.
func NewFramework(input CreateFrameworkInput) (*Framework, error) {
	var errors shared.ValidationErrors

	id, err := shared.NewFrameworkID(input.ID)
	if err != nil {
		if ve, ok := err.(shared.ValidationError); ok {
			errors = append(errors, ve)
		}
	}

	if input.Name == "" {
		errors.Add("name", "Framework name is required", "REQUIRED")
	}

	if !semverPattern.MatchString(input.Version) {
		errors.Add("version", "Version must be in semver format (e.g., 1.0 or 1.0.0)", "INVALID_VERSION")
	}

	if errors.HasErrors() {
		return nil, errors
	}

	return &Framework{
		id:          id,
		fwType:      input.Type,
		name:        input.Name,
		version:     input.Version,
		description: input.Description,
		status:      FrameworkStatusDraft,
		controlIDs:  []shared.ControlID{},
	}, nil
}

// WithStatus returns a new Framework with the updated status.
func (f *Framework) WithStatus(newStatus FrameworkStatus) (*Framework, error) {
	// Business rule: Cannot reactivate a deprecated framework
	if f.status == FrameworkStatusDeprecated && newStatus == FrameworkStatusActive {
		return nil, shared.NewValidationError(
			"status",
			"Cannot reactivate a deprecated framework",
			"INVALID_TRANSITION",
		)
	}

	// Business rule: Cannot activate a framework without controls
	if newStatus == FrameworkStatusActive && len(f.controlIDs) == 0 {
		return nil, shared.NewValidationError(
			"status",
			"Cannot activate a framework without controls",
			"NO_CONTROLS",
		)
	}

	controlIDsCopy := make([]shared.ControlID, len(f.controlIDs))
	copy(controlIDsCopy, f.controlIDs)

	return &Framework{
		id:          f.id,
		fwType:      f.fwType,
		name:        f.name,
		version:     f.version,
		description: f.description,
		status:      newStatus,
		controlIDs:  controlIDsCopy,
	}, nil
}

// WithControl returns a new Framework with the added control.
func (f *Framework) WithControl(controlID shared.ControlID) *Framework {
	// Check for duplicate
	for _, id := range f.controlIDs {
		if id == controlID {
			return f
		}
	}

	newControlIDs := make([]shared.ControlID, len(f.controlIDs)+1)
	copy(newControlIDs, f.controlIDs)
	newControlIDs[len(f.controlIDs)] = controlID

	return &Framework{
		id:          f.id,
		fwType:      f.fwType,
		name:        f.name,
		version:     f.version,
		description: f.description,
		status:      f.status,
		controlIDs:  newControlIDs,
	}
}
