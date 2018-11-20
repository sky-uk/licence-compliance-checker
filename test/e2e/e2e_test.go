package e2e

import (
	"encoding/json"
	"flag"
	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/licence-compliance-checker/pkg/compliance"
	"os/exec"
	"testing"
)

var junitReportDir string

func init() {
	flag.StringVar(&junitReportDir, "junit-report-dir", ".", "path to the directory that will contain the test reports")
}

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter(fmt.Sprintf("%s/e2e_test.xml", junitReportDir))
	RunSpecsWithDefaultAndCustomReporters(t, "E2E Test Suite", []Reporter{junitReporter})
}

var _ = Describe("License Compliance Checker", func() {

	It("should fail when project paths do not exist", func() {
		output, err := exec.Command("licence-compliance-checker", "-A", "-r", "MIT", "testdata/does-not-exist").CombinedOutput()
		Expect(err).To(HaveOccurred())

		results := resultsFromJSON(string(output))
		Expect(results.Unidentifiable).To(HaveLen(1))
		Expect(results.Unidentifiable[0].Project).To(Equal("testdata/does-not-exist"))
		Expect(results.Compliant).To(BeNil())
		Expect(results.Ignored).To(BeNil())
		Expect(results.Restricted).To(BeNil())
	})

	It("should find project licence not compliant when is found in the restricted list", func() {
		output, err := exec.Command("licence-compliance-checker", "-A", "-r", "MIT", "testdata/MIT", "testdata/BSD3").CombinedOutput()
		Expect(err).To(HaveOccurred())

		results := resultsFromJSON(string(output))
		Expect(results.Restricted).To(HaveLen(1))
		Expect(results.Restricted[0].Project).To(Equal("testdata/MIT"))
		Expect(results.Compliant).To(HaveLen(1))
		Expect(results.Compliant[0].Project).To(Equal("testdata/BSD3"))
		Expect(results.Ignored).To(BeNil())
		Expect(results.Unidentifiable).To(BeNil())
	})

	It("should fail when project does not have license file", func() {
		output, err := exec.Command("licence-compliance-checker", "-A", "-r", "MIT", "testdata/no-licence").CombinedOutput()
		Expect(err).To(HaveOccurred())

		results := resultsFromJSON(string(output))
		Expect(results.Unidentifiable).To(HaveLen(1))
		Expect(results.Unidentifiable[0].Project).To(Equal("testdata/no-licence"))
	})

	It("should find project licences not on the restricted list to be compliant", func() {
		output, err := exec.Command("licence-compliance-checker", "-A", "-r", "BSD", "testdata/MIT", "testdata/BSD3").CombinedOutput()
		Expect(err).NotTo(HaveOccurred())

		results := resultsFromJSON(string(output))
		Expect(results.Compliant).To(HaveLen(2))
		Expect(results.Compliant[0].Project).To(Equal("testdata/BSD3"))
		Expect(results.Compliant[1].Project).To(Equal("testdata/MIT"))
		Expect(results.Ignored).To(BeNil())
		Expect(results.Restricted).To(BeNil())
		Expect(results.Unidentifiable).To(BeNil())
	})

	It("should find project licence compliant when the restricted licence is overridden", func() {
		output, err := exec.Command("licence-compliance-checker", "-A", "-r", "MIT", "-o", "testdata/MIT=BSD", "testdata/MIT").CombinedOutput()
		Expect(err).NotTo(HaveOccurred())

		results := resultsFromJSON(string(output))
		Expect(results.Compliant).To(HaveLen(1))
		Expect(results.Compliant[0].Project).To(Equal("testdata/MIT"))
	})

	It("should find project licence compliant when the project using the restricted licence is ignored", func() {
		output, err := exec.Command("licence-compliance-checker", "-A", "-r", "MIT", "-i", "testdata/MIT", "testdata/MIT", "testdata/BSD3").CombinedOutput()
		Expect(err).NotTo(HaveOccurred())

		results := resultsFromJSON(string(output))
		Expect(results.Ignored).To(HaveLen(1))
		Expect(results.Ignored[0].Project).To(Equal("testdata/MIT"))
		Expect(results.Compliant).To(HaveLen(1))
		Expect(results.Compliant[0].Project).To(Equal("testdata/BSD3"))
	})

	Context("output", func() {
		It("should not show anything with default options", func() {
			output, err := exec.Command("licence-compliance-checker", "-r", "BSD", "testdata/MIT").CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(Equal(""))
		})

		It("should show only log messages when only log option is chosen", func() {
			output, err := exec.Command("licence-compliance-checker", "-L", "info", "-r", "BSD", "testdata/MIT").CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(ContainSubstring("Licences are compliant"))
		})

		It("should not show anything when show-compliance-errors chosen but no compliance checks fail", func() {
			output, err := exec.Command("licence-compliance-checker", "-E", "-r", "BSD", "testdata/MIT").CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(Equal(""))
		})

		It("should show json output when show-compliance-errors chosen and compliance checks fail", func() {
			output, err := exec.Command("licence-compliance-checker", "-E", "-r", "MIT", "testdata/MIT").CombinedOutput()
			Expect(err).To(HaveOccurred())
			Expect(string(output)).To(Not(Equal("")))
		})

		It("should show json output when show-compliance-all chosen", func() {
			output, err := exec.Command("licence-compliance-checker", "-A", "-r", "BSD", "testdata/MIT").CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(Not(Equal("")))
		})
	})

})

func resultsFromJSON(document string) *compliance.Results {
	var v compliance.Results
	err := json.Unmarshal([]byte(document), &v)
	Expect(err).NotTo(HaveOccurred())
	return &v
}
