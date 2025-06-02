package file_test

import (
	"fmt"
	"io/fs"
	"os"
	"testing"

	uuid "github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/deckhouse/deckhouse/pkg/log"

	bindingcontext "github.com/deckhouse/module-sdk/internal/binding-context"
	fileTransport "github.com/deckhouse/module-sdk/internal/transport/file"
)

const (
	valuesFilePath          = "values.json"
	configValuesFilePath    = "config_values.json"
	bindingContextsFilePath = "binding_contexts.json"
)

type file struct {
	Name    string
	Content string
}

func Test_RequestGetValues(t *testing.T) {
	t.Parallel()

	const (
		values  = `{"some-module":{},"global":{"modules":{"publicDomainTemplate":"%s.com"}}}`
		badJSON = `{{{{{`
	)

	type meta struct {
		name    string
		enabled bool
	}

	type fields struct {
	}

	type args struct {
		filesContent     map[string]file
		filesPermissions fs.FileMode
	}

	type wants struct {
		configValues map[string]any
		err          string
	}

	tests := []struct {
		meta   meta
		fields fields
		args   args
		wants  wants
	}{
		{
			meta: meta{
				name:    "successful get values",
				enabled: true,
			},
			fields: fields{},
			args: args{
				filesContent: map[string]file{
					valuesFilePath: {
						Name:    generateFileNameWithTS(valuesFilePath),
						Content: values,
					},
				},
				filesPermissions: 0777,
			},
			wants: wants{
				configValues: map[string]any{
					"global": map[string]any{
						"modules": map[string]any{
							"publicDomainTemplate": "%s.com"},
					},
					"some-module": map[string]any{},
				},
			},
		},
		{
			meta: meta{
				name:    "no file",
				enabled: true,
			},
			fields: fields{},
			args: args{
				filesContent:     map[string]file{},
				filesPermissions: 0777,
			},
			wants: wants{
				configValues: nil,
			},
		},
		{
			meta: meta{
				name:    "can not open file",
				enabled: true,
			},
			fields: fields{},
			args: args{
				filesContent: map[string]file{
					valuesFilePath: {
						Name:    generateFileNameWithTS(valuesFilePath),
						Content: values,
					},
				},
				filesPermissions: 0000,
			},
			wants: wants{
				configValues: nil,
				err:          `load values from file:`,
			},
		},
		{
			meta: meta{
				name:    "bad json",
				enabled: true,
			},
			fields: fields{},
			args: args{
				filesContent: map[string]file{
					valuesFilePath: {
						Name:    generateFileNameWithTS(valuesFilePath),
						Content: badJSON,
					},
				},
				filesPermissions: 0777,
			},
			wants: wants{
				configValues: nil,
				err:          "load values from file: bad values data: error converting YAML to JSON: yaml: line 1: did not find expected node content\n{{{{{",
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
			for _, f := range tt.args.filesContent {
				err := os.WriteFile(f.Name, []byte(f.Content), tt.args.filesPermissions)
				assert.NoError(t, err)
			}

			tcfg := &fileTransport.Config{
				ValuesPath: tt.args.filesContent[valuesFilePath].Name,
			}

			tr := fileTransport.NewTransport(tcfg, "hook-name", nil, log.NewNop())
			req := tr.NewRequest()
			bcs, err := req.GetValues()
			// cleanup
			for _, f := range tt.args.filesContent {
				_ = os.Remove(f.Name)
			}

			if tt.wants.err != "" {
				assert.Contains(t, err.Error(), tt.wants.err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wants.configValues, bcs)
		})
	}
}

func Test_RequestGetConfigValues(t *testing.T) {
	t.Parallel()

	const (
		configValues = `{"global":{"modules":{"publicDomainTemplate":"%s.com"}},"some-module":{}}`
		badJSON      = `{{{{{`
	)

	type meta struct {
		name    string
		enabled bool
	}

	type fields struct {
	}

	type args struct {
		filesContent     map[string]file
		filesPermissions fs.FileMode
	}

	type wants struct {
		configValues map[string]any
		err          string
	}

	tests := []struct {
		meta   meta
		fields fields
		args   args
		wants  wants
	}{
		{
			meta: meta{
				name:    "successful get config values",
				enabled: true,
			},
			fields: fields{},
			args: args{
				filesContent: map[string]file{
					configValuesFilePath: {
						Name:    generateFileNameWithTS(configValuesFilePath),
						Content: configValues,
					},
				},
				filesPermissions: 0777,
			},
			wants: wants{
				configValues: map[string]any{
					"global": map[string]any{
						"modules": map[string]any{
							"publicDomainTemplate": "%s.com"},
					},
					"some-module": map[string]any{},
				},
			},
		},
		{
			meta: meta{
				name:    "no file",
				enabled: true,
			},
			fields: fields{},
			args: args{
				filesContent:     map[string]file{},
				filesPermissions: 0777,
			},
			wants: wants{
				configValues: nil,
			},
		},
		{
			meta: meta{
				name:    "can not open file",
				enabled: true,
			},
			fields: fields{},
			args: args{
				filesContent: map[string]file{
					configValuesFilePath: {
						Name:    generateFileNameWithTS(configValuesFilePath),
						Content: configValues,
					},
				},
				filesPermissions: 0000,
			},
			wants: wants{
				configValues: nil,
				err:          `load values from file:`,
			},
		},
		{
			meta: meta{
				name:    "bad json",
				enabled: true,
			},
			fields: fields{},
			args: args{
				filesContent: map[string]file{
					configValuesFilePath: {
						Name:    generateFileNameWithTS(configValuesFilePath),
						Content: badJSON,
					},
				},
				filesPermissions: 0777,
			},
			wants: wants{
				configValues: nil,
				err:          "load values from file: bad values data: error converting YAML to JSON: yaml: line 1: did not find expected node content\n{{{{{",
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
			for _, f := range tt.args.filesContent {
				err := os.WriteFile(f.Name, []byte(f.Content), tt.args.filesPermissions)
				assert.NoError(t, err)
			}

			tcfg := &fileTransport.Config{
				ConfigValuesPath: tt.args.filesContent[configValuesFilePath].Name,
			}

			tr := fileTransport.NewTransport(tcfg, "hook-name", nil, log.NewNop())
			req := tr.NewRequest()
			bcs, err := req.GetConfigValues()
			// cleanup
			for _, f := range tt.args.filesContent {
				_ = os.Remove(f.Name)
			}

			if tt.wants.err != "" {
				assert.Contains(t, err.Error(), tt.wants.err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wants.configValues, bcs)
		})
	}
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
		bindingContextBadJSON               = `{{{{`
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
		filesContent     map[string]file
		filesPermissions fs.FileMode
	}

	type wants struct {
		bcs []bindingcontext.BindingContext
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
				name:    "successful get binding context",
				enabled: true,
			},
			fields: fields{},
			args: args{
				filesContent: map[string]file{
					bindingContextsFilePath: {
						Name:    generateFileNameWithTS(bindingContextsFilePath),
						Content: bindingContextObject,
					},
				},
				filesPermissions: 0777,
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
				filesContent: map[string]file{
					bindingContextsFilePath: {
						Name:    generateFileNameWithTS(bindingContextsFilePath),
						Content: bindingContextEmptySnapshotsObjects,
					},
				},
				filesPermissions: 0777,
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
				filesContent: map[string]file{
					bindingContextsFilePath: {
						Name:    generateFileNameWithTS(bindingContextsFilePath),
						Content: bindingContextEmptyObjectAndFilterResult,
					},
				},
				filesPermissions: 0777,
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
		}, {
			meta: meta{
				name:    "bad json",
				enabled: true,
			},
			fields: fields{},
			args: args{
				filesContent: map[string]file{
					bindingContextsFilePath: {
						Name:    generateFileNameWithTS(bindingContextsFilePath),
						Content: bindingContextBadJSON,
					},
				},
				filesPermissions: 0777,
			},
			wants: wants{
				bcs: nil,
				err: "decode binding context: invalid character '{' looking for beginning of object key string",
			},
		},
		{
			meta: meta{
				name:    "file read error",
				enabled: true,
			},
			fields: fields{},
			args: args{
				filesContent: map[string]file{
					bindingContextsFilePath: {
						Name:    generateFileNameWithTS(bindingContextsFilePath),
						Content: bindingContextObject,
					},
				},
				filesPermissions: 0000,
			},
			wants: wants{
				bcs: nil,
				err: "open binding context file:",
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
			for _, f := range tt.args.filesContent {
				err := os.WriteFile(f.Name, []byte(f.Content), tt.args.filesPermissions)
				assert.NoError(t, err)
			}

			tcfg := &fileTransport.Config{
				BindingContextPath: tt.args.filesContent[bindingContextsFilePath].Name,
			}

			tr := fileTransport.NewTransport(tcfg, "hook-name", nil, log.NewNop())
			req := tr.NewRequest()
			bcs, err := req.GetBindingContexts()

			// cleanup
			for _, f := range tt.args.filesContent {
				_ = os.Remove(f.Name)
			}

			if tt.wants.err != "" {
				assert.Contains(t, err.Error(), tt.wants.err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wants.bcs, bcs)
		})
	}
}

func generateFileNameWithTS(defaultPath string) string {
	return fmt.Sprintf("%s-%s", uuid.New().String(), defaultPath)
}
