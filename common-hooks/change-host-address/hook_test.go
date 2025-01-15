package changehostaddress_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	changehostaddress "github.com/deckhouse/module-sdk/common-hooks/change-host-address"
	"github.com/deckhouse/module-sdk/pkg/jq"
)

func Test_JQFilterGetAddress(t *testing.T) {
	t.Run("get address", func(t *testing.T) {
		const rawSecret = `
	{
		"apiVersion": "v1",
		"kind": "Pod",
		"metadata": {
			"annotations": {
			"node.deckhouse.io/initial-host-ip": "192.168.0.1"
			},
			"name": "some-pod",
			"namespace": "some-ns"
		},
		"status": {
			"hostIP": "192.168.0.2"
		}
	}`

		q, err := jq.NewQuery(changehostaddress.JQFilterGetAddress)
		assert.NoError(t, err)

		res, err := q.FilterStringObject(context.Background(), rawSecret)
		assert.NoError(t, err)

		cert := new(changehostaddress.Address)
		err = json.NewDecoder(bytes.NewBufferString(res.String())).Decode(cert)
		assert.NoError(t, err)

		assert.Equal(t, "some-pod", cert.Name)
		assert.Equal(t, "192.168.0.2", cert.Host)
		assert.Equal(t, "192.168.0.1", cert.InitialHost)
	})
}
