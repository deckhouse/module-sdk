package crdinstaller

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apimachineryv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/dynamic/fake"
	k8stesting "k8s.io/client-go/testing"
)

func countUpdates(actions []k8stesting.Action, name string) int {
	var n int

	for _, action := range actions {
		// k8stesting.CreateAction and UpdateAction are the same interface, so a create
		// satisfies both: the verb is the only way to tell them apart. The subresource check
		// keeps the storedVersions status write out of the count.
		if action.GetVerb() != "update" || action.GetSubresource() != "" {
			continue
		}

		obj, ok := action.(k8stesting.UpdateAction).GetObject().(*unstructured.Unstructured)
		if ok && obj.GetName() == name {
			n++
		}
	}

	return n
}

// storeAsWire makes the fake client behave the way the wire does. The tracker keeps the
// exact Go values it is handed, so without this an object built with float64 numbers
// reads back as float64 and a test can pass on state a real apiserver would never
// return.
func storeAsWire(t *testing.T, fc *fake.FakeDynamicClient) {
	t.Helper()

	react := func(action k8stesting.Action) (bool, runtime.Object, error) {
		withObject, ok := action.(interface{ GetObject() runtime.Object })
		if !ok {
			return false, nil, nil
		}

		obj, ok := withObject.GetObject().(*unstructured.Unstructured)
		if !ok {
			return false, nil, nil
		}

		data, err := json.Marshal(obj.Object)
		require.NoError(t, err)

		wire := map[string]any{}
		require.NoError(t, json.Unmarshal(data, &wire))

		// ObjectMeta.Labels/Annotations are omitempty, so the apiserver never returns an
		// empty map for them — it returns nothing
		for _, field := range []string{"labels", "annotations"} {
			if m, ok := nestedValue(wire, "metadata", field).(map[string]any); ok && len(m) == 0 {
				unstructured.RemoveNestedField(wire, "metadata", field)
			}
		}

		obj.Object = wire

		// fall through to the object tracker, which stores a deep copy of what we just fixed
		return false, nil, nil
	}

	fc.PrependReactor("create", "*", react)
	fc.PrependReactor("update", "*", react)
}

func TestCRDInstaller(t *testing.T) {
	crdScheme := runtime.NewScheme()

	// Регистрируем v1 версию CRD в схеме
	if err := v1.AddToScheme(crdScheme); err != nil {
		fmt.Println("Error adding apiextensions.k8s.io/v1 to scheme:", err)
		t.Fatal(err)
	}
	gvr := schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1",
		Resource: "customresourcedefinitions",
	}

	fc := fake.NewSimpleDynamicClient(crdScheme)
	storeAsWire(t, fc)

	t.Run("install CRD", func(t *testing.T) {
		inst := NewCRDsInstaller(fc, []string{"testdata/1_example.yaml"}, WithExtraLabels(map[string]string{"heritage": "deckhouse"}))
		err := inst.Run(context.Background())
		require.NoError(t, err)

		un, err := fc.Resource(gvr).Get(context.Background(), "widgets.example.com", apimachineryv1.GetOptions{})
		require.NoError(t, err)

		assert.Equal(t, "widgets.example.com", un.GetName())
		assert.Equal(t, map[string]string{"foo": "bar", "heritage": "deckhouse"}, un.GetLabels())
		assert.Equal(t, map[string]string{"bar": "baz"}, un.GetAnnotations())
		var crd v1.CustomResourceDefinition
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(un.Object, &crd)
		require.NoError(t, err)

		f1 := crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["spec"].Properties["field1"].Type
		assert.Equal(t, "string", f1)
	})

	t.Run("update CRD", func(t *testing.T) {
		inst := NewCRDsInstaller(fc, []string{"testdata/2_example.yaml"}, WithExtraLabels(map[string]string{"another": "lab"}))
		err := inst.Run(context.Background())
		require.NoError(t, err)

		un, err := fc.Resource(gvr).Get(context.Background(), "widgets.example.com", apimachineryv1.GetOptions{})
		require.NoError(t, err)

		assert.Equal(t, map[string]string{"foo": "bar", "one": "new", "another": "lab"}, un.GetLabels())
		assert.Equal(t, map[string]string{"bar": "baz", "two": "new"}, un.GetAnnotations())
		var crd v1.CustomResourceDefinition
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(un.Object, &crd)
		require.NoError(t, err)

		f4 := crd.Spec.Versions[0].Schema.OpenAPIV3Schema.Properties["spec"].Properties["field4"].Type
		assert.Equal(t, "boolean", f4)
	})

	// Regression: vendor schema extensions (x-kubernetes-sensitive-data and friends) must
	// survive the install path. Decoding CRDs into the typed struct drops them, so the
	// installer must apply the object as unstructured.
	t.Run("preserves vendor schema extensions", func(t *testing.T) {
		inst := NewCRDsInstaller(fc, []string{"testdata/3_sensitive.yaml"})
		require.NoError(t, inst.Run(context.Background()))

		un, err := fc.Resource(gvr).Get(context.Background(), "secrets.example.com", apimachineryv1.GetOptions{})
		require.NoError(t, err)

		versions, found, err := unstructured.NestedSlice(un.Object, "spec", "versions")
		require.NoError(t, err)
		require.True(t, found, "spec.versions must be present")

		token, found, err := unstructured.NestedMap(versions[0].(map[string]any),
			"schema", "openAPIV3Schema", "properties", "spec", "properties", "token")
		require.NoError(t, err)
		require.True(t, found, "token property must be present")
		assert.Equal(t, true, token["x-kubernetes-sensitive-data"],
			"vendor extension must survive the install path")
	})

	// Regression: updating an existing CRD must apply the manifest spec (with vendor
	// extensions) while preserving server-managed metadata (finalizers) and the in-cluster
	// conversion config, which is never present in the manifest.
	t.Run("update preserves finalizers and in-cluster conversion", func(t *testing.T) {
		seed := &unstructured.Unstructured{Object: map[string]any{
			"apiVersion": "apiextensions.k8s.io/v1",
			"kind":       "CustomResourceDefinition",
			"metadata": map[string]any{
				"name":       "things.example.com",
				"finalizers": []any{"customresourcecleanup.apiextensions.k8s.io"},
			},
			"spec": map[string]any{
				"group": "example.com",
				"names": map[string]any{
					"kind": "Thing", "listKind": "ThingList", "plural": "things", "singular": "thing",
				},
				"scope": "Namespaced",
				"conversion": map[string]any{
					"strategy": "Webhook",
					"webhook": map[string]any{
						"conversionReviewVersions": []any{"v1"},
						"clientConfig": map[string]any{
							"service": map[string]any{"name": "conv", "namespace": "default"},
						},
					},
				},
				"versions": []any{
					map[string]any{"name": "v1", "served": true, "storage": true},
				},
			},
		}}
		_, err := fc.Resource(gvr).Create(context.Background(), seed, apimachineryv1.CreateOptions{})
		require.NoError(t, err)

		inst := NewCRDsInstaller(fc, []string{"testdata/4_conversion.yaml"})
		require.NoError(t, inst.Run(context.Background()))

		un, err := fc.Resource(gvr).Get(context.Background(), "things.example.com", apimachineryv1.GetOptions{})
		require.NoError(t, err)

		finalizers, found, err := unstructured.NestedStringSlice(un.Object, "metadata", "finalizers")
		require.NoError(t, err)
		require.True(t, found, "finalizers must be preserved")
		assert.Contains(t, finalizers, "customresourcecleanup.apiextensions.k8s.io")

		_, found, err = unstructured.NestedMap(un.Object, "spec", "conversion")
		require.NoError(t, err)
		assert.True(t, found, "in-cluster conversion must be preserved")

		versions, _, err := unstructured.NestedSlice(un.Object, "spec", "versions")
		require.NoError(t, err)
		token, found, err := unstructured.NestedMap(versions[0].(map[string]any),
			"schema", "openAPIV3Schema", "properties", "spec", "properties", "token")
		require.NoError(t, err)
		require.True(t, found, "manifest schema must be applied")
		assert.Equal(t, true, token["x-kubernetes-sensitive-data"])
	})

	// Regression: keys the apiserver does not know must never be sent. It prunes them
	// anyway, one "unknown field" warning per occurrence, and the pruning is what made
	// the desired spec permanently differ from the stored one.
	//
	// This only checks that sanitize is reached from Run and applies to the whole schema
	// tree; which keys survive at which nesting position is openapi.Prune's own job and is
	// covered by TestRoundTrip* in that package.
	t.Run("strips unknown schema extensions", func(t *testing.T) {
		inst := NewCRDsInstaller(fc, []string{"testdata/6_unknown_extensions.yaml"})
		require.NoError(t, inst.Run(context.Background()))

		un, err := fc.Resource(gvr).Get(context.Background(), "extensions.example.com", apimachineryv1.GetOptions{})
		require.NoError(t, err)

		versions, _, err := unstructured.NestedSlice(un.Object, "spec", "versions")
		require.NoError(t, err)

		root, found, err := unstructured.NestedMap(versions[0].(map[string]any), "schema", "openAPIV3Schema")
		require.NoError(t, err)
		require.True(t, found)

		assert.NotContains(t, root, "x-doc-examples")

		token, found, err := unstructured.NestedMap(root, "properties", "spec", "properties", "token")
		require.NoError(t, err)
		require.True(t, found)

		assert.Equal(t, true, token["x-kubernetes-sensitive-data"], "the one extension that must survive")
		assert.NotContains(t, token, "x-doc-examples")
	})

	// Regression: a CRD field this build's apiextensions-apiserver does not model must
	// still reach an apiserver that understands it — sanitize prunes schemas, not the
	// document.
	t.Run("keeps CRD fields outside the schema", func(t *testing.T) {
		inst := NewCRDsInstaller(fc, []string{"testdata/9_unknown_crd_field.yaml"})
		require.NoError(t, inst.Run(context.Background()))

		un, err := fc.Resource(gvr).Get(context.Background(), "futures.example.com", apimachineryv1.GetOptions{})
		require.NoError(t, err)

		versions, _, err := unstructured.NestedSlice(un.Object, "spec", "versions")
		require.NoError(t, err)

		assert.Equal(t, "a field from a newer apiserver", versions[0].(map[string]any)["fieldFromTheFuture"])
		assert.NotContains(t, un.Object, "status", "the manifest declares no status, so none must be sent")

		// the manifest has no labels and the installer adds none, so the second run must see
		// the nil the cluster returns as equal to what it would write
		before := countUpdates(fc.Actions(), "futures.example.com")

		require.NoError(t, NewCRDsInstaller(fc, []string{"testdata/9_unknown_crd_field.yaml"}).Run(context.Background()))

		assert.Equal(t, before, countUpdates(fc.Actions(), "futures.example.com"),
			"a CRD without labels must not churn")
	})

	// Regression: applying a manifest over the state the apiserver derives from it must be
	// a no-op. The apiserver prunes the x-doc-* keys, defaults spec.names and returns whole
	// numbers as int64, so before the fix the desired spec could never equal the stored one
	// and every single run issued a full Update of every CRD.
	//
	// The fake client prunes and defaults nothing, so the derived state has to be installed
	// explicitly — reinstalling the same file twice would pass either way and prove nothing.
	t.Run("applying a manifest over its stored form does not update", func(t *testing.T) {
		inst := NewCRDsInstaller(fc, []string{"testdata/7_churn_stored.yaml"})
		require.NoError(t, inst.Run(context.Background()))

		require.Zero(t, countUpdates(fc.Actions(), "churns.example.com"),
			"a CRD that is not in the cluster is created, not updated")

		inst = NewCRDsInstaller(fc, []string{"testdata/8_churn_manifest.yaml"})
		require.NoError(t, inst.Run(context.Background()))

		assert.Zero(t, countUpdates(fc.Actions(), "churns.example.com"),
			"the doc keys, the defaulted names and the numeric bounds must not read as a diff")
	})

	// Regression: sanitize runs before the CRD is queued, so an error from it used to abort
	// the whole file and silently skip every document after the bad one.
	t.Run("a bad document does not skip the rest of the file", func(t *testing.T) {
		inst := NewCRDsInstaller(fc, []string{"testdata/10_multi_document.yaml"})
		require.Error(t, inst.Run(context.Background()), "the broken document must still be reported")

		for _, name := range []string{"firsts.example.com", "lasts.example.com"} {
			_, err := fc.Resource(gvr).Get(context.Background(), name, apimachineryv1.GetOptions{})
			require.NoError(t, err, "%s is a valid CRD in the same file and must be installed", name)
		}
	})

	// Regression: annotations belong to whoever wrote them. Replacing the map wholesale
	// dropped the Helm ownership keys, and the next helm upgrade of that chart then failed
	// on "invalid ownership metadata".
	t.Run("keeps annotations written by other actors", func(t *testing.T) {
		// widgets.example.com is already in the cluster from the subtests above; give it the
		// ownership annotations a helm-installed CRD carries
		_, err := fc.Resource(gvr).Patch(context.Background(), "widgets.example.com", types.MergePatchType,
			[]byte(`{"metadata":{"annotations":{"meta.helm.sh/release-name":"some-chart"}}}`),
			apimachineryv1.PatchOptions{})
		require.NoError(t, err)

		inst := NewCRDsInstaller(fc, []string{"testdata/2_example.yaml"})
		require.NoError(t, inst.Run(context.Background()))

		un, err := fc.Resource(gvr).Get(context.Background(), "widgets.example.com", apimachineryv1.GetOptions{})
		require.NoError(t, err)

		assert.Equal(t, map[string]string{
			"bar": "baz", "two": "new", "meta.helm.sh/release-name": "some-chart",
		}, un.GetAnnotations())

		before := countUpdates(fc.Actions(), "widgets.example.com")

		require.NoError(t, NewCRDsInstaller(fc, []string{"testdata/2_example.yaml"}).Run(context.Background()))

		assert.Equal(t, before, countUpdates(fc.Actions(), "widgets.example.com"),
			"a foreign annotation must not read as a diff either")
	})

	// Regression: a comment-only yaml document decodes to a nil object and must be skipped,
	// not panic on the nil dereference.
	t.Run("skips comment-only documents", func(t *testing.T) {
		inst := NewCRDsInstaller(fc, []string{"testdata/5_comment.yaml"})
		require.NoError(t, inst.Run(context.Background()))

		_, err := fc.Resource(gvr).Get(context.Background(), "comments.example.com", apimachineryv1.GetOptions{})
		require.NoError(t, err)
	})
}
