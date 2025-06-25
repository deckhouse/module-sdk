package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	objectpatch "github.com/deckhouse/module-sdk/internal/object-patch"
	"github.com/deckhouse/module-sdk/pkg"
	"github.com/deckhouse/module-sdk/pkg/jq"
)

// input snapshots must be in the format: snapshotName: ["yaml1", "yaml2", ...]
// it can handle multiple YAML documents in a single string, separated by "---"
// it will apply the JQ filter from the hook config to each snapshot
// and return a map of snapshot names to their filtered JSON representations
func PrepareHookSnapshots(t *testing.T, config *pkg.HookConfig, inputSnapshots map[string][]string) pkg.Snapshots {
	formattedSnapshots := make(objectpatch.Snapshots, len(inputSnapshots))
	for snapBindingName, rawSnaps := range inputSnapshots {
		var (
			err   error
			query *jq.Query
		)

		for _, v := range config.Kubernetes {
			if v.Name == snapBindingName {
				fmt.Println("Using JQ filter:", v.JqFilter)
				query, err = jq.NewQuery(v.JqFilter)
				assert.NoError(t, err, "Failed to create JQ query from filter: %s", v.JqFilter)
			}
		}

		for _, rawSnap := range rawSnaps {
			snaps := strings.Split(rawSnap, "---")

			for _, snap := range snaps {
				var yml map[string]interface{}

				err := yaml.Unmarshal([]byte(snap), &yml)
				assert.NoError(t, err, "Failed to unmarshal snapshot YAML: %s", snap)

				jsonSnap, err := json.Marshal(yml)
				assert.NoError(t, err, "Failed to marshal snapshot to JSON: %s", snap)

				fmt.Println("JSON Snapshot:", string(jsonSnap))

				res, err := query.FilterStringObject(context.TODO(), string(jsonSnap))
				assert.NoError(t, err, "Failed to filter snapshot with JQ query: %s", jsonSnap)
				fmt.Println("JSON Snapshot:", res.String())

				formattedSnapshots[snapBindingName] = append(formattedSnapshots[snapBindingName], objectpatch.Snapshot(res.String()))
			}
		}
	}

	return formattedSnapshots
}
