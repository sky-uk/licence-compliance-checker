package compliance

import (
	"flag"
	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/sky-uk/licence-compliance-checker/pkg/detection"
	"reflect"
	"testing"
)

var junitReportDir string

func init() {
	flag.StringVar(&junitReportDir, "junit-report-dir", ".", "path to the directory that will contain the test reports")
}

func TestCompliance(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter(fmt.Sprintf("%s/compliance.xml", junitReportDir))
	RunSpecsWithDefaultAndCustomReporters(t, "Compliance Suite", []Reporter{junitReporter})
}

var _ = Describe("compliance check", func() {

	It("should find all projects to comply when no restricted licences", func() {
		// given
		licenceDetector := newFakeLicenceDetector(
			aProjectWithLicence("project1", map[string]float32{"MIT": 0.9}),
			aProjectWithLicence("project2", map[string]float32{"BSD": 0.9}),
		)
		c := New(&Config{}, licenceDetector)

		// when
		results, err := c.Validate([]string{"project1", "project2"})

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(results.Compliant).To(HaveLen(2))
		Expect(results.Compliant).To(HaveProjectLicences("project1", "MIT"))
		Expect(results.Compliant).To(HaveProjectLicences("project2", "BSD"))
		Expect(results.Restricted).To(HaveLen(0))
	})

	It("should find projects not complying with restricted licences", func() {
		// given
		licenceDetector := newFakeLicenceDetector(
			aProjectWithLicence("project1", map[string]float32{"MIT": 0.9}),
			aProjectWithLicence("project2", map[string]float32{"BSD": 0.9}),
			aProjectWithLicence("project3", map[string]float32{"BSD3": 0.9}),
		)
		c := New(&Config{RestrictedLicences: []string{"MIT", "BSD2", "BSD3"}}, licenceDetector)

		// when
		results, err := c.Validate([]string{"project1", "project2", "project3"})

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(results.Compliant).To(HaveLen(1))
		Expect(results.Compliant).To(HaveProjectLicences("project2", "BSD"))
		Expect(results.Restricted).To(HaveLen(2))
		Expect(results.Restricted).To(HaveProjectLicences("project1", "MIT"))
		Expect(results.Restricted).To(HaveProjectLicences("project3", "BSD3"))
	})

	It("should use the most probable licence when checking restrictions", func() {
		// given
		licenceDetector := newFakeLicenceDetector(
			aProjectWithLicence("project1", map[string]float32{"MIT": 0.9, "BSD": 0.7}),
			aProjectWithLicence("project2", map[string]float32{"BSD2": 0.9, "MIT": 0.91}),
			aProjectWithLicence("project3", map[string]float32{"BSD3": 0.91, "MIT": 0.9}),
		)
		c := New(&Config{RestrictedLicences: []string{"MIT"}}, licenceDetector)

		// when
		results, err := c.Validate([]string{"project1", "project2", "project3"})

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(results.Compliant).To(HaveLen(1))
		Expect(results.Compliant).To(HaveProjectLicences("project3", "BSD3", "MIT"))
		Expect(results.Restricted).To(HaveLen(2))
		Expect(results.Restricted).To(HaveProjectLicences("project1", "MIT", "BSD"))
		Expect(results.Restricted).To(HaveProjectLicences("project2", "MIT", "BSD2"))
	})

	It("should find project licences unidentifiable when the licence cannot be detected", func() {
		// given
		licenceDetector := newFakeLicenceDetector(
			aProjectWithNoLicence("project1"),
			aProjectWithNoLicence("project2"),
		)
		c := New(&Config{}, licenceDetector)

		// when
		results, err := c.Validate([]string{"project1", "project1"})

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(results.Unidentifiable).To(HaveLen(2))
		Expect(results.Unidentifiable).To(HaveNoProjectLicences("project1"))
		Expect(results.Unidentifiable).To(HaveNoProjectLicences("project2"))
	})

	It("should order the project alphabetically", func() {
		// given
		licenceDetector := newFakeLicenceDetector(
			aProjectWithNoLicence("Ab"),
			aProjectWithNoLicence("Aa"),
			aProjectWithNoLicence("C"),
			aProjectWithLicence("B", map[string]float32{"MIT": 0.9, "BSD": 0.7}),
			aProjectWithLicence("A", map[string]float32{"MIT": 0.9, "BSD": 0.7}),
			aProjectWithLicence("B1", map[string]float32{"MIT": 0.9, "BSD": 0.7}),
			aProjectWithLicence("D", map[string]float32{"BSD": 0.7}),
			aProjectWithLicence("F", map[string]float32{"BSD": 0.7}),
			aProjectWithLicence("E", map[string]float32{"BSD": 0.7}),
		)
		c := New(&Config{RestrictedLicences: []string{"MIT"}}, licenceDetector)

		// when
		results, err := c.Validate([]string{"Ab", "Aa", "C", "B", "A", "B1", "B", "D", "F", "E"})

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(results.Unidentifiable).To(HaveLen(3))
		Expect(results.Unidentifiable[0].Project).To(Equal("Aa"))
		Expect(results.Unidentifiable[1].Project).To(Equal("Ab"))
		Expect(results.Unidentifiable[2].Project).To(Equal("C"))
		Expect(results.Compliant).To(HaveLen(3))
		Expect(results.Compliant[0].Project).To(Equal("D"))
		Expect(results.Compliant[1].Project).To(Equal("E"))
		Expect(results.Compliant[2].Project).To(Equal("F"))
		Expect(results.Restricted).To(HaveLen(3))
		Expect(results.Restricted[0].Project).To(Equal("A"))
		Expect(results.Restricted[1].Project).To(Equal("B"))
		Expect(results.Restricted[2].Project).To(Equal("B1"))
	})

	It("should order licences with highest confidence first then alphabetically", func() {
		// given
		licenceDetector := newFakeLicenceDetector(
			aProjectWithLicence("C1", map[string]float32{"MPL2.0": 0.9, "MPL2.0-no-copyleft": 0.9, "Y": 0.92, "Z": 0.92}),
			aProjectWithLicence("C2", map[string]float32{"MPL2.0-no-copyleft": 0.9, "MPL2.0": 0.9, "Z": 0.89, "Y": 0.89}),
			aProjectWithLicence("C3", map[string]float32{"Z": 0.91, "MIT-other": 0.9, "MIT": 0.9}),
			aProjectWithLicence("R", map[string]float32{"Z": 0.89, "MIT-other": 0.9, "MIT": 0.9}),
			aProjectWithLicence("I", map[string]float32{"MPL2.0-no-copyleft": 0.9, "MPL2.0": 0.9, "Z": 0.91}),
		)

		c := New(&Config{RestrictedLicences: []string{"MIT"}, IgnoredProjects: []string{"I"}}, licenceDetector)

		// when
		results, err := c.Validate([]string{"A", "B"})

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(results.Compliant).To(HaveLen(3))
		Expect(results.Compliant).To(HaveProjectLicences("C1", "Y", "Z", "MPL2.0", "MPL2.0-no-copyleft"))
		Expect(results.Compliant).To(HaveProjectLicences("C2", "MPL2.0", "MPL2.0-no-copyleft", "Y", "Z"))
		Expect(results.Compliant).To(HaveProjectLicences("C3", "Z", "MIT", "MIT-other"))
		Expect(results.Restricted).To(HaveProjectLicences("R", "MIT", "MIT-other", "Z"))
		Expect(results.Ignored).To(HaveProjectLicences("I", "Z", "MPL2.0", "MPL2.0-no-copyleft"))
	})

	Context("when a licence is overridden for a project", func() {

		It("should be used when identifying Restricted licences", func() {
			// given
			licenceDetector := newFakeLicenceDetector(
				aProjectWithLicence("project1", map[string]float32{"MIT": 0.9}),
				aProjectWithLicence("project2", map[string]float32{"MIT": 0.9, "BSD3": 0.1}),
			)
			c := New(&Config{RestrictedLicences: []string{"MIT"}, OverriddenProjectLicences: map[string]string{"project2": "BSD"}}, licenceDetector)

			// when
			results, err := c.Validate([]string{"project1", "project1"})

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(results.Compliant).To(HaveLen(1))
			Expect(results.Compliant).To(HaveProjectLicences("project2", "BSD"))
			Expect(results.Restricted).To(HaveLen(1))
			Expect(results.Restricted).To(HaveProjectLicences("project1", "MIT"))
		})

		It("should be used even though no licence were found for that project", func() {
			// given
			licenceDetector := newFakeLicenceDetector(
				aProjectWithNoLicence("project1"),
				aProjectWithNoLicence("project2"),
			)
			c := New(&Config{RestrictedLicences: []string{"MIT"}, OverriddenProjectLicences: map[string]string{"project1": "MIT", "project2": "BSD"}}, licenceDetector)

			// when
			results, err := c.Validate([]string{"project1", "project1"})

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(results.Compliant).To(HaveLen(1))
			Expect(results.Compliant).To(HaveProjectLicences("project2", "BSD"))
			Expect(results.Restricted).To(HaveLen(1))
			Expect(results.Restricted).To(HaveProjectLicences("project1", "MIT"))
		})
	})

	Context("when a project is ignored", func() {
		It("licence restrictions check do not apply", func() {
			// given
			licenceDetector := newFakeLicenceDetector(
				aProjectWithNoLicence("project1"),
				aProjectWithLicence("project2", map[string]float32{"MIT": 0.9}),
				aProjectWithLicence("project3", map[string]float32{"MIT": 0.9}),
			)
			c := New(&Config{RestrictedLicences: []string{"MIT"}, IgnoredProjects: []string{"project1", "project2"}}, licenceDetector)

			// when
			results, err := c.Validate([]string{"project1", "project2", "project3"})

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(results.Ignored).To(HaveLen(2))
			Expect(results.Ignored).To(HaveProjectLicences("project1"))
			Expect(results.Ignored).To(HaveProjectLicences("project2", "MIT"))
			Expect(results.Restricted).To(HaveLen(1))
			Expect(results.Restricted).To(HaveProjectLicences("project3", "MIT"))
		})
	})

})

func aProjectWithLicence(project string, licencesConfidence map[string]float32) detection.Result {
	var matches []detection.LicenceMatch
	for licence, confidence := range licencesConfidence {
		matches = append(matches, detection.LicenceMatch{Licence: licence, Confidence: confidence})
	}
	return detection.Result{Project: project, Matches: matches}
}

func aProjectWithNoLicence(project string) detection.Result {
	return detection.Result{Project: project, ErrStr: "no licence found"}
}

type FakeLicenceDetector struct {
	detectionResults []detection.Result
}

func newFakeLicenceDetector(result ...detection.Result) *FakeLicenceDetector {
	return &FakeLicenceDetector{result}
}

func (d *FakeLicenceDetector) Detect(paths []string) ([]detection.Result, error) {
	return d.detectionResults, nil
}

func HaveNoProjectLicences(project string) types.GomegaMatcher {
	return &haveProjectLicences{project, nil}
}

func HaveProjectLicences(project string, licences ...string) types.GomegaMatcher {
	return &haveProjectLicences{project, licences}
}

type haveProjectLicences struct {
	project  string
	licences []string
}

func (matcher *haveProjectLicences) Match(actual interface{}) (success bool, err error) {
	detectionResults := actual.([]detection.Result)

	for _, result := range detectionResults {
		if result.Project == matcher.project {
			var licences []string
			for _, licence := range result.Matches {
				licences = append(licences, licence.Licence)
			}
			if reflect.DeepEqual(licences, matcher.licences) {
				return true, nil
			}
		}
	}
	return false, nil
}

func (matcher *haveProjectLicences) FailureMessage(actual interface{}) (message string) {
	results := actual.([]detection.Result)
	return fmt.Sprintf("Expected detection results to contain project %s with licences %v. Actual: %v", matcher.project, matcher.licences, results)
}

func (matcher *haveProjectLicences) NegatedFailureMessage(actual interface{}) (message string) {
	results := actual.([]detection.Result)
	return fmt.Sprintf("Expected detection results not to contain project %s with licences %v. Actual: %v", matcher.project, matcher.licences, results)
}
