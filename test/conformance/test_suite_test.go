package conformance

import (
	"flag"
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/clean"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/reporters"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"

	testclient "github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/client"

	// Test files in this package must not end with `_test.go` suffix, as they are imported as go package
	_ "github.com/k8snetworkplumbingwg/sriov-network-operator/test/conformance/tests"

	"github.com/k8snetworkplumbingwg/sriov-network-operator/test/util/k8sreporter"
)

var (
	junitPath      *string
	dumpOutput     *bool
	reporterFile   string
	customReporter *k8sreporter.KubernetesReporter
)

func init() {
	dumpOutput = flag.Bool("dump", false, "dump informations for failed tests")
	junitPath = flag.String("junit", "", "the path for the junit format report")
}

func TestTest(t *testing.T) {
	RegisterFailHandler(Fail)

	reporterFile = os.Getenv("REPORTER_OUTPUT")

	clients := testclient.New("")

	if reporterFile != "" {
		f, err := os.OpenFile(reporterFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open the file: %v\n", err)
			return
		}
		defer f.Close()
		customReporter = k8sreporter.New(clients, f)
	} else if *dumpOutput {
		customReporter = k8sreporter.New(clients, os.Stdout)
	}

	RunSpecs(t, "SRIOV Operator conformance tests")
}

var _ = ReportAfterSuite("conformance", func(report types.Report) {
	if *junitPath != "" {
		junitFile := path.Join(*junitPath, "junit_sriov_conformance.xml")
		reporters.GenerateJUnitReportWithConfig(report, junitFile, reporters.JunitReportConfig{
			OmitTimelinesForSpecState: types.SpecStatePassed | types.SpecStateSkipped,
			OmitLeafNodeType:          true,
			OmitSuiteSetupNodes:       true,
		})
	}
})

var _ = ReportAfterEach(func(sr types.SpecReport) {
	if sr.Failed() == false {
		return
	}

	if reporterFile != "" || *dumpOutput {
		customReporter.Report(sr)
	}
})

var _ = BeforeSuite(func() {
	err := clean.All()
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	err := clean.All()
	Expect(err).NotTo(HaveOccurred())
})
