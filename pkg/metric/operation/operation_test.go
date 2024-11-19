package operation

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ValidateOperations(t *testing.T) {
	var stubFloat float64 = 1.0

	tests := []struct {
		name string
		op   Operation
		err  error
	}{
		{
			"simple",
			Operation{
				Action: "expire",
				Group:  "someGroup",
			},
			nil,
		},
		{
			"action set",
			Operation{
				Action: "set",
				Name:   "metrics_1",
				Value:  &stubFloat,
			},
			nil,
		},
		{
			"histgoram",
			Operation{
				Name:    "metrics_1",
				Action:  "observe",
				Value:   &stubFloat,
				Buckets: []float64{1, 2, 3},
			},
			nil,
		},
		{
			"invalid",
			Operation{
				Action: "expired",
				Group:  "someGroup",
			},
			errors.New("'name' is required when action is not 'expire': {Name: Group:someGroup Action:expired Value:<nil> Buckets:[] Labels:map[]}"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.op.Validate()
			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())

				return
			}

			assert.NoError(t, err)
		})
	}
}
