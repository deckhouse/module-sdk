package hookinfolder

import (
	"context"

	"github.com/deckhouse/module-sdk/pkg"
	objectpatch "github.com/deckhouse/module-sdk/pkg/object-patch"
	"github.com/deckhouse/module-sdk/pkg/registry"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = registry.RegisterFunc(configPatch, HandlerHookPatch)

var configPatch = &pkg.HookConfig{
	OnBeforeHelm: &pkg.OrderedConfig{Order: 1},
}

func HandlerHookPatch(ctx context.Context, input *pkg.HookInput) error {
	input.Logger.Info("hello from patch hook")

	// CREATE
	firstPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-first-pod",
			Namespace: "default",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}
	input.PatchCollector.Create(firstPod, objectpatch.WithSubresource("/status"))

	secondPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-second-pod",
			Namespace: "default",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}
	input.PatchCollector.CreateOrUpdate(secondPod, objectpatch.WithSubresource("/status"))

	thirdPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-third-pod",
			Namespace: "default",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}
	input.PatchCollector.CreateIfNotExists(thirdPod, objectpatch.WithSubresource("/status"))

	// DELETE
	input.PatchCollector.Delete(
		firstPod.APIVersion,
		firstPod.Kind,
		firstPod.Namespace,
		firstPod.Name,
		objectpatch.WithSubresource("/status"),
	)

	input.PatchCollector.DeleteInBackground(
		secondPod.APIVersion,
		secondPod.Kind,
		secondPod.Namespace,
		secondPod.Name,
		objectpatch.WithSubresource("/status"),
	)

	input.PatchCollector.DeleteNonCascading(
		thirdPod.APIVersion,
		thirdPod.Kind,
		thirdPod.Namespace,
		thirdPod.Name,
		objectpatch.WithSubresource("/status"),
	)

	// PATCH
	statusPatch := map[string]interface{}{
		"/status": "newStatus",
	}

	input.PatchCollector.MergePatch(
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
