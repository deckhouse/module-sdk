package hookinfolder_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/gojuno/minimock/v3"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/deckhouse/deckhouse/pkg/log"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/testing/mock"

	subfolder "dependency-example-module/subfolder"
)

const (
	firstTag  = "v1.0.0"
	secondTag = "v2.0.0"
)

var _ = Describe("registry client hook example", func() {
	Context("refoncile func", func() {
		When("all services works correctly", func() {
			dc := mock.NewDependencyContainerMock(GinkgoT())

			regClient := mock.NewRegistryClientMock(GinkgoT())
			regClient.ListTagsMock.Set(func(_ context.Context) ([]string, error) {
				return []string{
					firstTag, secondTag,
				}, nil
			})

			regClient.ImageMock.When(minimock.AnyContext, firstTag).
				Then(mock.NewRegistryImageMock(GinkgoT()).ConfigNameMock.Expect().
					Return(v1.Hash{Algorithm: "sha256", Hex: "abcdef1"}, nil), nil)
			regClient.DigestMock.When(minimock.AnyContext, firstTag).
				Then("first digest", nil)

			regClient.ImageMock.When(minimock.AnyContext, secondTag).
				Then(mock.NewRegistryImageMock(GinkgoT()).ConfigNameMock.Expect().
					Return(v1.Hash{Algorithm: "sha256", Hex: "abcdef2"}, nil), nil)
			regClient.DigestMock.When(minimock.AnyContext, secondTag).
				Then("second digest", nil)

			dc.MustGetRegistryClientMock.When(subfolder.RegistryAddress).
				Then(regClient)

			var input = &pkg.HookInput{
				DC:     dc,
				Logger: log.NewNop(),
			}

			It("reconcile func executed correctly", func() {
				err := subfolder.HandlerRegistryClient(context.Background(), input)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		When("no tags listed", func() {
			dc := mock.NewDependencyContainerMock(GinkgoT())

			regClient := mock.NewRegistryClientMock(GinkgoT())
			regClient.ListTagsMock.Set(func(_ context.Context) ([]string, error) {
				return []string{}, nil
			})

			dc.MustGetRegistryClientMock.When(subfolder.RegistryAddress).
				Then(regClient)

			var input = &pkg.HookInput{
				DC:     dc,
				Logger: log.NewNop(),
			}

			It("reconcile func executed correctly", func() {
				err := subfolder.HandlerRegistryClient(context.Background(), input)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})

		When("list tags error", func() {
			dc := mock.NewDependencyContainerMock(GinkgoT())

			regClient := mock.NewRegistryClientMock(GinkgoT())
			regClient.ListTagsMock.Set(func(_ context.Context) ([]string, error) {
				return nil, errors.New("error")
			})

			dc.MustGetRegistryClientMock.When(subfolder.RegistryAddress).
				Then(regClient)

			var input = &pkg.HookInput{
				DC:     dc,
				Logger: log.NewNop(),
			}

			It("error has occurred", func() {
				err := subfolder.HandlerRegistryClient(context.Background(), input)
				Expect(err).Should(HaveOccurred())
				Expect(err).Should(Equal(fmt.Errorf("list tags: %w", errors.New("error"))))
			})
		})

		When("getting image errror", func() {
			dc := mock.NewDependencyContainerMock(GinkgoT())

			regClient := mock.NewRegistryClientMock(GinkgoT())
			regClient.ListTagsMock.Set(func(_ context.Context) ([]string, error) {
				return []string{
					firstTag, secondTag,
				}, nil
			})

			regClient.ImageMock.When(minimock.AnyContext, firstTag).
				Then(nil, errors.New("error"))

			dc.MustGetRegistryClientMock.When(subfolder.RegistryAddress).
				Then(regClient)

			var input = &pkg.HookInput{
				DC:     dc,
				Logger: log.NewNop(),
			}

			It("error has occurred", func() {
				err := subfolder.HandlerRegistryClient(context.Background(), input)
				Expect(err).Should(HaveOccurred())
				Expect(err).Should(Equal(fmt.Errorf("image: %w", errors.New("error"))))
			})
		})

		When("config name error", func() {
			dc := mock.NewDependencyContainerMock(GinkgoT())

			regClient := mock.NewRegistryClientMock(GinkgoT())
			regClient.ListTagsMock.Set(func(_ context.Context) ([]string, error) {
				return []string{
					firstTag, secondTag,
				}, nil
			})

			regClient.ImageMock.When(minimock.AnyContext, firstTag).
				Then(mock.NewRegistryImageMock(GinkgoT()).ConfigNameMock.Expect().
					Return(v1.Hash{}, errors.New("error")), nil)

			dc.MustGetRegistryClientMock.When(subfolder.RegistryAddress).
				Then(regClient)

			var input = &pkg.HookInput{
				DC:     dc,
				Logger: log.NewNop(),
			}

			It("error has occurred", func() {
				err := subfolder.HandlerRegistryClient(context.Background(), input)
				Expect(err).Should(HaveOccurred())
				Expect(err).Should(Equal(fmt.Errorf("config name: %w", errors.New("error"))))
			})
		})

		When("get digest error", func() {
			dc := mock.NewDependencyContainerMock(GinkgoT())

			regClient := mock.NewRegistryClientMock(GinkgoT())
			regClient.ListTagsMock.Set(func(_ context.Context) ([]string, error) {
				return []string{
					firstTag, secondTag,
				}, nil
			})

			regClient.ImageMock.When(minimock.AnyContext, firstTag).
				Then(mock.NewRegistryImageMock(GinkgoT()).ConfigNameMock.Expect().
					Return(v1.Hash{Algorithm: "sha256", Hex: "abcdef1"}, nil), nil)
			regClient.DigestMock.When(minimock.AnyContext, firstTag).
				Then("", errors.New("error"))

			dc.MustGetRegistryClientMock.When(subfolder.RegistryAddress).
				Then(regClient)

			var input = &pkg.HookInput{
				DC:     dc,
				Logger: log.NewNop(),
			}

			It("error has occurred", func() {
				err := subfolder.HandlerRegistryClient(context.Background(), input)
				Expect(err).Should(HaveOccurred())
				Expect(err).Should(Equal(fmt.Errorf("digest: %w", errors.New("error"))))
			})
		})
	})
})
