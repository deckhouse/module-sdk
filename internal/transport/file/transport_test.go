package file_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/deckhouse/deckhouse/pkg/log"
	bindingcontext "github.com/deckhouse/module-sdk/internal/binding-context"
	"github.com/deckhouse/module-sdk/internal/transport/file"
	"github.com/stretchr/testify/assert"
)

const (
	valuesFilePath          = "values.json"
	configValuesFilePath    = "config_values.json"
	bindingContextsFilePath = "binding_contexts.json"
)

type File struct {
	Name    string
	Content string
}

func Test_Request_GetBindingContexts(t *testing.T) {
	t.Parallel()

	const (
		bindingContextObject = `
[
	{
	"binding": "node_roles",
	"groupName": "policy",
	"snapshots": {
		"node_roles": [
		{
			"object": {
				"apiVersion": "v1",
				"metadata": {
					"name": "test-object"
				}
			}
		},
		{
			"filterResult": {
				"apiVersion": "v1",
				"metadata": {
					"name": "test-filter-result"
				}
			}
		}
		]
	},
	"type": "Group"
	}
]`
		bindingContextEmptySnapshotsObjects = `
[
  {
    "binding": "node_roles",
    "groupName": "policy",
    "snapshots": {
      "node_roles": [
        {},
        {}
      ]
    },
    "type": "Group"
  }
]`
		bindingContextEmptyObjectAndFilterResult = `
[
  {
    "binding": "node_roles",
    "groupName": "policy",
    "snapshots": {
      "node_roles": [
        {
			"object":"{}"
	  	},
        {
			"filterResult":"{}"
		},
        {
			"object":{}
	  	},
        {
			"filterResult":{}
		}
      ]
    },
    "type": "Group"
  }
]`
	)

	type meta struct {
		name    string
		enabled bool
	}

	type fields struct {
	}

	type args struct {
		filesContent map[string]File
	}

	type wants struct {
		bcs []bindingcontext.BindingContext
	}

	tests := []struct {
		meta   meta
		fields fields
		args   args
		wants  wants
	}{
		{
			meta: meta{
				name:    "successfull get binding context",
				enabled: true,
			},
			fields: fields{},
			args: args{
				filesContent: map[string]File{
					valuesFilePath: {
						Name:    generateFileNameWithTs(valuesFilePath),
						Content: "",
					},
					configValuesFilePath: {
						Name:    generateFileNameWithTs(configValuesFilePath),
						Content: "",
					},
					bindingContextsFilePath: {
						Name:    generateFileNameWithTs(bindingContextsFilePath),
						Content: bindingContextObject,
					},
				},
			},
			wants: wants{
				bcs: []bindingcontext.BindingContext{
					{
						Binding: "node_roles",
						Type:    "Group",
						Snapshots: map[string]bindingcontext.ObjectAndFilterResults{
							"node_roles": {
								{
									Object: []byte(`{
				"apiVersion": "v1",
				"metadata": {
					"name": "test-object"
				}
			}`),
								},
								{
									FilterResult: []byte(`{
				"apiVersion": "v1",
				"metadata": {
					"name": "test-filter-result"
				}
			}`),
								},
							},
						},
					},
				},
			},
		},
		{
			meta: meta{
				name:    "empty binding contexts snapshot objects",
				enabled: true,
			},
			fields: fields{},
			args: args{
				filesContent: map[string]File{
					valuesFilePath: {
						Name:    generateFileNameWithTs(valuesFilePath),
						Content: "",
					},
					configValuesFilePath: {
						Name:    generateFileNameWithTs(configValuesFilePath),
						Content: "",
					},
					bindingContextsFilePath: {
						Name:    generateFileNameWithTs(bindingContextsFilePath),
						Content: bindingContextEmptySnapshotsObjects,
					},
				},
			},
			wants: wants{
				bcs: []bindingcontext.BindingContext{
					{
						Binding: "node_roles",
						Type:    "Group",
						Snapshots: map[string]bindingcontext.ObjectAndFilterResults{
							"node_roles": {
								{}, {},
							},
						},
					},
				},
			},
		},
		{
			meta: meta{
				name:    "empty binding contexts objects and filter result must be nil",
				enabled: true,
			},
			fields: fields{},
			args: args{
				filesContent: map[string]File{
					valuesFilePath: {
						Name:    generateFileNameWithTs(valuesFilePath),
						Content: "",
					},
					configValuesFilePath: {
						Name:    generateFileNameWithTs(configValuesFilePath),
						Content: "",
					},
					bindingContextsFilePath: {
						Name:    generateFileNameWithTs(bindingContextsFilePath),
						Content: bindingContextEmptyObjectAndFilterResult,
					},
				},
			},
			wants: wants{
				bcs: []bindingcontext.BindingContext{
					{
						Binding: "node_roles",
						Type:    "Group",
						Snapshots: map[string]bindingcontext.ObjectAndFilterResults{
							"node_roles": {
								{}, {},
								{}, {},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		if !tt.meta.enabled {
			continue
		}

		t.Run(tt.meta.name, func(t *testing.T) {
			t.Parallel()

			// create files with info
			for _, file := range tt.args.filesContent {
				err := os.WriteFile(file.Name, []byte(file.Content), 0777)
				assert.NoError(t, err)
			}

			tcfg := &file.Config{
				ValuesPath:         tt.args.filesContent[valuesFilePath].Name,
				ConfigValuesPath:   tt.args.filesContent[configValuesFilePath].Name,
				BindingContextPath: tt.args.filesContent[bindingContextsFilePath].Name,
			}

			tr := file.NewTransport(tcfg, "hook-name", nil, log.NewNop())
			req := tr.NewRequest()
			bcs, err := req.GetBindingContexts()
			assert.NoError(t, err)
			assert.Equal(t, tt.wants.bcs, bcs)

			// cleanup
			for _, file := range tt.args.filesContent {
				err := os.Remove(file.Name)
				assert.NoError(t, err)
			}
		})
	}
}

func generateFileNameWithTs(defaultPath string) string {
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), defaultPath)
}
