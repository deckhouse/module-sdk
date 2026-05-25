package hookinfolder_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/deckhouse/module-sdk/pkg/registry"
)

func TestRegisteredHookConfigs_AreValid(t *testing.T) {
	for _, hook := range registry.Registry().ModuleHooks() {
		assert.NoError(t, hook.Config.Validate(), "hook config must be valid")
	}
}
