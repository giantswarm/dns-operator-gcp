package registrar_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
	"google.golang.org/api/googleapi"

	"github.com/giantswarm/dns-operator-gcp/v2/tests"
)

const slowTestThreshold = 30 * time.Second

var (
	baseDomain    string
	parentDNSZone string
	gcpProject    string
)

func TestRegistrar(t *testing.T) {
	suiteConfig, reporterConfig := GinkgoConfiguration()
	reporterConfig.SlowSpecThreshold = slowTestThreshold

	RegisterFailHandler(Fail)
	RunSpecs(t, "Registrar Suite", suiteConfig, reporterConfig)
}

var _ = BeforeSuite(func() {
	tests.GetEnvOrSkip("GOOGLE_APPLICATION_CREDENTIALS")
	baseDomain = tests.GetEnvOrSkip("CLOUD_DNS_BASE_DOMAIN")
	parentDNSZone = tests.GetEnvOrSkip("CLOUD_DNS_PARENT_ZONE")
	gcpProject = tests.GetEnvOrSkip("GCP_PROJECT_ID")
})

type beGoogleAPIErrorWithStatusMatcher struct {
	expected int
}

func BeGoogleAPIErrorWithStatus(expected int) types.GomegaMatcher {
	return &beGoogleAPIErrorWithStatusMatcher{expected: expected}
}

func (m *beGoogleAPIErrorWithStatusMatcher) Match(actual interface{}) (bool, error) {
	if actual == nil {
		return false, nil
	}

	actualError, isError := actual.(error)
	if !isError {
		return false, fmt.Errorf("%#v is not an error", actual)
	}

	matches, err := BeAssignableToTypeOf(actualError).Match(&googleapi.Error{})
	if err != nil || !matches {
		return false, err
	}

	googleAPIError, isGoogleAPIError := actual.(*googleapi.Error)
	if !isGoogleAPIError {
		return false, fmt.Errorf("%#v is not a google api error", actual)
	}
	return Equal(googleAPIError.Code).Match(m.expected)
}

func (m *beGoogleAPIErrorWithStatusMatcher) FailureMessage(actual interface{}) (message string) {
	return format.Message(
		actual,
		fmt.Sprintf("to be a google api error with status code: %s", m.getExpectedStatusText()),
	)
}

func (m *beGoogleAPIErrorWithStatusMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(
		actual,
		fmt.Sprintf("to not be a google api error with status: %s", m.getExpectedStatusText()),
	)
}

func (m *beGoogleAPIErrorWithStatusMatcher) getExpectedStatusText() string {
	return fmt.Sprintf("%d %s", m.expected, http.StatusText(m.expected))
}
