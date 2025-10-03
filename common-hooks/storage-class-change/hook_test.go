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
	"bytes"
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/deckhouse/module-sdk/pkg/jq"
	"github.com/deckhouse/module-sdk/pkg/utils/ptr"
)

type testCase struct {
	name     string
	filter   string
	object   any
	expected any
}

func TestStorageClassChangeFilter(t *testing.T) {
	cases := []testCase{
		{
			name:   "filter pvc",
			filter: pvcFilter,
			object: corev1.PersistentVolumeClaim{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test",
					Namespace:         "test1",
					DeletionTimestamp: ptr.To(metav1.NewTime(time.Now())),
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					StorageClassName: ptr.To("class"),
				},
			},
			expected: filteredPvc{
				Name:             "test",
				Namespace:        "test1",
				StorageClassName: "class",
				IsDeleted:        true,
			},
		},
		{
			name:   "filter pod",
			filter: podFilter,
			object: corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "test",
					Namespace:         "test1",
					DeletionTimestamp: ptr.To(metav1.NewTime(time.Now())),
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "test",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "testpvc",
								},
							},
						},
					},
				},
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
				},
			},
			expected: filteredPod{
				Name:      "test",
				Namespace: "test1",
				Pvc:       "testpvc",
				Phase:     corev1.PodRunning,
			},
		},
		{
			name:   "filter storage class",
			filter: storageClassFilter,
			object: storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
					Annotations: map[string]string{
						"storageclass.kubernetes.io/is-default-class": "true",
					},
				},
			},
			expected: filteredStorageClass{
				Name:      "test",
				IsDefault: true,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			q, err := jq.NewQuery(tc.filter)
			assert.NoError(t, err)

			res, err := q.FilterObject(context.Background(), tc.object)
			assert.NoError(t, err)

			switch reflect.TypeOf(tc.expected) {
			case reflect.TypeOf(filteredPvc{}):
				obj := new(filteredPvc)
				err = json.NewDecoder(bytes.NewBufferString(res.String())).Decode(obj)
				assert.NoError(t, err)

				assert.Equal(t, *obj, tc.expected)
			case reflect.TypeOf(filteredPod{}):
				obj := new(filteredPod)
				err = json.NewDecoder(bytes.NewBufferString(res.String())).Decode(obj)
				assert.NoError(t, err)

				assert.Equal(t, *obj, tc.expected)
			case reflect.TypeOf(filteredStorageClass{}):
				obj := new(filteredStorageClass)
				err = json.NewDecoder(bytes.NewBufferString(res.String())).Decode(obj)
				assert.NoError(t, err)

				assert.Equal(t, *obj, tc.expected)
			default:
				assert.Fail(t, "unhandled jq query")
			}
		})
	}
}
