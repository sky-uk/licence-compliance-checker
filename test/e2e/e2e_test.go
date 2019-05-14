package e2e

import (
	"encoding/json"
	"flag"
	"fmt"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/licence-compliance-checker/pkg/compliance"
	"go/build"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var commandPath string
var testModulePath string

var junitReportDir string

func init() {
	flag.StringVar(&junitReportDir, "junit-report-dir", ".", "path to the directory that will contain the test reports")

	var err error
	commandPath, err = filepath.Abs("../../build/bin/licence-compliance-checker")
	if err != nil {
		fmt.Printf("Can't expand path to licence-compliance-checker binary: %s\n", err)
		os.Exit(1)
	}

	testModulePath, err = filepath.Abs("./testdata/go-module")
	if err != nil {
		fmt.Printf("Can't expand path to test module: %s\n", err)
		os.Exit(1)
	}
}

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)
	junitReporter := reporters.NewJUnitReporter(fmt.Sprintf("%s/e2e_test.xml", junitReportDir))
	RunSpecsWithDefaultAndCustomReporters(t, "E2E Test Suite", []Reporter{junitReporter})
}

var _ = Describe("License Compliance Checker", func() {

	It("should fail when project paths do not exist", func() {
		output, err := exec.Command(commandPath, "-A", "-r", "MIT", "testdata/does-not-exist").CombinedOutput()
		Expect(err).To(HaveOccurred())

		results := resultsFromJSON(string(output))
		Expect(results.Unidentifiable).To(HaveLen(1))
		Expect(results.Unidentifiable[0].Project).To(Equal("testdata/does-not-exist"))
		Expect(results.Compliant).To(BeNil())
		Expect(results.Ignored).To(BeNil())
		Expect(results.Restricted).To(BeNil())
	})

	It("should find project licence not compliant when is found in the restricted list", func() {
		output, err := exec.Command(commandPath, "-A", "-r", "MIT", "testdata/MIT", "testdata/BSD3").CombinedOutput()
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
		output, err := exec.Command(commandPath, "-A", "-r", "MIT", "testdata/no-licence").CombinedOutput()
		Expect(err).To(HaveOccurred())

		results := resultsFromJSON(string(output))
		Expect(results.Unidentifiable).To(HaveLen(1))
		Expect(results.Unidentifiable[0].Project).To(Equal("testdata/no-licence"))
	})

	It("should find project licences not on the restricted list to be compliant", func() {
		output, err := exec.Command(commandPath, "-A", "-r", "BSD", "testdata/MIT", "testdata/BSD3").CombinedOutput()
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
		output, err := exec.Command(commandPath, "-A", "-r", "MIT", "-o", "testdata/MIT=BSD", "testdata/MIT").CombinedOutput()
		Expect(err).NotTo(HaveOccurred())

		results := resultsFromJSON(string(output))
		Expect(results.Compliant).To(HaveLen(1))
		Expect(results.Compliant[0].Project).To(Equal("testdata/MIT"))
	})

	It("should find project licence compliant when the project using the restricted licence is ignored", func() {
		output, err := exec.Command(commandPath, "-A", "-r", "MIT", "-i", "testdata/MIT", "testdata/MIT", "testdata/BSD3").CombinedOutput()
		Expect(err).NotTo(HaveOccurred())

		results := resultsFromJSON(string(output))
		Expect(results.Ignored).To(HaveLen(1))
		Expect(results.Ignored[0].Project).To(Equal("testdata/MIT"))
		Expect(results.Compliant).To(HaveLen(1))
		Expect(results.Compliant[0].Project).To(Equal("testdata/BSD3"))
	})

	Context("output", func() {
		It("should not show anything with default options", func() {
			output, err := exec.Command(commandPath, "-r", "BSD", "testdata/MIT").CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(Equal(""))
		})

		It("should show only log messages when only log option is chosen", func() {
			output, err := exec.Command(commandPath, "-L", "info", "-r", "BSD", "testdata/MIT").CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(ContainSubstring("Licences are compliant"))
		})

		It("should not show anything when show-compliance-errors chosen but no compliance checks fail", func() {
			output, err := exec.Command(commandPath, "-E", "-r", "BSD", "testdata/MIT").CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(Equal(""))
		})

		It("should show json output when show-compliance-errors chosen and compliance checks fail", func() {
			output, err := exec.Command(commandPath, "-E", "-r", "MIT", "testdata/MIT").CombinedOutput()
			Expect(err).To(HaveOccurred())
			Expect(string(output)).To(Not(Equal("")))
		})

		It("should show json output when show-compliance-all chosen", func() {
			output, err := exec.Command(commandPath, "-A", "-r", "BSD", "testdata/MIT").CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(Not(Equal("")))
		})
	})

	Context("modules", func() {
		It("should check a project's modules", func() {
			cmd := exec.Command(commandPath, "-r", "BSD", "--check-go-modules")
			cmd.Dir = testModulePath
			cmd.Env = os.Environ()
			cmd.Env = append(cmd.Env, "GO111MODULE=on")

			output, err := cmd.CombinedOutput()

			Expect(err).NotTo(HaveOccurred())
			Expect(string(output)).To(Equal(""))
		})

		It("should fail for a non-compliant module", func() {
			cmd := exec.Command(commandPath, "-A", "-r", "BSD-3-Clause", "--check-go-modules")
			cmd.Dir = testModulePath
			cmd.Env = os.Environ()
			cmd.Env = append(cmd.Env, "GO111MODULE=on")

			output, err := cmd.CombinedOutput()
			Expect(err).To(HaveOccurred())

			results := resultsFromJSON(string(output))
			Expect(results.Restricted).To(HaveLen(5))
			Expect(results.Restricted[0].Project).To(ContainSubstring("golang.org/x/crypto"))
		})

		It("should fail with an overridden non-compliant module", func() {
			cmd := exec.Command(commandPath, "-A", "-r", "MIT", "-m", "golang.org/x/crypto=MIT", "--check-go-modules")
			cmd.Dir = testModulePath
			cmd.Env = os.Environ()
			cmd.Env = append(cmd.Env, "GO111MODULE=on")

			output, err := cmd.CombinedOutput()
			Expect(err).To(HaveOccurred())

			results := resultsFromJSON(string(output))
			Expect(results.Restricted).To(HaveLen(1))
			Expect(results.Restricted[0].Project).To(ContainSubstring("golang.org/x/crypto"))
		})

		It("should succeed with an overridden compliant module", func() {
			cmd := exec.Command(commandPath, "-A", "-r", "BSD-3-Clause", "-m", "golang.org/x/crypto=MIT", "--check-go-modules")
			cmd.Dir = testModulePath
			cmd.Env = os.Environ()
			cmd.Env = append(cmd.Env, "GO111MODULE=on")

			output, err := cmd.CombinedOutput()
			Expect(err).To(HaveOccurred())

			results := resultsFromJSON(string(output))
			Expect(results.Restricted).To(HaveLen(4))
			Expect(results.Restricted[0].Project).To(ContainSubstring("golang.org/x/net"))
			Expect(results.Restricted[3].Project).To(ContainSubstring("github.com/sky-uk/licence-compliance-checker"))
		})

		// Note: this test will fail if project isn't running inside the GOPATH
		It("should fail when run inside GOPATH and GO111MODULE=auto", func() {
			cmd := exec.Command(commandPath, "-L", "info", "-A", "-r", "BSD-3-Clause", "--check-go-modules")
			cmd.Dir = testModulePath
			cmd.Env = os.Environ()
			cmd.Env = append(cmd.Env, "GO111MODULE=auto")

			output, err := cmd.CombinedOutput()
			Expect(err).To(HaveOccurred())

			Expect(string(output)).To(ContainSubstring("not using modules"))
		})

		It("should allow an overridden module using positional arguments", func() {
			gopath := os.Getenv("GOPATH")
			if gopath == "" {
				gopath = build.Default.GOPATH
			}

			cmd := exec.Command(commandPath, "-A", "-r", "MIT", "-m", "golang.org/x/text=MIT", fmt.Sprintf("%s/%s", gopath, "pkg/mod/golang.org/x/text@v0.3.0"))
			cmd.Dir = testModulePath
			cmd.Env = os.Environ()
			cmd.Env = append(cmd.Env, "GO111MODULE=on")

			output, err := cmd.CombinedOutput()
			Expect(err).To(HaveOccurred())

			results := resultsFromJSON(string(output))
			Expect(results.Restricted).To(HaveLen(1))
			Expect(results.Restricted[0].Project).To(ContainSubstring("golang.org/x/text"))
		})
	})

})

func resultsFromJSON(document string) *compliance.Results {
	var v compliance.Results
	err := json.Unmarshal([]byte(document), &v)
	Expect(err).NotTo(HaveOccurred())
	return &v
}
