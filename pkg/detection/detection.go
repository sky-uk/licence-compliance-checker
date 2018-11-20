package detection

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os/exec"
)

// LicenceDetector defines the behaviour for Licence detectors
type LicenceDetector interface {
	Detect(paths []string) ([]Result, error)
}

// GoLicenseDetector is an implementation of LicenceDetector that uses `go-license-detector` to identify licences
type GoLicenseDetector struct {
}

// Detect actually invokes `go-Licence-detector` to perform the licence detection
func (d *GoLicenseDetector) Detect(paths []string) ([]Result, error) {
	var args []string
	args = append(args, "-f")
	args = append(args, "json")
	args = append(args, paths...)
	cmd := exec.Command("license-detector", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error while executing command with args %v: %v", args, err)
	}
	log.Debugf("Licence detection raw results: %s", string(output))

	var results []Result
	if err := json.Unmarshal(output, &results); err != nil {
		return nil, fmt.Errorf("error while unmarshalling detection output %s: %v", string(output), err)
	}
	log.Tracef("Licence detection marshalled results: %v", results)
	return results, nil
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
