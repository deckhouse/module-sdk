package hookinfolder

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/registry"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = registry.RegisterFunc(config, handlerCRD)

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

const nodeInfoSnapshotName = "node_info"

var config = &pkg.HookConfig{
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

func handlerCRD(input *pkg.HookInput) error {
	input.Logger.Info("hello from first root hook")

	// getting info from snapshot
	// no info about key not found, if you need it - check length
	nodes := input.Snapshots.Get(nodeInfoSnapshotName)

	for _, o := range nodes {
		nodeInfo := new(NodeInfo)

		err := o.UnmarhalTo(nodeInfo)
		if err != nil {
			return fmt.Errorf("unmarshal to: %w", err)
		}

		input.Logger.Info("unmarshal hook node", slog.Any("node", nodeInfo))
	}

	// using kubernetes client from dependency container
	k8sClient := input.DC.MustGetK8sClient()

	const (
		podNamespace = "test-namespace"
		podName      = "test-pod"
	)

	pod := new(corev1.Pod)
	err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: podNamespace, Name: podName}, pod)
	if err != nil {
		return fmt.Errorf("get pod: %w", err)
	}

	input.Logger.Info("pod", slog.String("name", pod.GetName()), slog.String("namespace", pod.GetNamespace()))

	// using http client from dependency container
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
