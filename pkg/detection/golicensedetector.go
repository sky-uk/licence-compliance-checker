package detection

import (
	golicensedetection "github.com/sebbonnet/go-license-detector/detection"
	log "github.com/sirupsen/logrus"
)

// goLicenseDetector is an implementation of LicenceDetector that uses `go-license-detector` to identify licences
type goLicenseDetector struct {
}

// Detect actually invokes `go-Licence-detector` to perform the licence detection
func (d *goLicenseDetector) Detect(paths []string) ([]Result, error) {
	var results []Result

	gldResults := golicensedetection.Detect(paths...)
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

	if gldResult.Err == nil {
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
