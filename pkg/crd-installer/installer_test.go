package crdinstaller

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apimachineryv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/fake"
	"testing"
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
}
