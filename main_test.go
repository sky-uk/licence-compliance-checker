package main

import (
	"os/exec"
	"testing"

	"flag"
	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
)

var junitReportDir string

func init() {
	flag.StringVar(&junitReportDir, "junit-report-dir", ".", "path to the directory that will contain the test reports")
}

func TestCommandLine(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter(fmt.Sprintf("%s/command_line.xml", junitReportDir))
	RunSpecsWithDefaultAndCustomReporters(t, "Command Line Suite", []Reporter{junitReporter})
}

var _ = Describe("licence-compliance-checker command line", func() {
	Describe("--help", func() {
		It("should print available flags", func() {
			output, err := exec.Command("build/bin/licence-compliance-checker", "--help").CombinedOutput()
			Expect(err).ToNot(HaveOccurred())
			Expect(string(output)).To(ContainSubstring("-h, --help"))
			Expect(string(output)).To(ContainSubstring("-o, --override-licence"))
			Expect(string(output)).To(ContainSubstring("-i, --ignore-project"))
			Expect(string(output)).To(ContainSubstring("-r, --restricted-licence"))
			Expect(string(output)).To(ContainSubstring("-L, --log-level"))
			Expect(string(output)).To(ContainSubstring("-A, --show-compliance-all"))
			Expect(string(output)).To(ContainSubstring("-E, --show-compliance-errors"))
		})
	})
})
