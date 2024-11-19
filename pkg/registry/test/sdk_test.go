package test

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"

	gohook "github.com/deckhouse/module-sdk/pkg/hook"
	"github.com/deckhouse/module-sdk/pkg/registry"
	_ "github.com/deckhouse/module-sdk/pkg/registry/test/simple_module/hooks/go-hooks/001-hook-one"
	_ "github.com/deckhouse/module-sdk/pkg/registry/test/simple_module/hooks/go-hooks/002-hook-two/level1/sublevel"
)

func Test_HookMetadata_from_runtime(t *testing.T) {
	g := NewWithT(t)

	hookList := registry.Registry().Hooks()
	g.Expect(len(hookList)).Should(Equal(2))

	hooks := map[string]*gohook.GoHook{}

	for _, h := range hookList {
		hooks[h.GetName()] = h
		fmt.Println(h.GetName())
	}

	hm, ok := hooks["001-hook-one"]
	g.Expect(ok).To(BeTrue(), "module-one-hook.go should be registered")
	g.Expect(hm.GetPath()).To(Equal("001-hook-one/module-one-hook.go"))

	hm, ok = hooks["002-hook-two"]
	g.Expect(ok).To(BeTrue(), "sub-sub-hook.go should be registered")
	g.Expect(hm.GetPath()).To(Equal("002-hook-two/level1/sublevel/sub-sub-hook.go"))
}
