/*
Copyright 2025 Flant JSC

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

package storageclasschange

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/deckhouse/deckhouse/pkg/log"
	"github.com/stoewer/go-strcase"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/deckhouse/module-sdk/pkg"
	objectpatch "github.com/deckhouse/module-sdk/pkg/object-patch"
	"github.com/deckhouse/module-sdk/pkg/registry"
)

type Args struct {
	ModuleName                    string `json:"moduleName"`
	Namespace                     string `json:"namespace"`
	LabelSelectorKey              string `json:"labelSelectorKey"`
	LabelSelectorValue            string `json:"labelSelectorValue"`
	ObjectKind                    string `json:"objectKind"`
	ObjectName                    string `json:"objectName"`
	InternalValuesSubPath         string `json:"internalValuesSubPath,omitempty"`
	D8ConfigStorageClassParamName string `json:"d8ConfigStorageClassParamName,omitempty"`

	// if return value is false - hook will stop its execution
	BeforeHookCheck func(input *pkg.HookInput) bool
}

func RegisterHook(args Args) bool {
	return registry.RegisterFunc(&pkg.HookConfig{
		OnBeforeHelm: &pkg.OrderedConfig{Order: 1},
		Kubernetes: []pkg.KubernetesConfig{
			{
				Name:       "pvcs",
				APIVersion: "v1",
				Kind:       "PersistentVolumeClaim",
				NamespaceSelector: &pkg.NamespaceSelector{
					NameSelector: &pkg.NameSelector{
						MatchNames: []string{args.Namespace},
					},
				},
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						args.LabelSelectorKey: args.LabelSelectorValue,
					},
				},
				JqFilter: pvcFilter,
			},
			{
				Name:       "pods",
				APIVersion: "v1",
				Kind:       "Pod",
				NamespaceSelector: &pkg.NamespaceSelector{
					NameSelector: &pkg.NameSelector{
						MatchNames: []string{args.Namespace},
					},
				},
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						args.LabelSelectorKey: args.LabelSelectorValue,
					},
				},
				JqFilter: podFilter,
			},
			{
				Name:       "storageclasses",
				APIVersion: "storage.k8s.io/v1",
				Kind:       "Storageclass",
				JqFilter:   storageClassFilter,
			},
		},
	}, func(ctx context.Context, input *pkg.HookInput) error {
		return storageClassChange(ctx, input, args)
	})
}

type filteredPvc struct {
	Name             string `json:"name"`
	Namespace        string `json:"namespace"`
	StorageClassName string `json:"storageClassName"`
	IsDeleted        bool   `json:"isDeleted"`
}

var pvcFilter = `{
	"name": .metadata.name,
	"namespace": .metadata.namespace,
    "storageClassName": .spec.storageClassName,
    "isDeleted": .metadata.deletionTimestamp != null,
}`

type filteredStorageClass struct {
	Name      string `json:"name"`
	IsDefault bool   `json:"isDefault"`
}

var storageClassFilter = `{
	"name": .metadata.name,
    "isDefault": (
    (.metadata.annotations["storageclass.kubernetes.io/is-default-class"] == "true")
    or
    (.metadata.annotations["storageclass.beta.kubernetes.io/is-default-class"] == "true")
    )
}`

type filteredPod struct {
	Name      string          `json:"name"`
	Namespace string          `json:"namespace"`
	Pvc       string          `json:"pvc"`
	Phase     corev1.PodPhase `json:"phase"`
}

var podFilter = `{
	"name": .metadata.name,
	"namespace": .metadata.namespace,
	"pvc": (
    	.spec.volumes[]? 
    	| select(.persistentVolumeClaim != null) 
    	| .persistentVolumeClaim.claimName
	),
	"phase": .status.phase,
}`

func storageClassChange(ctx context.Context, input *pkg.HookInput, args Args) error {
	kubeClient, err := input.DC.GetK8sClient()
	if err != nil {
		return fmt.Errorf("get k8s client: %w", err)
	}

	pvcs, err := objectpatch.UnmarshalToStruct[filteredPvc](input.Snapshots, "pvcs")
	if err != nil {
		return fmt.Errorf("unmarshal pvcs snapshot: %w", err)
	}

	pods, err := objectpatch.UnmarshalToStruct[filteredPod](input.Snapshots, "pods")
	if err != nil {
		return fmt.Errorf("unmarshal pods snapshot: %w", err)
	}

	findPodByPVCName := func(pvcName string) (filteredPod, error) {
		for _, pod := range pods {
			if pod.Pvc == pvcName {
				return pod, nil
			}
		}

		return filteredPod{}, fmt.Errorf("pod with volume name [%s] not found", pvcName)
	}

	var existingPvcs []filteredPvc
	for _, pvc := range pvcs {
		if !pvc.IsDeleted {
			existingPvcs = append(existingPvcs, pvc)
			continue
		}

		pod, err := findPodByPVCName(pvc.Name)
		if err != nil {
			input.Logger.Warn("find pod by fpvc name failed", slog.String("pvc", pvc.Name), log.Err(err))
			continue
		}

		// if someone deleted pvc then evict the pod.
		evict := &policyv1.Eviction{
			ObjectMeta: metav1.ObjectMeta{
				Name: pod.Name,
			},
		}

		input.Logger.Info("evict pod due",
			slog.String("namespace", pod.Namespace),
			slog.String("pod", pod.Name),
			slog.String("pvc", pvc.Name))

		if err = kubeClient.Create(ctx, evict); err != nil {
			input.Logger.Warn("failed to evict pod",
				slog.String("namespace", pod.Namespace),
				slog.String("pod", pod.Name), log.Err(err))
		}
	}

	var currentStorageClass string
	if len(existingPvcs) > 0 {
		currentStorageClass = existingPvcs[0].StorageClassName
	}

	effectiveStorageClass, err := calculateEffectiveStorageClass(input, args, currentStorageClass)
	if err != nil {
		return err
	}

	if !storageClassesAreEqual(currentStorageClass, effectiveStorageClass) {
		if !isEmptyOrFalseStr(currentStorageClass) {
			for _, pvc := range existingPvcs {
				input.Logger.Info("pvc storage class changed, delete pvc", slog.String("namespace", pvc.Namespace), slog.String("pvc", pvc.Name))
				obj := &corev1.PersistentVolumeClaim{
					ObjectMeta: metav1.ObjectMeta{
						Name:      pvc.Name,
						Namespace: pvc.Namespace,
					},
				}
				if err = kubeClient.Delete(ctx, obj); err != nil {
					input.Logger.Error("failed to delete pvc", log.Err(err))
				}
			}
		}

		input.Logger.Info("storageClass changed. Deleting objects",
			slog.String("namespace", args.Namespace),
			slog.String("object_kind", args.ObjectKind),
			slog.String("name", args.ObjectName))

		switch args.ObjectKind {
		case "Prometheus":
			err = kubeClient.Dynamic().Resource(schema.GroupVersionResource{
				Group: "monitoring.coreos.com", Version: "v1", Resource: "prometheuses.monitoring.coreos.com",
			}).Namespace(args.Namespace).Delete(ctx, args.ObjectName, metav1.DeleteOptions{})
		case "StatefulSet":
			obj := &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      args.ObjectName,
					Namespace: args.Namespace,
				},
			}
			err = kubeClient.Delete(ctx, obj)
		default:
			return fmt.Errorf("unknown object kind %s", args.ObjectKind)
		}

		if err != nil && !errors.IsNotFound(err) {
			input.Logger.Error(err.Error())
		}
	}

	return nil
}

// effective storage class is the target storage class. If it changes, the PVC will be recreated.
func calculateEffectiveStorageClass(input *pkg.HookInput, args Args, currentStorageClass string) (string, error) {
	if args.BeforeHookCheck != nil && !args.BeforeHookCheck(input) {
		return "", nil
	}

	var effectiveStorageClass string

	storageClasses, err := objectpatch.UnmarshalToStruct[filteredStorageClass](input.Snapshots, "storageClasses")
	if err != nil {
		return "", fmt.Errorf("unmarshal storageClasses snapshot: %w", err)
	}

	for _, sc := range storageClasses {
		if sc.IsDefault {
			effectiveStorageClass = sc.Name
			break
		}
	}

	if input.ConfigValues.Exists("global.modules.storageClass") {
		effectiveStorageClass = input.ConfigValues.Get("global.modules.storageClass").String()
	}

	// storage class from pvc
	if currentStorageClass != "" {
		effectiveStorageClass = currentStorageClass
	}

	var configValuesPath = fmt.Sprintf("%s.storageClass", args.ModuleName)

	if args.D8ConfigStorageClassParamName != "" {
		configValuesPath = fmt.Sprintf("%s.%s", args.ModuleName, args.D8ConfigStorageClassParamName)
	}

	if input.ConfigValues.Exists(configValuesPath) {
		effectiveStorageClass = input.ConfigValues.Get(configValuesPath).String()
	}

	var internalValuesPath = fmt.Sprintf("%s.internal.effectiveStorageClass", strcase.LowerCamelCase(args.ModuleName))

	if args.InternalValuesSubPath != "" {
		internalValuesPath = fmt.Sprintf("%s.internal.%s.effectiveStorageClass", strcase.LowerCamelCase(args.ModuleName), args.InternalValuesSubPath)
	}

	emptydirUsageMetricValue := 0.0
	if len(effectiveStorageClass) == 0 || effectiveStorageClass == "false" {
		input.Values.Set(internalValuesPath, false)
		emptydirUsageMetricValue = 1.0
	} else {
		input.Values.Set(internalValuesPath, effectiveStorageClass)
	}

	input.MetricsCollector.Set(
		"d8_emptydir_usage",
		emptydirUsageMetricValue,
		map[string]string{
			"namespace":   args.Namespace,
			"module_name": args.ModuleName,
		},
	)

	return effectiveStorageClass, nil
}

func storageClassesAreEqual(sc1, sc2 string) bool {
	if sc1 == sc2 {
		return true
	}
	return isEmptyOrFalseStr(sc1) && isEmptyOrFalseStr(sc2)
}

// isEmptyOrFalseStr returns true if sc is empty string or "false". For storage class values or
// configuration, empty strings and "false" mean the same: no storage class specified. "false" is
// set by humans, while absent values resolve to empty strings.
func isEmptyOrFalseStr(sc string) bool {
	return sc == "" || sc == "false"
}
