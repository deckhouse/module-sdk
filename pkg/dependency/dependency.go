/*
Copyright 2021 Flant JSC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package dependency

import (
	stdhttp "net/http"
	"os"
	"sync"
	"time"

	"github.com/jonboulle/clockwork"
	"github.com/pkg/errors"
	"k8s.io/client-go/rest"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/dependency/cr"
	"github.com/deckhouse/module-sdk/pkg/dependency/http"
	"github.com/deckhouse/module-sdk/pkg/dependency/k8s"
)

var (
	TestTimeZone = time.UTC
)

// NewDependencyContainer creates new Dependency container with external clients
func NewDependencyContainer() pkg.DependencyContainer {
	return &dependencyContainer{}
}

type dependencyContainer struct {
	k8sClient *k8s.Client

	httpmu     sync.RWMutex
	httpClient *stdhttp.Client

	crmu     sync.RWMutex
	crClient *cr.Client
}

func (dc *dependencyContainer) GetHTTPClient(options ...pkg.HTTPOption) pkg.HTTPClient {
	dc.httpmu.RLock()
	if dc.httpClient != nil {
		defer dc.httpmu.RUnlock()
		return dc.httpClient
	}
	dc.httpmu.RUnlock()

	dc.httpmu.Lock()
	defer dc.httpmu.Unlock()

	var opts []pkg.HTTPOption
	opts = append(opts, options...)

	contentCA, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/ca.crt")
	if err == nil {
		opts = append(opts, http.WithAdditionalCACerts([][]byte{contentCA}))
	}

	dc.httpClient = http.NewClient(opts...)

	return dc.httpClient
}

func (dc *dependencyContainer) GetK8sClient(options ...pkg.KubernetesOption) (pkg.KubernetesClient, error) {
	if dc.k8sClient == nil {
		kc, err := k8s.NewClient(options...)
		if err != nil {
			return nil, err
		}

		dc.k8sClient = kc
	}

	return dc.k8sClient, nil
}

func (dc *dependencyContainer) MustGetK8sClient(options ...pkg.KubernetesOption) pkg.KubernetesClient {
	client, err := dc.GetK8sClient(options...)
	if err != nil {
		panic(err)
	}

	return client
}

func (dc *dependencyContainer) GetRegistryClient(repo string, options ...pkg.RegistryOption) (pkg.RegistryClient, error) {
	dc.crmu.RLock()
	if dc.crClient != nil {
		defer dc.crmu.RUnlock()
		return dc.crClient, nil
	}
	dc.crmu.RUnlock()

	dc.crmu.Lock()
	defer dc.crmu.Unlock()

	var err error
	dc.crClient, err = cr.NewClient(repo, options...)
	if err != nil {
		return nil, err
	}

	return dc.crClient, nil
}

func (dc *dependencyContainer) MustGetRegistryClient(repo string, options ...pkg.RegistryOption) pkg.RegistryClient {
	client, err := dc.GetRegistryClient(repo, options...)
	if err != nil {
		panic(err)
	}

	return client
}

func (dc *dependencyContainer) GetClientConfig() (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	caCert, err := os.ReadFile(config.TLSClientConfig.CAFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read CA file")
	}

	config.CAData = caCert

	return config, nil
}

func (dc *dependencyContainer) GetClock() clockwork.Clock {
	return clockwork.NewRealClock()
}
