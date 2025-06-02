package hookinfolder

import (
	"context"
	"fmt"
	"net/http"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/registry"
)

var configHTTPCLient = &pkg.HookConfig{}

var _ = registry.RegisterFunc(configHTTPCLient, HandlerHTTPClient)

func HandlerHTTPClient(ctx context.Context, input *pkg.HookInput) error {
	httpClient := input.DC.GetHTTPClient()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://127.0.0.1", nil)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer func() {
		if resp.Body != nil {
			resp.Body.Close()
		}
	}()

	return nil
}
