package hook_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/deckhouse/deckhouse/pkg/log"
	bindingcontext "github.com/deckhouse/module-sdk/internal/binding-context"
	"github.com/deckhouse/module-sdk/internal/hook"
	"github.com/deckhouse/module-sdk/pkg"
	"github.com/stretchr/testify/assert"
)

// TODO: make test after transport

func Test_Go_Hook_Execute(t *testing.T) {
	t.Parallel()

	const ()

	type meta struct {
		name    string
		enabled bool
	}

	type fields struct {
		setupHookRequest func(t *testing.T) hook.HookRequest
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
				setupHookRequest: func(t *testing.T) hook.HookRequest {
					hr := NewHookRequestMock(t)

					vals := hr.GetValuesMock.Expect()
					vals.Return(map[string]any{}, nil)

					cvals := hr.GetConfigValuesMock.Expect()
					cvals.Return(map[string]any{}, nil)

					bctxs := hr.GetBindingContextsMock.Expect()
					bctxs.Return([]bindingcontext.BindingContext{}, nil)

					dc := hr.GetDependencyContainerMock.Expect()
					dc.Return(nil)

					return hr
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

			fn := func(_ context.Context, input *pkg.HookInput) error {
				snapshots := input.Snapshots.Get("test_snap")
				for _, snap := range snapshots {
					str := new(string)
					err := snap.UnmarhalTo(str)
					assert.NoError(t, err)

					fmt.Printf("%+v\n", snap.String())
				}

				return fmt.Errorf("sas %+v", snapshots)
			}

			h := hook.NewGoHook(cfg, fn).SetLogger(log.NewNop())

			_, err := h.Execute(context.Background(), tt.fields.setupHookRequest(t))
			assert.NoError(t, err)
		})
	}
}
