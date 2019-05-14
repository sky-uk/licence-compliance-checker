package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/sky-uk/licence-compliance-checker/pkg/compliance"
	"github.com/sky-uk/licence-compliance-checker/pkg/detection"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

var rootCmd = &cobra.Command{
	Use:   "licence-compliance-checker",
	Short: "Check licences compliance based on list of restricted licences",
	Run:   validateCompliance,
}

var (
	overriddenLicences       map[string]string
	overriddenModuleLicences map[string]string
	ignoredProjects          []string
	restrictedLicences       []string
	logLevel                 string
	showComplianceErrors     bool
	showComplianceAll        bool
	checkGoModules           bool
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringSliceVarP(&ignoredProjects, "ignore-project", "i", []string{}, "project which licence will not be checked for compliance. Repeat this flag to specify multiple values.")
	rootCmd.PersistentFlags().StringToStringVarP(&overriddenLicences, "override-licence", "o", map[string]string{}, "can be used to override the licence detected for a project directory - e.g. vendor/github.com/spf13/cobra=MIT. Repeat this flag to specify multiple values.")
	rootCmd.PersistentFlags().StringToStringVarP(&overriddenModuleLicences, "override-module-licence", "m", map[string]string{}, "can be used to override the licence detected for a go module - e.g. github.com/spf13/cobra=MIT. Repeat this flag to specify multiple values.")
	rootCmd.PersistentFlags().StringSliceVarP(&restrictedLicences, "restricted-licence", "r", []string{}, "licence that will fail the compliance check if found for a project. Repeat this flag to specify multiple values.")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "L", "", "(output) should be one of: (none), debug, info, warn, error, fatal, panic. default (none)")
	rootCmd.PersistentFlags().BoolVarP(&showComplianceAll, "show-compliance-all", "A", false, "(output) to show compliance checks as JSON regardless of outcome")
	rootCmd.PersistentFlags().BoolVarP(&showComplianceErrors, "show-compliance-errors", "E", false, "(output) to show compliance checks as JSON only in case of errors")
	rootCmd.PersistentFlags().BoolVarP(&checkGoModules, "check-go-modules", "", false, "check all go modules a project depends on. This replaces specifying multiple project directories as positional arguments.")
	rootCmd.MarkPersistentFlagRequired("restricted-licence")
}

func validateCompliance(_ *cobra.Command, args []string) {
	setLogLevel(logLevel)

	if len(overriddenModuleLicences) > 0 && len(overriddenLicences) > 0 {
		logAndExit("Only use one of --override-module-licence (%d uses) and --override-licence (%d uses)", len(overriddenModuleLicences), len(overriddenLicences))
	}

	for module, licence := range overriddenModuleLicences {
		cmd := exec.Command("go", "list", "-m", "-f", "\"{{.Dir}}\"", module)

		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr

		err := cmd.Run()
		if err != nil {
			logAndExit("Failed to find directory for go module %s: %s %s (try setting GO111MODULE=on)", module, stderr.String(), err)
		}

		pkgDir := strings.Trim(strings.TrimSpace(out.String()), "\"")
		overriddenLicences[pkgDir] = licence
	}

	config := compliance.Config{
		RestrictedLicences:        restrictedLicences,
		IgnoredProjects:           ignoredProjects,
		OverriddenProjectLicences: overriddenLicences,
	}

	if checkGoModules {
		if len(args) > 0 {
			logAndExit("--check-go-modules and positional args cannot be set at the same time (received %d)", len(args))
		}

		var err error
		args, err = getGoModules()
		if err != nil {
			logAndExit("Failed to list go modules: %s", err)
		}
		log.Info("Found go modules:", args)
	} else {
		if len(args) == 0 {
			logAndExit("requires at least 1 arg (received %d)", len(args))
		}
	}

	log.Infof("Validating licence compliance with config: %v", config)
	c := compliance.New(&config, detection.NewLicenceDetector())
	result, err := c.Validate(args)
	if err != nil {
		logAndExit("Error validating licence compliance: %v", err)
	}
	log.Debugf("Licence compliance results: %v", result)

	if len(result.Restricted) > 0 || len(result.Unidentifiable) > 0 {
		if showComplianceErrors || showComplianceAll {
			printAsJSON(result)
		}
		logAndExit("Some licences are not compliant and/or cannot be identified: restricted: %v, unidentifiable: %v", result.Restricted, result.Unidentifiable)
	}

	if showComplianceAll {
		printAsJSON(result)
	}

	log.Info("Licences are compliant")
}

func getGoModules() ([]string, error) {
	cmd := exec.Command("go", "list", "-m", "-f", "\"{{.Dir}}\"", "all")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		logAndExit("Failed to list go modules: %s %s (try setting GO111MODULE=on)", stderr.String(), err)
	}

	modulePaths := strings.Split(out.String(), "\n")

	var r []string
	for _, str := range modulePaths {
		str = strings.Trim(strings.TrimSpace(str), "\"")
		if str != "" {
			r = append(r, str)
		}
	}
	return r, nil
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
	log.Errorf(message, args...)
	os.Exit(1)
}
