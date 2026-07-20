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
	"k8s.io/client-go/dynamic/fake"
)

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

	// Regression: a comment-only yaml document decodes to a nil object and must be skipped,
	// not panic on the nil dereference.
	t.Run("skips comment-only documents", func(t *testing.T) {
		inst := NewCRDsInstaller(fc, []string{"testdata/5_comment.yaml"})
		require.NoError(t, inst.Run(context.Background()))

		_, err := fc.Resource(gvr).Get(context.Background(), "comments.example.com", apimachineryv1.GetOptions{})
		require.NoError(t, err)
	})
}
