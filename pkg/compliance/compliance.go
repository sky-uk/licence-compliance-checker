package compliance

import (
	log "github.com/sirupsen/logrus"
	"github.com/sky-uk/licence-compliance-checker/pkg/detection"
	"sort"
)

// Config holds configuration values for the compliance check
type Config struct {
	IgnoredProjects           []string
	RestrictedLicences        []string
	OverriddenProjectLicences map[string]string
}

// Compliance exposes method to validate the licences compliance
type Compliance struct {
	config          *Config
	licenceDetector detection.LicenceDetector
}

// Results results of the compliance checks
type Results struct {
	Compliant      []detection.Result `json:"compliant"`
	Restricted     []detection.Result `json:"restricted"`
	Unidentifiable []detection.Result `json:"unidentifiable"`
	Ignored        []detection.Result `json:"ignored"`
}

// New creates a new compliance checker
func New(config *Config, licenceDetector detection.LicenceDetector) *Compliance {
	return &Compliance{config: config, licenceDetector: licenceDetector}
}

// Validate performs the licence compliance checks against the given project paths
func (c *Compliance) Validate(projectPaths []string) (*Results, error) {
	var complianceResults Results
	detectionResults, err := c.licenceDetector.Detect(projectPaths)
	if err != nil {
		return nil, err
	}

	sort.Slice(detectionResults, func(i, j int) bool {
		return detectionResults[i].Project < detectionResults[j].Project
	})

	for _, detectionResult := range detectionResults {
		if c.projectIgnored(detectionResult) {
			complianceResults.Ignored = append(complianceResults.Ignored, detectionResult)
			continue
		}

		if licenceOverride, ok := c.config.OverriddenProjectLicences[detectionResult.Project]; ok {
			detectionResult = detection.Result{
				Project: detectionResult.Project,
				Matches: []detection.LicenceMatch{{Licence: licenceOverride, Confidence: 0}},
			}
		}

		if detectionResult.ErrStr != "" {
			complianceResults.Unidentifiable = append(complianceResults.Unidentifiable, detectionResult)
			continue
		}

		if c.restrictedLicence(detectionResult) {
			complianceResults.Restricted = append(complianceResults.Restricted, detectionResult)
			continue
		}
		complianceResults.Compliant = append(complianceResults.Compliant, detectionResult)
	}
	return &complianceResults, nil
}

func (c *Compliance) restrictedLicence(detectionResult detection.Result) bool {
	sort.Slice(detectionResult.Matches, func(i, j int) bool {
		return detectionResult.Matches[i].Confidence > detectionResult.Matches[j].Confidence
	})

	mostProbableLicence := detectionResult.Matches[0].Licence
	for _, restrictedLicence := range c.config.RestrictedLicences {
		if mostProbableLicence == restrictedLicence {
			log.Infof("Project '%s' most probable license '%s' is restricted", detectionResult.Project, mostProbableLicence)
			return true
		}
	}
	return false
}

func (c *Compliance) projectIgnored(detectionResult detection.Result) bool {
	for _, ignored := range c.config.IgnoredProjects {
		if ignored == detectionResult.Project {
			return true
		}
	}
	return false
}
