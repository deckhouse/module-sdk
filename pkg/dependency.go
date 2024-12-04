package pkg

import (
	"context"
	"net/http"
	"time"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/jonboulle/clockwork"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Container with external dependencies
// Avoid using dependencies, if you can, because of it cost
type DependencyContainer interface {
	GetHTTPClient(options ...HTTPOption) HTTPClient

	GetK8sClient(options ...KubernetesOption) (KubernetesClient, error)
	MustGetK8sClient(options ...KubernetesOption) KubernetesClient
	GetClientConfig() (*rest.Config, error)

	GetRegistryClient(repo string, options ...RegistryOption) (RegistryClient, error)
	MustGetRegistryClient(repo string, options ...RegistryOption) RegistryClient

	GetClock() clockwork.Clock
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type HTTPOption interface {
	Apply(optsApplier HTTPOptionApplier)
}

type HTTPOptionApplier interface {
	WithTimeout(t time.Duration)
	WithInsecureSkipVerify()
	WithAdditionalCACerts(certs [][]byte)
	WithTLSServerName(name string)
}

type RegistryClient interface {
	Image(ctx context.Context, tag string) (v1.Image, error)
	Digest(ctx context.Context, tag string) (string, error)
	ListTags(ctx context.Context) ([]string, error)
}

type RegistryOption interface {
	Apply(optsApplier RegistryOptionApplier)
}

type RegistryOptionApplier interface {
	WithCA(ca string)
	WithInsecureSchema(insecure bool)
	WithAuth(dockerCfg string)
	WithUserAgent(ua string)
	WithTimeout(timeout time.Duration)
}

type KubernetesClient interface {
	client.Client
	Dynamic() dynamic.Interface
}

type KubernetesOption interface {
	Apply(optsApplier KubernetesOptionApplier)
}

type KubernetesOptionApplier interface {
	WithSchemeBuilder(builder runtime.SchemeBuilder)
}
