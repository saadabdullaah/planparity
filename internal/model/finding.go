package model

// VerificationState communicates how directly a conclusion was observed.
type VerificationState string

const (
	VerificationVerified    VerificationState = "VERIFIED"
	VerificationInferred    VerificationState = "INFERRED"
	VerificationUnverified  VerificationState = "UNVERIFIED"
	VerificationUnsupported VerificationState = "UNSUPPORTED"
	VerificationFailed      VerificationState = "FAILED"
)

// Severity is a finding's review priority.
type Severity string

const (
	SeverityInfo     Severity = "INFO"
	SeverityWarning  Severity = "WARNING"
	SeverityHigh     Severity = "HIGH"
	SeverityCritical Severity = "CRITICAL"
	SeverityUnknown  Severity = "UNKNOWN"
)

// FindingCategory identifies the reason for a finding.
type FindingCategory string

const (
	CategoryCompatibility     FindingCategory = "COMPATIBILITY"
	CategoryDestructiveChange FindingCategory = "DESTRUCTIVE_CHANGE"
	CategoryResourceIdentity  FindingCategory = "RESOURCE_IDENTITY"
	CategoryProviderSchema    FindingCategory = "PROVIDER_SCHEMA"
	CategoryEngineDifference  FindingCategory = "ENGINE_DIFFERENCE"
	CategoryStateIntegrity    FindingCategory = "STATE_INTEGRITY"
	CategorySafetyLimitation  FindingCategory = "SAFETY_LIMITATION"
)

// Evidence explains the inputs behind a finding without carrying values.
type Evidence struct {
	Kind    string `json:"kind"`
	Message string `json:"message"`
}

// Finding is the stable analysis finding contract.
type Finding struct {
	ID             string            `json:"id"`
	Category       FindingCategory   `json:"category"`
	Severity       Severity          `json:"severity"`
	Verification   VerificationState `json:"verification"`
	Confidence     float64           `json:"confidence"`
	Address        string            `json:"address,omitempty"`
	Title          string            `json:"title"`
	Explanation    string            `json:"explanation"`
	Evidence       []Evidence        `json:"evidence"`
	Recommendation string            `json:"recommendation"`
}
