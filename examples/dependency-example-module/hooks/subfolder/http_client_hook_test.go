package hookinfolder_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	subfolder "dependency-example-module/subfolder"

	"github.com/deckhouse/deckhouse/pkg/log"
	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/testing/mock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("http client hook example", func() {
	Context("refoncile func", func() {
		Context("when all services works correctly", func() {
			dc := mock.NewDependencyContainerMock(GinkgoT())
			dc.GetHTTPClientMock.Set(func(options ...pkg.HTTPOption) (h1 pkg.HTTPClient) {
				return mock.NewHTTPClientMock(GinkgoT()).DoMock.Set(func(req *http.Request) (rp1 *http.Response, err error) {
					Expect(req.Method).Should(Equal(http.MethodGet))
					Expect(req.URL.String()).Should(Equal("http://127.0.0.1"))

					return &http.Response{}, nil
				})
			})

			var input = &pkg.HookInput{
				DC:     dc,
				Logger: log.NewNop(),
			}

			It("reconcile func executed correctly", func() {
				err := subfolder.HandlerHTTPClient(context.Background(), input)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		Context("http client receive error", func() {
			dc := mock.NewDependencyContainerMock(GinkgoT())
			dc.GetHTTPClientMock.Set(func(options ...pkg.HTTPOption) (h1 pkg.HTTPClient) {
				return mock.NewHTTPClientMock(GinkgoT())
			})

			var input = &pkg.HookInput{
				DC:     dc,
				Logger: log.NewNop(),
			}

			It("error has occured", func() {
				err := subfolder.HandlerHTTPClient(nil, input)
				Expect(err).Should(HaveOccurred())
				Expect(err).Should(Equal(fmt.Errorf("new request: %w", errors.New("net/http: nil Context"))))
			})
		})

		Context("http client receive error", func() {
			dc := mock.NewDependencyContainerMock(GinkgoT())
			dc.GetHTTPClientMock.Set(func(options ...pkg.HTTPOption) (h1 pkg.HTTPClient) {
				return mock.NewHTTPClientMock(GinkgoT()).DoMock.Set(func(req *http.Request) (rp1 *http.Response, err error) {
					return &http.Response{}, errors.New("error")
				})
			})

			var input = &pkg.HookInput{
				DC:     dc,
				Logger: log.NewNop(),
			}

			It("error has occured", func() {
				err := subfolder.HandlerHTTPClient(context.Background(), input)
				Expect(err).Should(HaveOccurred())
				Expect(err).Should(Equal(fmt.Errorf("do request: %w", errors.New("error"))))
			})
		})
	})
})
