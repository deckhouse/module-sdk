package hookinfolder

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/registry"
)

const (
	RegistryAddress = "registry.address.com/"
)

var configRegistryCLient = &pkg.HookConfig{}

var _ = registry.RegisterFunc(configRegistryCLient, HandlerRegistryClient)

func HandlerRegistryClient(ctx context.Context, input *pkg.HookInput) error {
	registryClient := input.DC.MustGetRegistryClient(RegistryAddress)

	tags, err := registryClient.ListTags(ctx)
	if err != nil {
		return fmt.Errorf("list tags: %w", err)
	}

	if len(tags) == 0 {
		input.Logger.Warn("tags not found")
		return nil
	}

	input.Logger.Info("found some tags", slog.Any("tags", tags))

	for _, tag := range tags {
		img, err := registryClient.Image(ctx, tag)
		if err != nil {
			return fmt.Errorf("image: %w", err)
		}

		hash, err := img.ConfigName()
		if err != nil {
			return fmt.Errorf("config name: %w", err)
		}

		input.Logger.Info("image found", slog.String("config_name", hash.String()))

		dig, err := registryClient.Digest(ctx, tag)
		if err != nil {
			return fmt.Errorf("digest: %w", err)
		}

		input.Logger.Info("digest found", slog.String("digest", dig))
	}

	return nil
}
