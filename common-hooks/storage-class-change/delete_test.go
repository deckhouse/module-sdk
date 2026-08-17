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
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	crfake "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/deckhouse/deckhouse/pkg/log"

	"github.com/deckhouse/module-sdk/pkg"
)

// prometheusGVRWant is spelled out on purpose: the dynamic fake answers only on
// the resource it was registered with, so a wrong plural in prometheusGVR fails
// these tests instead of silently 404-ing like it did against a real cluster.
var prometheusGVRWant = schema.GroupVersionResource{
	Group: "monitoring.coreos.com", Version: "v1", Resource: "prometheuses",
}

// fakeKubeClient satisfies pkg.KubernetesClient with the controller-runtime and
// dynamic fake clients. The framework's own fake (testing/framework/dc.go) starts
// with an empty controller-runtime tracker, so deletions are covered here instead.
type fakeKubeClient struct {
	client.Client
	dyn dynamic.Interface
}

func (c *fakeKubeClient) Dynamic() dynamic.Interface { return c.dyn }

func newKubeClient(prometheuses []*unstructured.Unstructured, objs ...client.Object) *fakeKubeClient {
	prometheusObjects := make([]runtime.Object, 0, len(prometheuses))
	for _, p := range prometheuses {
		prometheusObjects = append(prometheusObjects, p)
	}

	return &fakeKubeClient{
		Client: crfake.NewClientBuilder().WithScheme(clientgoscheme.Scheme).WithObjects(objs...).Build(),
		dyn: dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
			runtime.NewScheme(),
			map[schema.GroupVersionResource]string{prometheusGVRWant: "PrometheusList"},
			prometheusObjects...,
		),
	}
}

func statefulSet(name string, labels map[string]string) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "test-ns", Labels: labels},
	}
}

func prometheus(name, label string) *unstructured.Unstructured {
	return &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "monitoring.coreos.com/v1",
		"kind":       "Prometheus",
		"metadata": map[string]any{
			"name":      name,
			"namespace": "test-ns",
			"labels":    map[string]any{"prometheus": label},
		},
	}}
}

func bufferedLogger() (pkg.Logger, *bytes.Buffer) {
	buf := bytes.NewBuffer(nil)
	return log.NewLogger(log.WithOutput(buf)), buf
}

func exists(t *testing.T, c client.Client, obj client.Object, name string) bool {
	t.Helper()
	err := c.Get(context.Background(), types.NamespacedName{Namespace: "test-ns", Name: name}, obj)
	if apierrors.IsNotFound(err) {
		return false
	}
	require.NoError(t, err)
	return true
}

func prometheusExists(t *testing.T, c *fakeKubeClient, name string) bool {
	t.Helper()
	_, err := c.Dynamic().Resource(prometheusGVRWant).Namespace("test-ns").Get(context.Background(), name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return false
	}
	require.NoError(t, err)
	return true
}

// TestDeleteWorkloads_LabelSelectorDeletesEveryMatch is the case the hook was
// missing: several StatefulSets share the label selector and all of them have to
// go, not just the one named in Args.
func TestDeleteWorkloads_LabelSelectorDeletesEveryMatch(t *testing.T) {
	args := newArgs()
	args.ObjectName = ""

	kubeClient := newKubeClient(nil,
		statefulSet("data-1", map[string]string{"app": "data"}),
		statefulSet("data-2", map[string]string{"app": "data"}),
		statefulSet("other", map[string]string{"app": "other"}),
	)

	logger, logs := bufferedLogger()

	require.NoError(t, deleteWorkloads(context.Background(), kubeClient, logger, args))

	assert.False(t, exists(t, kubeClient, &appsv1.StatefulSet{}, "data-1"))
	assert.False(t, exists(t, kubeClient, &appsv1.StatefulSet{}, "data-2"))
	assert.True(t, exists(t, kubeClient, &appsv1.StatefulSet{}, "other"))
	assert.Contains(t, logs.String(), "data-1")
	assert.Contains(t, logs.String(), "data-2")
	assert.NotContains(t, logs.String(), "deleting by objectName only")
}

// TestDeleteWorkloads_LabelSelectorMatchesNothing covers the dangerous case: the
// PVCs are already gone by then, so a selector that matches no workload has to be
// warned about instead of passing as success.
func TestDeleteWorkloads_LabelSelectorMatchesNothing(t *testing.T) {
	args := newArgs()
	args.ObjectName = ""

	kubeClient := newKubeClient(nil, statefulSet("other", map[string]string{"app": "other"}))

	logger, logs := bufferedLogger()

	require.NoError(t, deleteWorkloads(context.Background(), kubeClient, logger, args))

	assert.True(t, exists(t, kubeClient, &appsv1.StatefulSet{}, "other"))
	assert.Contains(t, logs.String(), "no objects matched the label selector")
}

// TestDeleteWorkloads_ObjectNameWinsOverLabelSelector keeps the behaviour every
// current consumer relies on, but the ignored selector has to be logged.
func TestDeleteWorkloads_ObjectNameWinsOverLabelSelector(t *testing.T) {
	args := newArgs()

	kubeClient := newKubeClient(nil,
		statefulSet("data-set", map[string]string{"app": "data"}),
		statefulSet("data-2", map[string]string{"app": "data"}),
	)

	logger, logs := bufferedLogger()

	require.NoError(t, deleteWorkloads(context.Background(), kubeClient, logger, args))

	assert.False(t, exists(t, kubeClient, &appsv1.StatefulSet{}, "data-set"))
	assert.True(t, exists(t, kubeClient, &appsv1.StatefulSet{}, "data-2"))
	assert.Contains(t, logs.String(), "deleting by objectName only")
	assert.Contains(t, logs.String(), "app=data")
}

// TestDeleteWorkloads_AlreadyDeletedIsNotAnError - the object may be gone already
// (previous run, manual cleanup); that must not surface as a hook error.
func TestDeleteWorkloads_AlreadyDeletedIsNotAnError(t *testing.T) {
	args := newArgs()

	logger, _ := bufferedLogger()

	require.NoError(t, deleteWorkloads(context.Background(), newKubeClient(nil), logger, args))
}

func TestDeleteWorkloads_DeploymentByName(t *testing.T) {
	args := newArgs()
	args.ObjectKind = "Deployment"
	args.ObjectName = "data-deploy"

	kubeClient := newKubeClient(nil, &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "data-deploy", Namespace: "test-ns"},
	})

	logger, _ := bufferedLogger()

	require.NoError(t, deleteWorkloads(context.Background(), kubeClient, logger, args))
	assert.False(t, exists(t, kubeClient, &appsv1.Deployment{}, "data-deploy"))
}

func TestDeleteWorkloads_PrometheusByName(t *testing.T) {
	args := newArgs()
	args.ObjectKind = "Prometheus"
	args.ObjectName = "main"
	args.LabelSelectorKey = "prometheus"
	args.LabelSelectorValue = "main"

	kubeClient := newKubeClient([]*unstructured.Unstructured{
		prometheus("main", "main"), prometheus("longterm", "longterm"),
	})

	logger, _ := bufferedLogger()

	require.NoError(t, deleteWorkloads(context.Background(), kubeClient, logger, args))

	assert.False(t, prometheusExists(t, kubeClient, "main"))
	assert.True(t, prometheusExists(t, kubeClient, "longterm"))
}

func TestDeleteWorkloads_PrometheusByLabelSelector(t *testing.T) {
	args := newArgs()
	args.ObjectKind = "Prometheus"
	args.ObjectName = ""
	args.LabelSelectorKey = "prometheus"
	args.LabelSelectorValue = "main"

	kubeClient := newKubeClient([]*unstructured.Unstructured{
		prometheus("main", "main"), prometheus("longterm", "longterm"),
	})

	logger, _ := bufferedLogger()

	require.NoError(t, deleteWorkloads(context.Background(), kubeClient, logger, args))

	assert.False(t, prometheusExists(t, kubeClient, "main"))
	assert.True(t, prometheusExists(t, kubeClient, "longterm"))
}

// TestDeleteWorkloads_NoNameAndNoSelector guards the destructive function itself:
// an empty selector would match every object of that kind in the namespace.
func TestDeleteWorkloads_NoNameAndNoSelector(t *testing.T) {
	args := newArgs()
	args.ObjectName = ""
	args.LabelSelectorKey = ""
	args.LabelSelectorValue = ""

	kubeClient := newKubeClient(nil, statefulSet("data-set", map[string]string{"app": "data"}))

	logger, _ := bufferedLogger()

	err := deleteWorkloads(context.Background(), kubeClient, logger, args)
	require.ErrorContains(t, err, "objectName or label selector must be set")
	assert.True(t, exists(t, kubeClient, &appsv1.StatefulSet{}, "data-set"))
}

func TestDeleteWorkloads_UnknownObjectKind(t *testing.T) {
	args := newArgs()
	args.ObjectKind = "CronJob"

	logger, _ := bufferedLogger()

	err := deleteWorkloads(context.Background(), newKubeClient(nil), logger, args)
	require.ErrorContains(t, err, "unknown object kind CronJob")
}

// TestStorageClassChange_MisconfiguredArgsFailTheHook - both checks have to fail the
// hook before it deletes any PVC, not just log something.
func TestStorageClassChange_MisconfiguredArgsFailTheHook(t *testing.T) {
	t.Run("no name and no selector", func(t *testing.T) {
		args := newArgs()
		args.ObjectName = ""
		args.LabelSelectorKey = ""
		args.LabelSelectorValue = ""

		err := storageClassChange(context.Background(), &pkg.HookInput{Logger: log.NewNop()}, args)
		require.ErrorContains(t, err, "objectName or label selector must be set")
	})

	t.Run("unknown object kind", func(t *testing.T) {
		args := newArgs()
		args.ObjectKind = "CronJob"

		err := storageClassChange(context.Background(), &pkg.HookInput{Logger: log.NewNop()}, args)
		require.ErrorContains(t, err, "unknown object kind CronJob")
	})
}
