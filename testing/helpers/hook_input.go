package helpers

import (
	"testing"

	"github.com/deckhouse/deckhouse/pkg/log"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/testing/mock"
)

func NewHookInput(t *testing.T) *pkg.HookInput {
	return &pkg.HookInput{
		Snapshots:        mock.NewSnapshotsMock(t),
		Values:           mock.NewOutputPatchableValuesCollectorMock(t),
		ConfigValues:     mock.NewOutputPatchableValuesCollectorMock(t),
		PatchCollector:   mock.NewPatchCollectorMock(t),
		MetricsCollector: mock.NewMetricsCollectorMock(t),
		DC:               mock.NewDependencyContainerMock(t),
		Logger:           log.NewNop(),
	}
}
