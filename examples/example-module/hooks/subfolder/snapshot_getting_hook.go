package hookinfolder

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/registry"
)

var _ = registry.RegisterFunc(configSnapshots, HandlerHookSnapshots)

type NodeInfo struct {
	APIVersion string           `json:"apiVersion"`
	Kind       string           `json:"kind"`
	Metadata   NodeInfoMetadata `json:"metadata"`
}

type NodeInfoMetadata struct {
	Name            string `json:"name"`
	ResourceVersion string `json:"resourceVersion"`
	UID             string `json:"uid"`
}

const applyNodeJQFilter = `{
	"apiVersion": .apiVersion,
	"kind": .kind,
	"metadata": {
		"name": .metadata.name,
		"resourceVersion": .metadata.name,
		"uid": .metadata.uid
	}
}`

const NodeInfoSnapshotName = "node_info"

var configSnapshots = &pkg.HookConfig{
	OnBeforeHelm: &pkg.OrderedConfig{Order: 1},
	Kubernetes: []pkg.KubernetesConfig{
		{
			Name:       NodeInfoSnapshotName,
			APIVersion: "v1",
			Kind:       "Node",
			JqFilter:   applyNodeJQFilter,
		},
	},
}

func HandlerHookSnapshots(_ context.Context, input *pkg.HookInput) error {
	input.Logger.Info("hello from snapshot hook")

	// getting info from snapshot
	// no info about key not found, if you need it - check length
	nodes := input.Snapshots.Get(NodeInfoSnapshotName)

	for _, o := range nodes {
		nodeInfo := new(NodeInfo)

		err := o.UnmarshalTo(nodeInfo)
		if err != nil {
			return fmt.Errorf("unmarshal to: %w", err)
		}

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
