package detection

// LicenceDetector defines the behaviour for Licence detectors
type LicenceDetector interface {
	Detect(paths []string) ([]Result, error)
}

// NewLicenceDetector creates a new LicenceDetector
func NewLicenceDetector() LicenceDetector {
	return &goLicenseDetector{}
}

// Result is a representation of the Licence detection outcome for a project
type Result struct {
	Project string         `json:"project,omitempty"`
	Matches []LicenceMatch `json:"matches,omitempty"`
	ErrStr  string         `json:"error,omitempty"`
}

// LicenceMatch describes the level of confidence for the detected Licence
type LicenceMatch struct {
	Licence    string  `json:"license"`
	Confidence float32 `json:"confidence"`
}
