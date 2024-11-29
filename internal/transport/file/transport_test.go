package file_test

import (
	"testing"

	"github.com/deckhouse/deckhouse/pkg/log"
	"github.com/deckhouse/module-sdk/internal/transport/file"
)

func Test_Request(t *testing.T) {
	t.Parallel()

	const ()

	type meta struct {
		name    string
		enabled bool
	}

	type fields struct {
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
			fields: fields{},
			args:   args{},
			wants:  wants{},
		},
	}

	for _, tt := range tests {
		if !tt.meta.enabled {
			continue
		}

		t.Run(tt.meta.name, func(t *testing.T) {
			t.Parallel()

			tr := file.NewTransport(file.Config{}, "hook-name", nil, log.NewNop())
			req := tr.NewRequest()
			_ = req
		})
	}
}
