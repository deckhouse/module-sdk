/*
Copyright 2021 Flant JSC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package certificate_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/deckhouse/module-sdk/pkg/certificate"
	"github.com/deckhouse/module-sdk/pkg/jq"
)

func Test_JQFilterApplyCaSelfSignedCert(t *testing.T) {
	t.Run("apply ca self signed cert jq filter", func(t *testing.T) {
		const rawSecret = `
		{
	  "apiVersion": "v1",
	  "data": {
		"ca.crt": "c29tZS1jYQ==",
		"tls.crt": "c29tZS1jcnQ=",
		"tls.key": "c29tZS1rZXk="
	  },
	  "kind": "Secret",
	  "metadata": {
		"name": "some-cert",
		"namespace": "some-ns"
	  },
	  "type": "kubernetes.io/tls"
	}`

		q, err := jq.NewQuery(certificate.JQFilterApplyCaSelfSignedCert)
		assert.NoError(t, err)

		res, err := q.FilterStringObject(context.Background(), rawSecret)
		assert.NoError(t, err)

		auth := new(certificate.Authority)
		err = json.NewDecoder(bytes.NewBufferString(res.String())).Decode(auth)
		assert.NoError(t, err)

		assert.Equal(t, "some-key", string(auth.Key))
		assert.Equal(t, "some-crt", string(auth.Cert))
	})
}
