package hook_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/deckhouse/deckhouse/pkg/log"

	bindingcontext "github.com/deckhouse/module-sdk/internal/binding-context"
	"github.com/deckhouse/module-sdk/internal/hook"
	"github.com/deckhouse/module-sdk/pkg"
)

func Test_Go_Hook_Execute(t *testing.T) {
	t.Parallel()

	type meta struct {
		name    string
		enabled bool
	}

	type fields struct {
		setupHookRequest       func(t *testing.T) hook.HookRequest
		setupHookReconcileFunc func(t *testing.T) func(ctx context.Context, input *pkg.HookInput) error
	}

	type args struct {
	}

	type wants struct {
		err string
	}

	tests := []struct {
		meta   meta
		fields fields
		args   args
		wants  wants
	}{
		{
			meta: meta{
				name:    "binding contexts contains snapshot with objects",
				enabled: true,
			},
			fields: fields{
				setupHookRequest: func(t *testing.T) hook.HookRequest {
					hr := NewHookRequestMock(t)

					vals := hr.GetValuesMock.Expect()
					vals.Return(map[string]any{}, nil)

					cvals := hr.GetConfigValuesMock.Expect()
					cvals.Return(map[string]any{}, nil)

					bcs := []bindingcontext.BindingContext{
						{
							Snapshots: map[string]bindingcontext.ObjectAndFilterResults{
								"some_data": {
									{
										Object: []byte(`{"name":"stub"}`),
									},
									{
										Object: []byte(`{"name_2":"stub_2"}`),
									},
								},
							},
						},
					}
					bctxs := hr.GetBindingContextsMock.Expect()
					bctxs.Return(bcs, nil)

					dc := hr.GetDependencyContainerMock.Expect()
					dc.Return(nil)

					return hr
				},
				setupHookReconcileFunc: func(t *testing.T) func(ctx context.Context, input *pkg.HookInput) error {
					return func(_ context.Context, input *pkg.HookInput) error {
						data := input.Snapshots.Get("some_data")
						assert.Equal(t, 2, len(data))
						assert.Equal(t, `{"name":"stub"}`, data[0].String())
						assert.Equal(t, `{"name_2":"stub_2"}`, data[1].String())

						return nil
					}
				},
			},
			args:  args{},
			wants: wants{},
		},
		{
			meta: meta{
				name:    "binding contexts contains snapshot with filter results",
				enabled: true,
			},
			fields: fields{
				setupHookRequest: func(t *testing.T) hook.HookRequest {
					hr := NewHookRequestMock(t)

					vals := hr.GetValuesMock.Expect()
					vals.Return(map[string]any{}, nil)

					cvals := hr.GetConfigValuesMock.Expect()
					cvals.Return(map[string]any{}, nil)

					bcs := []bindingcontext.BindingContext{
						{
							Snapshots: map[string]bindingcontext.ObjectAndFilterResults{
								"some_data": {
									{
										FilterResult: []byte(`{"name":"stub"}`),
									},
									{
										FilterResult: []byte(`{"name_2":"stub_2"}`),
									},
								},
							},
						},
					}
					bctxs := hr.GetBindingContextsMock.Expect()
					bctxs.Return(bcs, nil)

					dc := hr.GetDependencyContainerMock.Expect()
					dc.Return(nil)

					return hr
				},
				setupHookReconcileFunc: func(t *testing.T) func(ctx context.Context, input *pkg.HookInput) error {
					return func(_ context.Context, input *pkg.HookInput) error {
						data := input.Snapshots.Get("some_data")
						assert.Equal(t, 2, len(data))
						assert.Equal(t, `{"name":"stub"}`, data[0].String())
						assert.Equal(t, `{"name_2":"stub_2"}`, data[1].String())

						return nil
					}
				},
			},
			args:  args{},
			wants: wants{},
		},
		{
			meta: meta{
				name:    "binding contexts contains snapshot with object and filter result",
				enabled: true,
			},
			fields: fields{
				setupHookRequest: func(t *testing.T) hook.HookRequest {
					hr := NewHookRequestMock(t)

					vals := hr.GetValuesMock.Expect()
					vals.Return(map[string]any{}, nil)

					cvals := hr.GetConfigValuesMock.Expect()
					cvals.Return(map[string]any{}, nil)

					bcs := []bindingcontext.BindingContext{
						{
							Snapshots: map[string]bindingcontext.ObjectAndFilterResults{
								"some_data": {
									{
										Object:       []byte(`{"name":"wrong_answer"}`),
										FilterResult: []byte(`{"name":"correct_answer"}`),
									},
									{
										FilterResult: []byte(`{"name_2":"stub_2"}`),
									},
								},
							},
						},
					}
					bctxs := hr.GetBindingContextsMock.Expect()
					bctxs.Return(bcs, nil)

					dc := hr.GetDependencyContainerMock.Expect()
					dc.Return(nil)

					return hr
				},
				setupHookReconcileFunc: func(t *testing.T) func(ctx context.Context, input *pkg.HookInput) error {
					return func(_ context.Context, input *pkg.HookInput) error {
						data := input.Snapshots.Get("some_data")
						assert.Equal(t, 2, len(data))
						assert.Equal(t, `{"name":"correct_answer"}`, data[0].String())
						assert.Equal(t, `{"name_2":"stub_2"}`, data[1].String())

						return nil
					}
				},
			},
			args:  args{},
			wants: wants{},
		},
		{
			meta: meta{
				name:    "get values error",
				enabled: true,
			},
			fields: fields{
				setupHookRequest: func(t *testing.T) hook.HookRequest {
					hr := NewHookRequestMock(t)

					vals := hr.GetValuesMock.Expect()
					vals.Return(nil, errors.New("error"))

					return hr
				},
				setupHookReconcileFunc: func(_ *testing.T) func(ctx context.Context, input *pkg.HookInput) error {
					return func(_ context.Context, _ *pkg.HookInput) error {
						return nil
					}
				},
			},
			args: args{},
			wants: wants{
				err: "get values: error",
			},
		},
		{
			meta: meta{
				name:    "get patchable values error",
				enabled: true,
			},
			fields: fields{
				setupHookRequest: func(t *testing.T) hook.HookRequest {
					hr := NewHookRequestMock(t)

					vals := hr.GetValuesMock.Expect()
					vals.Return(map[string]any{
						"func-to-crush-marshal": func() {},
					}, nil)

					return hr
				},
				setupHookReconcileFunc: func(_ *testing.T) func(ctx context.Context, input *pkg.HookInput) error {
					return func(_ context.Context, _ *pkg.HookInput) error {
						return nil
					}
				},
			},
			args: args{},
			wants: wants{
				err: `get patchable values: json: unsupported type: func()`,
			},
		},
		{
			meta: meta{
				name:    "get config values error",
				enabled: true,
			},
			fields: fields{
				setupHookRequest: func(t *testing.T) hook.HookRequest {
					hr := NewHookRequestMock(t)

					vals := hr.GetValuesMock.Expect()
					vals.Return(nil, nil)

					cvals := hr.GetConfigValuesMock.Expect()
					cvals.Return(nil, errors.New("error"))

					return hr
				},
				setupHookReconcileFunc: func(_ *testing.T) func(ctx context.Context, input *pkg.HookInput) error {
					return func(_ context.Context, _ *pkg.HookInput) error {
						return nil
					}
				},
			},
			args: args{},
			wants: wants{
				err: "get config values: error",
			},
		},
		{
			meta: meta{
				name:    "get patchable config values error",
				enabled: true,
			},
			fields: fields{
				setupHookRequest: func(t *testing.T) hook.HookRequest {
					hr := NewHookRequestMock(t)

					vals := hr.GetValuesMock.Expect()
					vals.Return(nil, nil)

					cvals := hr.GetConfigValuesMock.Expect()
					cvals.Return(map[string]any{
						"func-to-crush-marshal": func() {},
					}, nil)

					return hr
				},
				setupHookReconcileFunc: func(_ *testing.T) func(ctx context.Context, input *pkg.HookInput) error {
					return func(_ context.Context, _ *pkg.HookInput) error {
						return nil
					}
				},
			},
			args: args{},
			wants: wants{
				err: "get patchable config values: json: unsupported type: func()",
			},
		},
		{
			meta: meta{
				name:    "get binding context do not stop flow",
				enabled: true,
			},
			fields: fields{
				setupHookRequest: func(t *testing.T) hook.HookRequest {
					hr := NewHookRequestMock(t)

					vals := hr.GetValuesMock.Expect()
					vals.Return(nil, nil)

					cvals := hr.GetConfigValuesMock.Expect()
					cvals.Return(nil, nil)

					bctxs := hr.GetBindingContextsMock.Expect()
					bctxs.Return(nil, errors.New("error"))

					dc := hr.GetDependencyContainerMock.Expect()
					dc.Return(nil)

					return hr
				},
				setupHookReconcileFunc: func(_ *testing.T) func(ctx context.Context, input *pkg.HookInput) error {
					return func(_ context.Context, _ *pkg.HookInput) error {
						return nil
					}
				},
			},
			args: args{},
			wants: wants{
				err: "",
			},
		},
		{
			meta: meta{
				name:    "hook reconcile func returns error",
				enabled: true,
			},
			fields: fields{
				setupHookRequest: func(t *testing.T) hook.HookRequest {
					hr := NewHookRequestMock(t)

					vals := hr.GetValuesMock.Expect()
					vals.Return(nil, nil)

					cvals := hr.GetConfigValuesMock.Expect()
					cvals.Return(nil, nil)

					bctxs := hr.GetBindingContextsMock.Expect()
					bctxs.Return(nil, nil)

					dc := hr.GetDependencyContainerMock.Expect()
					dc.Return(nil)

					return hr
				},
				setupHookReconcileFunc: func(_ *testing.T) func(ctx context.Context, input *pkg.HookInput) error {
					return func(_ context.Context, _ *pkg.HookInput) error {
						return errors.New("error")
					}
				},
			},
			args: args{},
			wants: wants{
				err: "hook reconcile func: error",
			},
		},
	}

	for _, tt := range tests {
		if !tt.meta.enabled {
			continue
		}

		t.Run(tt.meta.name, func(t *testing.T) {
			t.Parallel()

			cfg := &pkg.HookConfig{}

			h := hook.NewHook(cfg, tt.fields.setupHookReconcileFunc(t)).SetLogger(log.NewNop())

			_, err := h.Execute(context.Background(), tt.fields.setupHookRequest(t))
			if tt.wants.err != "" {
				assert.Contains(t, err.Error(), tt.wants.err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
