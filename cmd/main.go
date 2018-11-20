package main

import (
	"encoding/json"
	"fmt"
	"github.com/sky-uk/licence-compliance-checker/pkg/compliance"
	"github.com/sky-uk/licence-compliance-checker/pkg/detection"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "licence-compliance-checker",
	Short: "Check licences compliance based on list of restricted licences",
	Run:   validateCompliance,
	Args:  cobra.MinimumNArgs(1),
}

var (
	overriddenLicences   map[string]string
	ignoredProjects      []string
	restrictedLicences   []string
	logLevel             string
	showComplianceErrors bool
	showComplianceAll    bool
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringSliceVarP(&ignoredProjects, "ignore-project", "i", []string{}, "project which licence will not be checked for compliance. Repeat this flag to specify multiple values.")
	rootCmd.PersistentFlags().StringToStringVarP(&overriddenLicences, "override-licence", "o", map[string]string{}, "can be used to override the licence detected for a project - e.g. github.com/spf13/cobra=MIT. Repeat this flag to specify multiple values.")
	rootCmd.PersistentFlags().StringSliceVarP(&restrictedLicences, "restricted-licence", "r", []string{}, "licence that will fail the compliance check if found for a project. Repeat this flag to specify multiple values.")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "L", "", "(output) should be one of: (none), debug, info, warn, error, fatal, panic. default (none)")
	rootCmd.PersistentFlags().BoolVarP(&showComplianceAll, "show-compliance-all", "A", false, "(output) to show compliance checks as JSON regardless of outcome")
	rootCmd.PersistentFlags().BoolVarP(&showComplianceErrors, "show-compliance-errors", "E", false, "(output) to show compliance checks as JSON only in case of errors")
	rootCmd.MarkPersistentFlagRequired("restricted-licence")
}

func validateCompliance(_ *cobra.Command, args []string) {
	setLogLevel(logLevel)

	config := compliance.Config{
		RestrictedLicences:        restrictedLicences,
		IgnoredProjects:           ignoredProjects,
		OverriddenProjectLicences: overriddenLicences,
	}

	log.Infof("Validating licence compliance with config: %v", config)
	c := compliance.New(&config, &detection.GoLicenseDetector{})
	result, err := c.Validate(args)
	if err != nil {
		logAndExit("Error validating licence compliance: %v", err)
	}
	log.Debugf("Licence compliance results: %v", result)

	if len(result.Restricted) > 0 || len(result.Unidentifiable) > 0 {
		if showComplianceErrors || showComplianceAll {
			printAsJSON(result)
		}
		logAndExit("Some licences are not compliant or cannot be identified: %v", result)
	}

	if showComplianceAll {
		printAsJSON(result)
	}

	log.Info("Licences are compliant")
}

func printAsJSON(results *compliance.Results) {
	bytes, err := json.Marshal(results)
	if err != nil {
		logAndExit("Unable to marshal compliance checks results as json %v", err)
	}
	fmt.Println(string(bytes))
}

func setLogLevel(logLevel string) {
	if logLevel == "" {
		log.SetOutput(ioutil.Discard)
		return
	}
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		logAndExit("invalid log-level")
	}
	log.SetLevel(level)
}

func logAndExit(message string, args ...interface{}) {
	log.Errorf(message, args)
	os.Exit(1)
}
