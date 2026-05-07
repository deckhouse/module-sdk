// Package framework provides a deckhouse-style testing framework for module-sdk hooks.
//
// It is inspired by deckhouse/testing/hooks but does not depend on
// addon-operator or shell-operator. Internally it uses a fake Kubernetes
// client (k8s.io/client-go/dynamic/fake) to simulate cluster state.
//
// Typical usage:
//
//	func TestMyHook(t *testing.T) {
//	    f := framework.HookExecutionConfigInit(t, hookConfig, MyHookHandler, `{}`, `{}`)
//
//	    f.KubeStateSet(`
//	---
//	apiVersion: v1
//	kind: Node
//	metadata:
//	  name: kube-worker-1
//	`)
//
//	    f.RunHook()
//
//	    require.NoError(t, f.HookError())
//	    require.Len(t, f.Snapshots().Get("nodes"), 1)
//	    require.Equal(t, "value", f.ValuesGet("my.field").String())
//	}
package framework
