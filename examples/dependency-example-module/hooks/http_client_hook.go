package hookinfolder

import (
	"context"
	"fmt"
	"net/http"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/registry"
)

var configHTTPCLient = &pkg.HookConfig{}

var _ = registry.RegisterFunc(configHTTPCLient, handlerHTTPClient)

func handlerHTTPClient(input *pkg.HookInput) error {
	httpClient := input.DC.GetHTTPClient()

	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, "http://127.0.0.1", nil)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	_, err = httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}

	return nil
}
