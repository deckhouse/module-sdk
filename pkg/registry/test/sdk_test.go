package test

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/registry"
	_ "github.com/deckhouse/module-sdk/pkg/registry/test/simple_module/hooks"
	_ "github.com/deckhouse/module-sdk/pkg/registry/test/simple_module/hooks/001-hook-one"
	_ "github.com/deckhouse/module-sdk/pkg/registry/test/simple_module/hooks/002-hook-two"
)

func Test_HookMetadata_from_runtime(t *testing.T) {
	g := NewWithT(t)

	hookList := registry.Registry().Hooks()
	g.Expect(len(hookList)).Should(Equal(4))

	hooks := map[string]*pkg.Hook{}

	for _, h := range hookList {
		hooks[h.Config.Metadata.Name] = h
		fmt.Println(h.Config.Metadata.Name)
	}

	hm, ok := hooks["001-hook-one/main"]
	g.Expect(ok).To(BeTrue(), "hook-one/main.go should be registered")
	g.Expect(hm.Config.Metadata.Path).To(Equal("001-hook-one/"))

	hm, ok = hooks["002-hook-two/main"]
	g.Expect(ok).To(BeTrue(), "hook-two/main.go should be registered")
	g.Expect(hm.Config.Metadata.Path).To(Equal("002-hook-two/"))

	hm, ok = hooks["first-hook"]
	g.Expect(ok).To(BeTrue(), "first-hook.go should be registered")
	g.Expect(hm.Config.Metadata.Path).To(Equal(""))

	hm, ok = hooks["second-hook"]
	g.Expect(ok).To(BeTrue(), "second-hook.go should be registered")
	g.Expect(hm.Config.Metadata.Path).To(Equal(""))
}
