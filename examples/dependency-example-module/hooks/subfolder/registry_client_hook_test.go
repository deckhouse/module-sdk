package hookinfolder_test

import (
	"context"
	"errors"
	"testing"

	"github.com/gojuno/minimock/v3"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/testing/helpers"
	"github.com/deckhouse/module-sdk/testing/mock"

	subfolder "dependency-example-module/subfolder"
)

const (
	firstTag  = "v1.0.0"
	secondTag = "v2.0.0"
)

// registryDC wires a DependencyContainerMock whose MustGetRegistryClient
// returns the provided RegistryClientMock for the configured registry
// address. The caller fully owns the mock (set up tags, image, digest, …).
func registryDC(t *testing.T, regClient *mock.RegistryClientMock) pkg.DependencyContainer {
	t.Helper()
	dc := mock.NewDependencyContainerMock(t)
	dc.MustGetRegistryClientMock.When(subfolder.RegistryAddress).Then(regClient)
	return dc
}

func TestHandlerRegistryClient_AllOK(t *testing.T) {
	rc := mock.NewRegistryClientMock(t)
	rc.ListTagsMock.Return([]string{firstTag, secondTag}, nil)

	rc.ImageMock.When(minimock.AnyContext, firstTag).
		Then(mock.NewRegistryImageMock(t).ConfigNameMock.Expect().
			Return(v1.Hash{Algorithm: "sha256", Hex: "abcdef1"}, nil), nil)
	rc.ImageMock.When(minimock.AnyContext, secondTag).
		Then(mock.NewRegistryImageMock(t).ConfigNameMock.Expect().
			Return(v1.Hash{Algorithm: "sha256", Hex: "abcdef2"}, nil), nil)

	rc.DigestMock.When(minimock.AnyContext, firstTag).Then("first digest", nil)
	rc.DigestMock.When(minimock.AnyContext, secondTag).Then("second digest", nil)

	in := helpers.NewInputBuilder(t).
		WithDependencyContainer(registryDC(t, rc)).
		Build()

	require.NoError(t, subfolder.HandlerRegistryClient(context.Background(), in))
}

func TestHandlerRegistryClient_NoTagsOK(t *testing.T) {
	rc := mock.NewRegistryClientMock(t)
	rc.ListTagsMock.Return([]string{}, nil)

	in := helpers.NewInputBuilder(t).
		WithDependencyContainer(registryDC(t, rc)).
		Build()

	require.NoError(t, subfolder.HandlerRegistryClient(context.Background(), in))
}

func TestHandlerRegistryClient_FailureCases(t *testing.T) {
	cases := []struct {
		name      string
		setup     func(*mock.RegistryClientMock)
		wantInMsg string
	}{
		{
			name: "list tags error",
			setup: func(rc *mock.RegistryClientMock) {
				rc.ListTagsMock.Return(nil, errors.New("boom"))
			},
			wantInMsg: "list tags:",
		},
		{
			name: "image error",
			setup: func(rc *mock.RegistryClientMock) {
				rc.ListTagsMock.Return([]string{firstTag, secondTag}, nil)
				rc.ImageMock.When(minimock.AnyContext, firstTag).Then(nil, errors.New("boom"))
			},
			wantInMsg: "image:",
		},
		{
			name: "config name error",
			setup: func(rc *mock.RegistryClientMock) {
				rc.ListTagsMock.Return([]string{firstTag, secondTag}, nil)
				rc.ImageMock.When(minimock.AnyContext, firstTag).
					Then(mock.NewRegistryImageMock(t).ConfigNameMock.Expect().
						Return(v1.Hash{}, errors.New("boom")), nil)
			},
			wantInMsg: "config name:",
		},
		{
			name: "digest error",
			setup: func(rc *mock.RegistryClientMock) {
				rc.ListTagsMock.Return([]string{firstTag, secondTag}, nil)
				rc.ImageMock.When(minimock.AnyContext, firstTag).
					Then(mock.NewRegistryImageMock(t).ConfigNameMock.Expect().
						Return(v1.Hash{Algorithm: "sha256", Hex: "abcdef1"}, nil), nil)
				rc.DigestMock.When(minimock.AnyContext, firstTag).Then("", errors.New("boom"))
			},
			wantInMsg: "digest:",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rc := mock.NewRegistryClientMock(t)
			tc.setup(rc)

			in := helpers.NewInputBuilder(t).
				WithDependencyContainer(registryDC(t, rc)).
				Build()

			err := subfolder.HandlerRegistryClient(context.Background(), in)
			require.Error(t, err)
			assert.ErrorContains(t, err, tc.wantInMsg)
		})
	}
}
