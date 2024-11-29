package hook_test

import (
	"fmt"
	"testing"

	bindingcontext "github.com/deckhouse/module-sdk/internal/binding-context"
	"github.com/deckhouse/module-sdk/internal/hook"
	"github.com/deckhouse/module-sdk/pkg"
	"github.com/stretchr/testify/assert"
)

var _ hook.HookRequest = (*HookRequest)(nil)

type HookRequest struct {
}

func (req *HookRequest) GetValues() (map[string]any, error) {
	return make(map[string]any, 1), nil
}

func (req *HookRequest) GetConfigValues() (map[string]any, error) {
	return make(map[string]any, 1), nil
}

func (req *HookRequest) GetBindingContexts() ([]bindingcontext.BindingContext, error) {
	return []bindingcontext.BindingContext{
		{
			Snapshots: map[string]bindingcontext.ObjectAndFilterResults{
				"test_snap": {
					{
						Object:       nil,
						FilterResult: nil,
					},
				},
			},
		},
	}, nil
}

func (req *HookRequest) GetDependencyContainer() pkg.DependencyContainer {
	return nil
}

func Test_Go_Hook_Execute(t *testing.T) {
	t.Parallel()

	const ()

	type meta struct {
		name    string
		enabled bool
	}

	type fields struct {
		setupHookRequest func() hook.HookRequest
	}

	type args struct {
	}

	type wants struct {
	}

	tests := []struct {
		meta   meta
		fields fields
		args   args
		wants  wants
	}{
		{
			meta: meta{
				name:    "logger default options is level info and add source false",
				enabled: true,
			},
			fields: fields{
				setupHookRequest: func() hook.HookRequest {
					return &HookRequest{}
				},
			},
			args:  args{},
			wants: wants{},
		},
	}

	for _, tt := range tests {
		if !tt.meta.enabled {
			continue
		}

		t.Run(tt.meta.name, func(t *testing.T) {
			t.Parallel()

			cfg := &pkg.HookConfig{}

			fn := func(input *pkg.HookInput) error {
				snapshots := input.Snapshots.Get("test_snap")
				return fmt.Errorf("sas %+v", snapshots)
			}
			h := hook.NewGoHook(cfg, fn)
			_, err := h.Execute(tt.fields.setupHookRequest())
			assert.NoError(t, err)
		})
	}
}
