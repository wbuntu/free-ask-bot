package e2e

import (
	"testing"

	"gitbub.com/wbuntu/gin-template/internal/pkg/tools"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestE2e(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "E2E Suite")
}

var (
	tenantID = "1234567890"
	baseURL  = "http://127.0.0.1:8080"
)

var httpClient *tools.HTTPClient

var _ = BeforeSuite(func() {
	httpClient = tools.NewHTTPClient(baseURL, map[string]string{"X-Tenant-ID": tenantID})
})

var _ = AfterSuite(func() {
})
