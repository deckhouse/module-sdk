package hookinfolder_test

import (
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/testing/framework"
	"github.com/deckhouse/module-sdk/testing/mock"

	subfolder "dependency-example-module/subfolder"
)

// TestHandlerHTTPClient_FrameworkLevel exercises the HTTP-client hook
// through testing/framework. The framework's DependencyContainer can be
// reconfigured before RunHook to inject a custom HTTPClient, which is
// what we do here.
func TestHandlerHTTPClient_FrameworkLevel(t *testing.T) {
	calls := atomic.Int32{}

	httpClient := mock.NewHTTPClientMock(t)
	httpClient.DoMock.Set(func(req *http.Request) (*http.Response, error) {
		calls.Add(1)
		assert.Equal(t, "http://127.0.0.1", req.URL.String())
		return &http.Response{}, nil
	})

	f := framework.HookExecutionConfigInit(t,
		&pkg.HookConfig{},
		subfolder.HandlerHTTPClient,
		`{}`, `{}`,
	)

	// Override the default error-returning HTTP client with our mock.
	f.DependencyContainer().SetHTTPClient(httpClient)

	f.RunHook()

	require.NoError(t, f.HookError())
	assert.Equal(t, int32(1), calls.Load(), "expected the hook to issue exactly one HTTP request")
}
