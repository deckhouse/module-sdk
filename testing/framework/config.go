package framework

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	admissionv1 "k8s.io/api/admission/v1"
	admissionregv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	authorizationv1 "k8s.io/api/authorization/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	certificatesv1 "k8s.io/api/certificates/v1"
	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	eventsv1 "k8s.io/api/events/v1"
	flowcontrolv1 "k8s.io/api/flowcontrol/v1"
	networkingv1 "k8s.io/api/networking/v1"
	nodev1 "k8s.io/api/node/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	storagev1 "k8s.io/api/storage/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"

	"github.com/deckhouse/deckhouse/pkg/log"

	"github.com/deckhouse/module-sdk/internal/metric"
	"github.com/deckhouse/module-sdk/pkg"
)

// HookFunc is the type of hook handler functions tested by the framework.
type HookFunc = pkg.HookFunc[*pkg.HookInput]

// HookExecutionConfig is the main entry point for hook tests. It encapsulates
// the hook under test, a fake Kubernetes cluster, values stores, and collectors
// for patch operations and metrics.
//
// A HookExecutionConfig is created with HookExecutionConfigInit (deckhouse-style)
// or NewHookExecutionConfig (with options).
type HookExecutionConfig struct {
	t testing.TB

	hookConfig  *pkg.HookConfig
	hookHandler HookFunc

	scheme             *runtime.Scheme
	unstructuredScheme *runtime.Scheme
	fakeClient         *dynamicfake.FakeDynamicClient
	gvrToListKind      map[schema.GroupVersionResource]string
	gvkToGVR           map[schema.GroupVersionKind]schema.GroupVersionResource

	values       *valuesStore
	configValues *valuesStore

	patchCollector   *recordingPatchCollector
	metricsCollector *metric.Collector
	snapshots        snapshotsMap
	hookError        error
	loggerOutput     *bytes.Buffer
	dc               *frameworkDC

	logger *log.Logger
}

// DependencyContainer returns the framework's dependency container so that
// tests can override its HTTP / registry / clock components before RunHook.
//
// Example:
//
//	hec.DependencyContainer().SetHTTPClient(myFakeHTTP)
func (h *HookExecutionConfig) DependencyContainer() *frameworkDC {
	if h.dc == nil {
		h.dc = newFrameworkDC(h.fakeClient, h.scheme)
	}
	return h.dc
}

// HookExecutionConfigInit creates a deckhouse-style execution config.
//
// initValues and initConfigValues are JSON or YAML strings representing the
// initial Helm values and module config values. Pass "{}" or "" if not needed.
func HookExecutionConfigInit(t testing.TB, config *pkg.HookConfig, handler HookFunc, initValues, initConfigValues string) *HookExecutionConfig {
	return NewHookExecutionConfig(t, config, handler,
		WithInitialValues(initValues),
		WithInitialConfigValues(initConfigValues),
	)
}

// NewHookExecutionConfig creates an execution config with options.
func NewHookExecutionConfig(t testing.TB, config *pkg.HookConfig, handler HookFunc, opts ...Option) *HookExecutionConfig {
	t.Helper()

	cfg := &execOptions{
		initValues:       "{}",
		initConfigValues: "{}",
	}
	for _, o := range opts {
		o.apply(cfg)
	}

	scheme := defaultScheme(cfg.extraSchemeBuilders...)
	unstructuredScheme := newUnstructuredScheme(scheme)

	hec := &HookExecutionConfig{
		t:                  t,
		hookConfig:         config,
		hookHandler:        handler,
		scheme:             scheme,
		unstructuredScheme: unstructuredScheme,
		gvrToListKind:      defaultGVRToListKind(scheme),
		gvkToGVR:           make(map[schema.GroupVersionKind]schema.GroupVersionResource),
		loggerOutput:       bytes.NewBuffer(nil),
	}

	hec.logger = log.NewLogger(log.WithOutput(hec.loggerOutput))

	var err error
	hec.values, err = newValuesStore(cfg.initValues)
	if err != nil {
		t.Fatalf("framework: parse initial values: %v", err)
	}
	hec.configValues, err = newValuesStore(cfg.initConfigValues)
	if err != nil {
		t.Fatalf("framework: parse initial config values: %v", err)
	}

	hec.fakeClient = dynamicfake.NewSimpleDynamicClientWithCustomListKinds(unstructuredScheme, hec.gvrToListKind)

	for _, crd := range cfg.crds {
		hec.RegisterCRD(crd.group, crd.version, crd.kind, crd.namespaced)
	}

	return hec
}

// defaultScheme returns the default Kubernetes scheme registering all standard
// API groups (the same set used by sigs.k8s.io/controller-runtime by default)
// plus apiextensions/v1 and any extra builders the caller provided.
func defaultScheme(extraBuilders ...runtime.SchemeBuilder) *runtime.Scheme {
	scheme := runtime.NewScheme()

	for _, b := range []runtime.SchemeBuilder{
		admissionv1.SchemeBuilder,
		admissionregv1.SchemeBuilder,
		apiextensionsv1.SchemeBuilder,
		appsv1.SchemeBuilder,
		authenticationv1.SchemeBuilder,
		authorizationv1.SchemeBuilder,
		autoscalingv1.SchemeBuilder,
		autoscalingv2.SchemeBuilder,
		batchv1.SchemeBuilder,
		certificatesv1.SchemeBuilder,
		coordinationv1.SchemeBuilder,
		corev1.SchemeBuilder,
		discoveryv1.SchemeBuilder,
		eventsv1.SchemeBuilder,
		flowcontrolv1.SchemeBuilder,
		networkingv1.SchemeBuilder,
		nodev1.SchemeBuilder,
		policyv1.SchemeBuilder,
		rbacv1.SchemeBuilder,
		schedulingv1.SchemeBuilder,
		storagev1.SchemeBuilder,
	} {
		utilruntime.Must(b.AddToScheme(scheme))
	}

	for _, builder := range extraBuilders {
		utilruntime.Must(builder.AddToScheme(scheme))
	}

	return scheme
}

// defaultGVRToListKind returns a list-kind mapping for all GVKs registered in
// the scheme. The fake dynamic client uses this for List operations.
func defaultGVRToListKind(scheme *runtime.Scheme) map[schema.GroupVersionResource]string {
	out := map[schema.GroupVersionResource]string{}
	for gvk := range scheme.AllKnownTypes() {
		gvr, _ := meta.UnsafeGuessKindToResource(gvk)
		out[gvr] = gvk.Kind + "List"
	}
	return out
}

// newUnstructuredScheme builds a scheme where every GVK from the typed scheme
// is rebound to *unstructured.Unstructured (or *unstructured.UnstructuredList
// for List kinds). This mirrors what client-go's NewSimpleDynamicClient does
// internally and is required to make the fake dynamic client work entirely
// with Unstructured objects.
func newUnstructuredScheme(typed *runtime.Scheme) *runtime.Scheme {
	s := runtime.NewScheme()
	for gvk := range typed.AllKnownTypes() {
		if s.Recognizes(gvk) {
			continue
		}
		registerUnstructuredGVK(s, gvk)
	}
	return s
}

// registerUnstructuredGVK adds a GVK (and its corresponding "List" kind)
// to the scheme as unstructured types.
func registerUnstructuredGVK(s *runtime.Scheme, gvk schema.GroupVersionKind) {
	if !s.Recognizes(gvk) {
		if isListKind(gvk.Kind) {
			s.AddKnownTypeWithName(gvk, &unstructured.UnstructuredList{})
		} else {
			s.AddKnownTypeWithName(gvk, &unstructured.Unstructured{})
		}
	}
	if !isListKind(gvk.Kind) {
		listGVK := schema.GroupVersionKind{Group: gvk.Group, Version: gvk.Version, Kind: gvk.Kind + "List"}
		if !s.Recognizes(listGVK) {
			s.AddKnownTypeWithName(listGVK, &unstructured.UnstructuredList{})
		}
	}
}

func isListKind(kind string) bool {
	return len(kind) > 4 && kind[len(kind)-4:] == "List"
}

// KubeClient returns the underlying fake dynamic client. Use it to inspect
// or seed cluster state directly.
func (h *HookExecutionConfig) KubeClient() dynamic.Interface {
	return h.fakeClient
}

// Logger returns the test logger (its output is captured in LoggerOutput).
func (h *HookExecutionConfig) Logger() *log.Logger {
	return h.logger
}

// LoggerOutput returns the buffer of captured log output, useful for assertions.
func (h *HookExecutionConfig) LoggerOutput() *bytes.Buffer {
	return h.loggerOutput
}

// HookError returns the error returned by the hook handler from the most
// recent RunHook call.
func (h *HookExecutionConfig) HookError() error { return h.hookError }

// Snapshots returns the snapshots that were passed to the hook on the most
// recent RunHook call.
func (h *HookExecutionConfig) Snapshots() pkg.Snapshots { return h.snapshots }

// PatchOperations returns the patch operations recorded by the hook during
// the most recent RunHook call.
func (h *HookExecutionConfig) PatchOperations() []pkg.PatchCollectorOperation {
	if h.patchCollector == nil {
		return nil
	}
	return h.patchCollector.Operations()
}

// PatchedOperations returns the typed slice of recorded patch operations
// (one entry per Create/Delete/Patch/... call). This is more convenient
// for assertions than PatchOperations.
func (h *HookExecutionConfig) PatchedOperations() []RecordedPatch {
	if h.patchCollector == nil {
		return nil
	}
	return h.patchCollector.Records()
}

// CollectedMetrics returns the metric operations recorded by the hook during
// the most recent RunHook call.
func (h *HookExecutionConfig) CollectedMetrics() []MetricOperation {
	if h.metricsCollector == nil {
		return nil
	}
	out := h.metricsCollector.CollectedMetrics()
	res := make([]MetricOperation, 0, len(out))
	for _, m := range out {
		res = append(res, MetricOperation{
			Name:   m.Name,
			Group:  m.Group,
			Action: m.Action,
			Value:  m.Value,
			Labels: m.Labels,
		})
	}
	return res
}

// MetricOperation is a stable, framework-friendly view of a metric operation.
type MetricOperation struct {
	Name   string
	Group  string
	Action string
	Value  *float64
	Labels map[string]string
}

// snapshotsMap is the framework's internal Snapshots type. It is exposed via
// the pkg.Snapshots interface; we keep it private to avoid leaking the type.
type snapshotsMap map[string][]pkg.Snapshot

// Get implements pkg.Snapshots.
func (s snapshotsMap) Get(key string) []pkg.Snapshot { return s[key] }

// rawSnapshot holds the raw filtered JSON for a single snapshot.
type rawSnapshot []byte

// UnmarshalTo implements pkg.Snapshot.
func (r rawSnapshot) UnmarshalTo(v any) error {
	return json.Unmarshal(r, v)
}

// String implements pkg.Snapshot.
func (r rawSnapshot) String() string { return string(r) }

// gvrFor returns the GroupVersionResource for the given APIVersion + Kind.
// It consults the explicit gvkToGVR overrides registered via RegisterCRD,
// and falls back to convention-based pluralization via meta.UnsafeGuessKindToResource.
func (h *HookExecutionConfig) gvrFor(apiVersion, kind string) (schema.GroupVersionResource, error) {
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return schema.GroupVersionResource{}, fmt.Errorf("parse apiVersion %q: %w", apiVersion, err)
	}
	gvk := gv.WithKind(kind)
	if gvr, ok := h.gvkToGVR[gvk]; ok {
		return gvr, nil
	}
	gvr, _ := meta.UnsafeGuessKindToResource(gvk)
	return gvr, nil
}

// resolveGVRForKind tries to find a GVR by kind alone, using the scheme and
// any registered CRDs.
func (h *HookExecutionConfig) resolveGVRForKind(kind string) (schema.GroupVersionResource, error) {
	for gvk := range h.scheme.AllKnownTypes() {
		if gvk.Kind == kind {
			gvr, _ := meta.UnsafeGuessKindToResource(gvk)
			return gvr, nil
		}
	}
	for gvk, gvr := range h.gvkToGVR {
		if gvk.Kind == kind {
			return gvr, nil
		}
	}
	return schema.GroupVersionResource{}, fmt.Errorf("kind %q not registered (call RegisterCRD or pass a SchemeBuilder)", kind)
}
