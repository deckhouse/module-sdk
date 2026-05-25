package framework

import (
	"context"

	"github.com/deckhouse/module-sdk/internal/metric"
	"github.com/deckhouse/module-sdk/pkg"
)

// RunHook executes the registered hook handler against the current state.
//
// The framework:
//  1. Generates snapshots from the fake cluster according to the hook's
//     KubernetesConfig bindings.
//  2. Builds a real pkg.HookInput backed by working values stores, a recording
//     PatchCollector, and a Collector for metrics.
//  3. Invokes the hook handler with that input.
//  4. Applies the values patches produced by the hook to the values store.
//  5. Replays the recorded patch operations against the fake cluster.
//
// After RunHook, use HookError, ValuesGet, ConfigValuesGet, KubernetesResource,
// PatchedOperations and CollectedMetrics to assert behaviour.
func (h *HookExecutionConfig) RunHook() {
	h.t.Helper()
	h.RunHookCtx(context.Background())
}

// RunHookCtx is like RunHook but accepts an explicit context.
func (h *HookExecutionConfig) RunHookCtx(ctx context.Context) {
	h.t.Helper()
	h.hookError = nil

	snaps, err := h.generateSnapshots(ctx)
	if err != nil {
		h.t.Fatalf("framework: generate snapshots: %v", err)
	}
	h.snapshots = snaps

	patchableValues, err := patchableValuesFor(h.values)
	if err != nil {
		h.t.Fatalf("framework: build patchable values: %v", err)
	}
	patchableConfigValues, err := patchableValuesFor(h.configValues)
	if err != nil {
		h.t.Fatalf("framework: build patchable config values: %v", err)
	}

	h.patchCollector = newRecordingPatchCollector()
	h.metricsCollector = metric.NewCollector()

	if h.dc == nil {
		h.dc = newFrameworkDC(h.fakeClient, h.scheme)
	}

	input := &pkg.HookInput{
		Snapshots:        h.snapshots,
		Values:           patchableValues,
		ConfigValues:     patchableConfigValues,
		PatchCollector:   h.patchCollector,
		MetricsCollector: h.metricsCollector,
		DC:               h.dc,
		Logger:           h.logger,
	}

	h.hookError = h.hookHandler(ctx, input)

	// Always merge values patches so callers can assert both happy and error
	// paths.
	if err := h.values.applyPatchOperations(patchableValues.GetPatches()); err != nil {
		h.t.Fatalf("framework: apply values patches: %v", err)
	}
	if err := h.configValues.applyPatchOperations(patchableConfigValues.GetPatches()); err != nil {
		h.t.Fatalf("framework: apply config values patches: %v", err)
	}

	if h.hookError == nil {
		if err := h.applyPatchesToCluster(); err != nil {
			h.t.Fatalf("framework: apply collected patches: %v", err)
		}
	}
}
