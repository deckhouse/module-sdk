package hookinfolder

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/deckhouse/module-sdk/pkg"
	objectpatch "github.com/deckhouse/module-sdk/pkg/object-patch"
	"github.com/deckhouse/module-sdk/pkg/registry"
)

var _ = registry.RegisterFunc(configPatch, HandlerHookPatch)

var configPatch = &pkg.HookConfig{
	OnBeforeHelm: &pkg.OrderedConfig{Order: 1},
}

func HandlerHookPatch(_ context.Context, input *pkg.HookInput) error {
	input.Logger.Info("hello from patch hook")

	// CREATE
	firstPod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-first-pod",
			Namespace: "default",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}
	input.PatchCollector.Create(firstPod)

	secondPod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-second-pod",
			Namespace: "default",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}
	input.PatchCollector.CreateOrUpdate(secondPod)

	thirdPod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-third-pod",
			Namespace: "default",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}
	input.PatchCollector.CreateIfNotExists(thirdPod)

	// DELETE
	input.PatchCollector.Delete(
		firstPod.APIVersion,
		firstPod.Kind,
		firstPod.Namespace,
		firstPod.Name,
	)

	input.PatchCollector.DeleteInBackground(
		secondPod.APIVersion,
		secondPod.Kind,
		secondPod.Namespace,
		secondPod.Name,
	)

	input.PatchCollector.DeleteNonCascading(
		thirdPod.APIVersion,
		thirdPod.Kind,
		thirdPod.Namespace,
		thirdPod.Name,
	)

	// PATCH
	statusPatch := map[string]any{
		"/status": "newStatus",
	}

	input.PatchCollector.PatchWithMerge(
		statusPatch,
		thirdPod.APIVersion,
		thirdPod.Kind,
		thirdPod.Namespace,
		thirdPod.Name,
		objectpatch.WithSubresource("/status"),
		objectpatch.WithIgnoreMissingObject(true),
	)

	return nil
}
