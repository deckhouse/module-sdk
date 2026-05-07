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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/testing/framework"
)

// hookConfigFor mirrors what RegisterHook builds, but without registering
// anything in the global registry. Functional tests use this to feed the
// framework a complete *pkg.HookConfig.
func hookConfigFor(args Args) *pkg.HookConfig {
	return &pkg.HookConfig{
		OnBeforeHelm: &pkg.OrderedConfig{Order: 1},
		Kubernetes: []pkg.KubernetesConfig{
			{
				Name:       "pvcs",
				APIVersion: "v1",
				Kind:       "PersistentVolumeClaim",
				NamespaceSelector: &pkg.NamespaceSelector{
					NameSelector: &pkg.NameSelector{MatchNames: []string{args.Namespace}},
				},
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{args.LabelSelectorKey: args.LabelSelectorValue},
				},
				JqFilter: pvcFilter,
			},
			{
				Name:       "pods",
				APIVersion: "v1",
				Kind:       "Pod",
				NamespaceSelector: &pkg.NamespaceSelector{
					NameSelector: &pkg.NameSelector{MatchNames: []string{args.Namespace}},
				},
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{args.LabelSelectorKey: args.LabelSelectorValue},
				},
				JqFilter: podFilter,
			},
			{
				Name:       "storageClasses",
				APIVersion: "storage.k8s.io/v1",
				Kind:       "StorageClass",
				JqFilter:   storageClassFilter,
			},
		},
	}
}

func newArgs() Args {
	return Args{
		ModuleName:         "myModule",
		Namespace:          "test-ns",
		LabelSelectorKey:   "app",
		LabelSelectorValue: "data",
		ObjectKind:         "StatefulSet",
		ObjectName:         "data-set",
	}
}

// TestStorageClassChange_DefaultStorageClassWritesEffectiveValue exercises
// the snapshot pipeline end-to-end: storage classes from the cluster,
// PVCs filtered by the label selector, and the resulting internal value.
func TestStorageClassChange_DefaultStorageClassWritesEffectiveValue(t *testing.T) {
	args := newArgs()

	const state = `
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: standard
  annotations:
    storageclass.kubernetes.io/is-default-class: "true"
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: fast
`

	f := framework.HookExecutionConfigInit(t, hookConfigFor(args), func(ctx context.Context, in *pkg.HookInput) error {
		return storageClassChange(ctx, in, args)
	}, `{}`, `{}`)
	f.KubeStateSet(state)
	f.RunHook()

	require.NoError(t, f.HookError())

	// The hook discovers the default SC and writes it to the internal path.
	val := f.ValuesGet("myModule.internal.effectiveStorageClass").String()
	assert.Equal(t, "standard", val)
}

// TestStorageClassChange_ConfigOverridesDefault asserts that an explicit
// global.modules.storageClass override beats the cluster default.
func TestStorageClassChange_ConfigOverridesDefault(t *testing.T) {
	args := newArgs()

	const state = `
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: standard
  annotations:
    storageclass.kubernetes.io/is-default-class: "true"
`

	const config = `{"global":{"modules":{"storageClass":"premium"}}}`

	f := framework.HookExecutionConfigInit(t, hookConfigFor(args), func(ctx context.Context, in *pkg.HookInput) error {
		return storageClassChange(ctx, in, args)
	}, `{}`, config)
	f.KubeStateSet(state)
	f.RunHook()

	require.NoError(t, f.HookError())
	assert.Equal(t, "premium", f.ValuesGet("myModule.internal.effectiveStorageClass").String())
}

// TestStorageClassChange_LabelSelectorScopesPVCs ensures the hook only
// considers PVCs whose labels match the configured selector.
func TestStorageClassChange_LabelSelectorScopesPVCs(t *testing.T) {
	args := newArgs()

	const state = `
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: standard
  annotations:
    storageclass.kubernetes.io/is-default-class: "true"
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: matching
  namespace: test-ns
  labels:
    app: data
spec:
  storageClassName: legacy
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: irrelevant
  namespace: test-ns
  labels:
    app: other
spec:
  storageClassName: ignored
`

	f := framework.HookExecutionConfigInit(t, hookConfigFor(args), func(ctx context.Context, in *pkg.HookInput) error {
		return storageClassChange(ctx, in, args)
	}, `{}`, `{}`)
	f.KubeStateSet(state)
	f.RunHook()

	require.NoError(t, f.HookError())
	// The current PVC's storageClassName ("legacy") wins over the default.
	assert.Equal(t, "legacy", f.ValuesGet("myModule.internal.effectiveStorageClass").String())
}

// TestStorageClassChange_BeforeHookCheckGate validates that returning false
// from BeforeHookCheck short-circuits the hook (no values, no errors).
func TestStorageClassChange_BeforeHookCheckGate(t *testing.T) {
	args := newArgs()
	args.BeforeHookCheck = func(_ *pkg.HookInput) bool { return false }

	f := framework.HookExecutionConfigInit(t, hookConfigFor(args), func(ctx context.Context, in *pkg.HookInput) error {
		return storageClassChange(ctx, in, args)
	}, `{}`, `{}`)
	f.KubeStateSet(`
---
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: standard
  annotations:
    storageclass.kubernetes.io/is-default-class: "true"
`)
	f.RunHook()

	require.NoError(t, f.HookError())
	assert.False(t, f.ValuesGet("myModule.internal.effectiveStorageClass").Exists())
}
