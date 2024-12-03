package hookinfolder

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/deckhouse/module-sdk/pkg"
	objectpatch "github.com/deckhouse/module-sdk/pkg/object-patch"
	"github.com/deckhouse/module-sdk/pkg/registry"
)

var _ = registry.RegisterFunc(configSnapshotsAlt, handlerHookSnapshotsAlt)

var configSnapshotsAlt = &pkg.HookConfig{
	OnBeforeHelm: &pkg.OrderedConfig{Order: 1},
	Kubernetes: []pkg.KubernetesConfig{
		{
			Name:       nodeInfoSnapshotName,
			APIVersion: "v1",
			Kind:       "Node",
			JqFilter:   applyNodeJQFilter,
		},
	},
}

func handlerHookSnapshotsAlt(_ context.Context, input *pkg.HookInput) error {
	input.Logger.Info("hello from snapshot alt hook")

	// getting info from snapshot
	// no info about key not found, if you need it - check length
	nodeInfos, err := objectpatch.UnmarshalToStruct[NodeInfo](input.Snapshots, nodeInfoSnapshotName)
	if err != nil {
		return fmt.Errorf("unmarshal to struct: %w", err)
	}

	for _, nodeInfo := range nodeInfos {
		input.Logger.Info(
			"node found",
			slog.String("APIVersion", nodeInfo.APIVersion),
			slog.String("Kind", nodeInfo.Kind),
			slog.String("Name", nodeInfo.Metadata.Name),
			slog.String("ResourceVersion", nodeInfo.Metadata.ResourceVersion),
			slog.String("UID", nodeInfo.Metadata.UID),
		)
	}

	return nil
}
