package detection

import (
	log "github.com/sirupsen/logrus"
	golicensedetection "gopkg.in/src-d/go-license-detector.v2/licensedb"
)

// goLicenseDetector is an implementation of LicenceDetector that uses `go-license-detector` to identify licences
type goLicenseDetector struct {
}

// Detect actually invokes `go-Licence-detector` to perform the licence detection
func (d *goLicenseDetector) Detect(paths []string) ([]Result, error) {
	var results []Result

	gldResults := golicensedetection.Analyse(paths...)
	log.Debugf("Licence detection raw results from go-license-detector: %v", gldResults)
	for _, gldResult := range gldResults {
		results = append(results, buildResultFrom(gldResult))
	}

	log.Tracef("Licence detection mapped results: %v", results)
	return results, nil
}

func buildResultFrom(gldResult golicensedetection.Result) Result {
	result := Result{
		Project: gldResult.Arg,
		ErrStr:  gldResult.ErrStr,
	}

	if gldResult.ErrStr == "" {
		result.Matches = buildLicenceMatchesFrom(gldResult.Matches)
	}
	return result
}

func buildLicenceMatchesFrom(gldMatches []golicensedetection.Match) []LicenceMatch {
	var licenceMatches []LicenceMatch
	for _, gldMatch := range gldMatches {
		licenceMatches = append(licenceMatches, LicenceMatch{
			Licence:    gldMatch.License,
			Confidence: gldMatch.Confidence,
		})
	}
	return licenceMatches
}
