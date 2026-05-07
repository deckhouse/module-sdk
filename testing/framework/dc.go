package framework

import (
	"errors"
	"net/http"

	"github.com/jonboulle/clockwork"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	crfake "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/deckhouse/module-sdk/pkg"
)

// frameworkDC is a minimal pkg.DependencyContainer wired to the framework's
// fake clients.
//
// HTTP and registry clients return errors by default; tests that need them can
// override via SetHTTPClient / SetRegistryClient.
type frameworkDC struct {
	k8sClient *fakeKubeClient

	clock clockwork.Clock

	httpClient pkg.HTTPClient
	regClient  pkg.RegistryClient
}

func newFrameworkDC(dynamicClient dynamic.Interface, scheme *runtime.Scheme) *frameworkDC {
	ctrlClient := crfake.NewClientBuilder().WithScheme(scheme).Build()
	return &frameworkDC{
		k8sClient: &fakeKubeClient{
			Client:  ctrlClient,
			dynamic: dynamicClient,
		},
		clock: clockwork.NewFakeClock(),
	}
}

var _ pkg.DependencyContainer = (*frameworkDC)(nil)

// SetHTTPClient overrides the HTTP client returned by GetHTTPClient.
func (d *frameworkDC) SetHTTPClient(c pkg.HTTPClient) { d.httpClient = c }

// SetRegistryClient overrides the registry client returned by GetRegistryClient.
func (d *frameworkDC) SetRegistryClient(c pkg.RegistryClient) { d.regClient = c }

// SetClock overrides the framework's clock.
func (d *frameworkDC) SetClock(c clockwork.Clock) { d.clock = c }

// GetClock implements pkg.DependencyContainer.
func (d *frameworkDC) GetClock() clockwork.Clock { return d.clock }

// GetHTTPClient implements pkg.DependencyContainer.
func (d *frameworkDC) GetHTTPClient(_ ...pkg.HTTPOption) pkg.HTTPClient {
	if d.httpClient != nil {
		return d.httpClient
	}
	return errHTTPClient{}
}

// GetK8sClient implements pkg.DependencyContainer.
func (d *frameworkDC) GetK8sClient(_ ...pkg.KubernetesOption) (pkg.KubernetesClient, error) {
	return d.k8sClient, nil
}

// MustGetK8sClient implements pkg.DependencyContainer.
func (d *frameworkDC) MustGetK8sClient(_ ...pkg.KubernetesOption) pkg.KubernetesClient {
	return d.k8sClient
}

// GetClientConfig implements pkg.DependencyContainer.
func (d *frameworkDC) GetClientConfig() (*rest.Config, error) {
	return &rest.Config{Host: "fake-test"}, nil
}

// GetRegistryClient implements pkg.DependencyContainer.
func (d *frameworkDC) GetRegistryClient(_ string, _ ...pkg.RegistryOption) (pkg.RegistryClient, error) {
	if d.regClient != nil {
		return d.regClient, nil
	}
	return nil, errors.New("framework: registry client not configured (use SetRegistryClient)")
}

// MustGetRegistryClient implements pkg.DependencyContainer.
func (d *frameworkDC) MustGetRegistryClient(repo string, opts ...pkg.RegistryOption) pkg.RegistryClient {
	c, err := d.GetRegistryClient(repo, opts...)
	if err != nil {
		panic(err)
	}
	return c
}

// fakeKubeClient implements pkg.KubernetesClient by composing a controller-runtime
// fake client with the framework's dynamic fake client.
type fakeKubeClient struct {
	client.Client
	dynamic dynamic.Interface
}

// Dynamic implements pkg.KubernetesClient.
func (c *fakeKubeClient) Dynamic() dynamic.Interface { return c.dynamic }

// errHTTPClient returns an error on every request, for tests that don't override the HTTP client.
type errHTTPClient struct{}

func (errHTTPClient) Do(_ *http.Request) (*http.Response, error) {
	return nil, errors.New("framework: HTTP client not configured (use SetHTTPClient)")
}
