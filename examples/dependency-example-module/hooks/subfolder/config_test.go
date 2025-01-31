package hookinfolder_test

import (
	"github.com/deckhouse/module-sdk/pkg/registry"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("validate hooks config", func() {
	It("hook configs must be valid", func() {
		hooks := registry.Registry().Hooks()
		for _, hook := range hooks {
			Expect(hook.Config.Validate()).ShouldNot(HaveOccurred())
		}
	})
})
