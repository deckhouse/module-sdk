/*
Copyright 2025 Flant JSC

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

package storageclasschange

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/deckhouse/module-sdk/pkg/utils/ptr"
	"github.com/deckhouse/module-sdk/testing/helpers"
)

// runFilter applies one of the package-private JQ filters and decodes the
// result into the provided destination. Keeping this as a single helper
// makes the tests below trivial table-driven cases.
func runFilter[T any](t *testing.T, filter string, input any) T {
	t.Helper()
	var got T
	require.NoError(t, helpers.JQRunOnObject(context.Background(), filter, input, &got))
	return got
}

func TestPVCFilter_DeletedPVCMarksAsDeleted(t *testing.T) {
	now := metav1.NewTime(time.Now())
	pvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "pvc-1",
			Namespace:         "ns-1",
			DeletionTimestamp: ptr.To(now),
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			StorageClassName: ptr.To("class"),
		},
	}

	got := runFilter[filteredPvc](t, pvcFilter, pvc)

	assert.Equal(t, filteredPvc{
		Name:             "pvc-1",
		Namespace:        "ns-1",
		StorageClassName: "class",
		IsDeleted:        true,
	}, got)
}

func TestPodFilter_ExtractsPVCAndPhase(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod-1",
			Namespace: "ns-1",
		},
		Spec: corev1.PodSpec{
			Volumes: []corev1.Volume{{
				Name: "data",
				VolumeSource: corev1.VolumeSource{
					PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
						ClaimName: "pvc-1",
					},
				},
			}},
		},
		Status: corev1.PodStatus{Phase: corev1.PodRunning},
	}

	got := runFilter[filteredPod](t, podFilter, pod)

	assert.Equal(t, filteredPod{
		Name:      "pod-1",
		Namespace: "ns-1",
		Pvc:       "pvc-1",
		Phase:     corev1.PodRunning,
	}, got)
}

func TestStorageClassFilter_DetectsDefaultByAnnotation(t *testing.T) {
	cases := []struct {
		name       string
		annotation map[string]string
		want       bool
	}{
		{
			name:       "default class",
			annotation: map[string]string{"storageclass.kubernetes.io/is-default-class": "true"},
			want:       true,
		},
		{
			name:       "default class beta",
			annotation: map[string]string{"storageclass.beta.kubernetes.io/is-default-class": "true"},
			want:       true,
		},
		{
			name:       "non-default",
			annotation: map[string]string{"foo": "bar"},
			want:       false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sc := storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Name:        "sc-test",
					Annotations: tc.annotation,
				},
			}

			got := runFilter[filteredStorageClass](t, storageClassFilter, sc)

			assert.Equal(t, "sc-test", got.Name)
			assert.Equal(t, tc.want, got.IsDefault)
		})
	}
}
