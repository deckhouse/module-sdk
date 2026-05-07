package hookinfolder_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/testing/helpers"
	"github.com/deckhouse/module-sdk/testing/mock"

	subfolder "dependency-example-module/subfolder"
)

// httpDC builds a DependencyContainerMock whose GetHTTPClient returns a
// minimock-controlled HTTPClient with the supplied Do behaviour.
func httpDC(t *testing.T, doFn func(*http.Request) (*http.Response, error)) pkg.DependencyContainer {
	t.Helper()
	dc := mock.NewDependencyContainerMock(t)
	dc.GetHTTPClientMock.Set(func(_ ...pkg.HTTPOption) pkg.HTTPClient {
		c := mock.NewHTTPClientMock(t)
		if doFn != nil {
			c.DoMock.Set(doFn)
		}
		return c
	})
	return dc
}

func TestHandlerHTTPClient_HappyPath(t *testing.T) {
	dc := httpDC(t, func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, http.MethodGet, req.Method)
		assert.Equal(t, "http://127.0.0.1", req.URL.String())
		return &http.Response{}, nil
	})

	in := helpers.NewInputBuilder(t).WithDependencyContainer(dc).Build()
	require.NoError(t, subfolder.HandlerHTTPClient(context.Background(), in))
}

func TestHandlerHTTPClient_NilContextRejected(t *testing.T) {
	// We never expect Do to be called; httpDC builds a mock with no Do
	// expectation, so a stray call would fail the test automatically.
	dc := httpDC(t, nil)

	in := helpers.NewInputBuilder(t).WithDependencyContainer(dc).Build()

	//nolint:staticcheck // intentionally passing nil context to exercise net/http error
	err := subfolder.HandlerHTTPClient(nil, in)
	require.Error(t, err)
	assert.ErrorContains(t, err, "new request:")
}

func TestHandlerHTTPClient_ClientReturnsError(t *testing.T) {
	wantErr := errors.New("boom")
	dc := httpDC(t, func(_ *http.Request) (*http.Response, error) {
		return &http.Response{}, wantErr
	})

	in := helpers.NewInputBuilder(t).WithDependencyContainer(dc).Build()

	err := subfolder.HandlerHTTPClient(context.Background(), in)
	require.Error(t, err)
	assert.ErrorIs(t, err, wantErr)
	assert.ErrorContains(t, err, "do request:")
}
