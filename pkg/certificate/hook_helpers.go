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

package certificate

import (
	"fmt"

	"github.com/deckhouse/module-sdk/pkg"
	objectpatch "github.com/deckhouse/module-sdk/pkg/object-patch"
)

var JQFilterApplyCaSelfSignedCert = `{
    "key": .data."tls.key",
    "crt": .data."tls.crt"
}`

func GetOrCreateCa(input *pkg.HookInput, snapshotKey, cn string) (*Authority, error) {
	authorities, err := objectpatch.UnmarshalToStruct[Authority](input.Snapshots, snapshotKey)
	if err != nil {
		return nil, fmt.Errorf("unmarshal to struct: %w", err)
	}

	// what if we have more authorities than one?
	if len(authorities) == 1 {
		return &authorities[0], nil
	}

	selfSignedCA, err := GenerateCA(cn)
	if err != nil {
		return nil, fmt.Errorf("cannot generate selfsigned ca: %w", err)
	}

	return selfSignedCA, nil
}
