package helpers

import (
	"bytes"
	"testing"

	"github.com/deckhouse/deckhouse/pkg/log"

	"github.com/deckhouse/module-sdk/internal/metric"
	"github.com/deckhouse/module-sdk/pkg"
)

// InputBuilder is a fluent builder for *pkg.HookInput. It bundles together
// the most common boilerplate of unit tests:
//
//   - StaticSnapshots seeded with JSON / YAML / Go-value snapshots
//   - real PatchableValuesCollector for Values / ConfigValues
//   - a RecordingPatchCollector for assertions
//   - a real metric.Collector
//   - a logger that can either be silent or write to a captured buffer
//
// Anything you don't configure has a sensible default, so a zero-config
// builder still produces a usable HookInput.
type InputBuilder struct {
	tb testing.TB

	snapshots StaticSnapshots
	values    pkg.PatchableValuesCollector
	config    pkg.PatchableValuesCollector
	patch     pkg.PatchCollector
	metrics   pkg.MetricsCollector
	dc        pkg.DependencyContainer

	logger    pkg.Logger
	logBuffer *bytes.Buffer

	// Set when the user explicitly provided a custom collector,
	// so we keep the typed reference for accessor methods.
	recordingPC *RecordingPatchCollector
}

// NewInputBuilder returns a builder bound to the given testing.TB. The TB
// is currently used for failing the test if an internal helper call fails;
// pass nil from non-test code.
func NewInputBuilder(tb testing.TB) *InputBuilder {
	return &InputBuilder{
		tb:        tb,
		snapshots: NewSnapshots(),
	}
}

// WithSnapshot adds one or more snapshots under the given key. May be
// called multiple times; snapshots are appended.
func (b *InputBuilder) WithSnapshot(key string, snaps ...pkg.Snapshot) *InputBuilder {
	b.snapshots.Add(key, snaps...)
	return b
}

// WithSnapshots replaces the entire snapshots map with the provided one.
// Use this when you have an already-constructed StaticSnapshots.
func (b *InputBuilder) WithSnapshots(s StaticSnapshots) *InputBuilder {
	b.snapshots = s
	return b
}

// WithValues replaces the values collector with the given one. By default,
// the builder constructs an empty PatchableValuesCollector lazily on Build.
func (b *InputBuilder) WithValues(v pkg.PatchableValuesCollector) *InputBuilder {
	b.values = v
	return b
}

// WithValuesJSON seeds the values collector from a JSON string.
func (b *InputBuilder) WithValuesJSON(raw string) *InputBuilder {
	b.values = NewValuesFromJSON(raw)
	return b
}

// WithValuesYAML seeds the values collector from a YAML string.
func (b *InputBuilder) WithValuesYAML(raw string) *InputBuilder {
	b.values = NewValuesFromYAML(raw)
	return b
}

// WithValuesMap seeds the values collector from a Go map.
func (b *InputBuilder) WithValuesMap(m map[string]any) *InputBuilder {
	b.values = NewValues(m)
	return b
}

// WithConfigValues replaces the config values collector.
func (b *InputBuilder) WithConfigValues(v pkg.PatchableValuesCollector) *InputBuilder {
	b.config = v
	return b
}

// WithConfigValuesJSON seeds the config values collector from JSON.
func (b *InputBuilder) WithConfigValuesJSON(raw string) *InputBuilder {
	b.config = NewValuesFromJSON(raw)
	return b
}

// WithConfigValuesYAML seeds the config values collector from YAML.
func (b *InputBuilder) WithConfigValuesYAML(raw string) *InputBuilder {
	b.config = NewValuesFromYAML(raw)
	return b
}

// WithPatchCollector replaces the patch collector with a custom one (for
// example, a minimock-generated mock).
func (b *InputBuilder) WithPatchCollector(c pkg.PatchCollector) *InputBuilder {
	b.patch = c
	b.recordingPC = nil
	return b
}

// WithRecordingPatchCollector wires a fresh RecordingPatchCollector into
// the input. The collector itself is returned by RecordingPatchCollector
// after Build.
func (b *InputBuilder) WithRecordingPatchCollector() *InputBuilder {
	b.recordingPC = NewRecordingPatchCollector()
	b.patch = b.recordingPC
	return b
}

// WithMetricsCollector replaces the metrics collector.
func (b *InputBuilder) WithMetricsCollector(c pkg.MetricsCollector) *InputBuilder {
	b.metrics = c
	return b
}

// WithDependencyContainer replaces the dependency container.
func (b *InputBuilder) WithDependencyContainer(dc pkg.DependencyContainer) *InputBuilder {
	b.dc = dc
	return b
}

// WithLogger replaces the logger used by the hook.
func (b *InputBuilder) WithLogger(l pkg.Logger) *InputBuilder {
	b.logger = l
	b.logBuffer = nil
	return b
}

// WithCapturedLogger installs a *log.Logger writing into a private buffer.
// The buffer is accessible via LogBuffer after Build.
func (b *InputBuilder) WithCapturedLogger() *InputBuilder {
	buf := bytes.NewBuffer(nil)
	b.logger = log.NewLogger(
		log.WithLevel(log.LevelDebug.Level()),
		log.WithOutput(buf),
	)
	b.logBuffer = buf
	return b
}

// Build assembles the *pkg.HookInput. Calling Build twice is allowed and
// returns the same shared values / patch collector references.
func (b *InputBuilder) Build() *pkg.HookInput {
	if b.values == nil {
		b.values = NewValues(nil)
	}
	if b.config == nil {
		b.config = NewValues(nil)
	}
	if b.patch == nil {
		b.recordingPC = NewRecordingPatchCollector()
		b.patch = b.recordingPC
	}
	if b.metrics == nil {
		b.metrics = metric.NewCollector()
	}
	if b.logger == nil {
		b.logger = log.NewNop()
	}

	return &pkg.HookInput{
		Snapshots:        b.snapshots,
		Values:           b.values,
		ConfigValues:     b.config,
		PatchCollector:   b.patch,
		MetricsCollector: b.metrics,
		DC:               b.dc,
		Logger:           b.logger,
	}
}

// Snapshots returns the StaticSnapshots backing the input. Useful if you
// want to add more snapshots after the input is built.
func (b *InputBuilder) Snapshots() StaticSnapshots { return b.snapshots }

// Values returns the values collector that will be used by the input.
func (b *InputBuilder) Values() pkg.PatchableValuesCollector { return b.values }

// ConfigValues returns the config values collector that will be used.
func (b *InputBuilder) ConfigValues() pkg.PatchableValuesCollector { return b.config }

// RecordingPatchCollector returns the typed RecordingPatchCollector if
// one was attached via WithRecordingPatchCollector or implicitly by Build.
// Returns nil when the user supplied a different PatchCollector.
func (b *InputBuilder) RecordingPatchCollector() *RecordingPatchCollector {
	return b.recordingPC
}

// LogBuffer returns the buffer behind WithCapturedLogger or nil.
func (b *InputBuilder) LogBuffer() *bytes.Buffer { return b.logBuffer }
