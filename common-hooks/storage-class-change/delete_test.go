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

// fakeKubeClient satisfies pkg.KubernetesClient with the controller-runtime and
// dynamic fake clients. The framework's own fake (testing/framework/dc.go) starts
// with an empty controller-runtime tracker, so deletions are covered here instead.
type fakeKubeClient struct {
	client.Client
	dyn dynamic.Interface
}

func (c *fakeKubeClient) Dynamic() dynamic.Interface { return c.dyn }

func statefulSet(name string, labels map[string]string) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "test-ns", Labels: labels},
	}
}

func typedClient(objs ...client.Object) *fakeKubeClient {
	return &fakeKubeClient{
		Client: crfake.NewClientBuilder().WithScheme(clientgoscheme.Scheme).WithObjects(objs...).Build(),
	}
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

// TestDeleteWorkloads_LabelSelectorDeletesEveryMatch is the case the hook was
// missing: several StatefulSets share the label selector and all of them have to
// go, not just the one named in Args.
func TestDeleteWorkloads_LabelSelectorDeletesEveryMatch(t *testing.T) {
	args := newArgs()
	args.ObjectName = ""

	kubeClient := typedClient(
		statefulSet("data-1", map[string]string{"app": "data"}),
		statefulSet("data-2", map[string]string{"app": "data"}),
		statefulSet("other", map[string]string{"app": "other"}),
	)

	logger, logs := bufferedLogger()

	require.NoError(t, deleteWorkloads(context.Background(), kubeClient, logger, args))

	assert.False(t, exists(t, kubeClient, &appsv1.StatefulSet{}, "data-1"))
	assert.False(t, exists(t, kubeClient, &appsv1.StatefulSet{}, "data-2"))
	assert.True(t, exists(t, kubeClient, &appsv1.StatefulSet{}, "other"))
	assert.NotContains(t, logs.String(), "deleting by objectName only")
}

// TestDeleteWorkloads_ObjectNameWinsOverLabelSelector keeps the behaviour every
// current consumer relies on, but the ignored selector has to be logged.
func TestDeleteWorkloads_ObjectNameWinsOverLabelSelector(t *testing.T) {
	args := newArgs()

	kubeClient := typedClient(
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

func TestDeleteWorkloads_DeploymentByName(t *testing.T) {
	args := newArgs()
	args.ObjectKind = "Deployment"
	args.ObjectName = "data-deploy"

	kubeClient := typedClient(&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "data-deploy", Namespace: "test-ns"},
	})

	logger, _ := bufferedLogger()

	require.NoError(t, deleteWorkloads(context.Background(), kubeClient, logger, args))
	assert.False(t, exists(t, kubeClient, &appsv1.Deployment{}, "data-deploy"))
}

func TestDeleteWorkloads_PrometheusByLabelSelector(t *testing.T) {
	args := newArgs()
	args.ObjectKind = "Prometheus"
	args.ObjectName = ""
	args.LabelSelectorKey = "prometheus"
	args.LabelSelectorValue = "main"

	prometheus := func(name, label string) *unstructured.Unstructured {
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

	dyn := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(
		runtime.NewScheme(),
		map[schema.GroupVersionResource]string{prometheusGVR: "PrometheusList"},
		prometheus("main", "main"), prometheus("longterm", "longterm"),
	)

	logger, _ := bufferedLogger()

	require.NoError(t, deleteWorkloads(context.Background(), &fakeKubeClient{dyn: dyn}, logger, args))

	resource := dyn.Resource(prometheusGVR).Namespace("test-ns")

	_, err := resource.Get(context.Background(), "main", metav1.GetOptions{})
	assert.True(t, apierrors.IsNotFound(err), "main prometheus must be deleted, got %v", err)

	_, err = resource.Get(context.Background(), "longterm", metav1.GetOptions{})
	assert.NoError(t, err, "longterm prometheus must be kept")
}

// TestStorageClassChange_NoNameAndNoSelectorFails checks the hook bails out before
// touching the cluster: with neither a name nor a selector the label lookup would
// match every object of that kind in the namespace.
func TestStorageClassChange_NoNameAndNoSelectorFails(t *testing.T) {
	args := newArgs()
	args.ObjectName = ""
	args.LabelSelectorKey = ""
	args.LabelSelectorValue = ""

	err := storageClassChange(context.Background(), &pkg.HookInput{Logger: log.NewNop()}, args)
	require.ErrorContains(t, err, "objectName or label selector must be set")
}

func TestDeleteWorkloads_UnknownObjectKind(t *testing.T) {
	args := newArgs()
	args.ObjectKind = "CronJob"

	logger, _ := bufferedLogger()

	err := deleteWorkloads(context.Background(), typedClient(), logger, args)
	require.ErrorContains(t, err, "unknown object kind CronJob")
}
