/*
Copyright 2023 The Kubernetes Authors All rights reserved.

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

package machine

import (
	"reflect"
	"sort"
	"testing"

	"k8s.io/minikube/pkg/minikube/cruntime"
)

type CacheImageTestCase struct {
	description string
	images      [][]cruntime.ListImage
	expected    []cruntime.ListImage
}

func TestMergeImageLists(t *testing.T) {
	testCases := []CacheImageTestCase{
		// test case 0: from #16556
		// e.g. on node 1 image1 have an item with k8s.gcr.io/image1:v1.0.0 tag
		// and another item with registry.k8s.io/image1:v1.0.0 tag too
		{
			description: "same images with multiple tags appear on one node",
			images: [][]cruntime.ListImage{
				{
					{
						ID:          "image_id_1",
						RepoDigests: []string{"image_digest_1"},
						RepoTags:    []string{"k8s.gcr.io/image1:v1.0.0"},
						Size:        "1",
					},
					{
						ID:          "image_id_2",
						RepoDigests: []string{"image_digest_2"},
						RepoTags:    []string{"registry.k8s.io/image2:v1.0.0"},
						Size:        "1",
					},

					{
						ID:          "image_id_1",
						RepoDigests: []string{"image_digest_1"},
						RepoTags:    []string{"registry.k8s.io/image1:v1.0.0"},
						Size:        "1",
					},
					{
						ID:          "image_id_2",
						RepoDigests: []string{"image_digest_2"},
						RepoTags:    []string{"k8s.gcr.io/image2:v1.0.0"},
						Size:        "1",
					},
				},
			},
			expected: []cruntime.ListImage{
				{
					ID:          "image_id_1",
					RepoDigests: []string{"image_digest_1"},
					RepoTags:    []string{"k8s.gcr.io/image1:v1.0.0", "registry.k8s.io/image1:v1.0.0"},
					Size:        "1",
				},
				{
					ID:          "image_id_2",
					RepoDigests: []string{"image_digest_2"},
					RepoTags:    []string{"k8s.gcr.io/image2:v1.0.0", "registry.k8s.io/image2:v1.0.0"},
					Size:        "1",
				},
			},
		},
		// test case 1: from #16557
		// e.g. image1 have k8s.gcr.io/image1:v1.0.0 tag on node 1 and registry.k8s.io/image1:v1.0.0 on node 2
		{
			description: "same images with multiple tags appear on two node",
			images: [][]cruntime.ListImage{
				{
					{
						ID:          "image_id_1",
						RepoDigests: []string{"image_digest_1"},
						RepoTags:    []string{"k8s.gcr.io/image1:v1.0.0"},
						Size:        "1",
					},
					{
						ID:          "image_id_2",
						RepoDigests: []string{"image_digest_2"},
						RepoTags:    []string{"registry.k8s.io/image2:v1.0.0"},
						Size:        "1",
					},
				},
				{
					{
						ID:          "image_id_1",
						RepoDigests: []string{"image_digest_1"},
						RepoTags:    []string{"registry.k8s.io/image1:v1.0.0"},
						Size:        "1",
					},
					{
						ID:          "image_id_2",
						RepoDigests: []string{"image_digest_2"},
						RepoTags:    []string{"k8s.gcr.io/image2:v1.0.0"},
						Size:        "1",
					},
				},
			},
			expected: []cruntime.ListImage{
				{
					ID:          "image_id_1",
					RepoDigests: []string{"image_digest_1"},
					RepoTags:    []string{"k8s.gcr.io/image1:v1.0.0", "registry.k8s.io/image1:v1.0.0"},
					Size:        "1",
				},
				{
					ID:          "image_id_2",
					RepoDigests: []string{"image_digest_2"},
					RepoTags:    []string{"k8s.gcr.io/image2:v1.0.0", "registry.k8s.io/image2:v1.0.0"},
					Size:        "1",
				},
			},
		},
		// test case 2: from #16557
		// e.g. image1 have k8s.gcr.io/image1:v1.0.0 tag on node 1
		// and both k8s.gcr.io/image1:v1.0.0 and registry.k8s.io/image1:v1.0.0 on node 2
		{
			description: "image has tag1 on node1 and both tag1 & tag2 on node 2",
			images: [][]cruntime.ListImage{
				{
					{
						ID:          "image_id_1",
						RepoDigests: []string{"image_digest_1"},
						RepoTags:    []string{"k8s.gcr.io/image1:v1.0.0"},
						Size:        "1",
					},
					{
						ID:          "image_id_2",
						RepoDigests: []string{"image_digest_2"},
						RepoTags:    []string{"registry.k8s.io/image2:v1.0.0", "k8s.gcr.io/image2:v1.0.0"},
						Size:        "1",
					},
				},
				{
					{
						ID:          "image_id_1",
						RepoDigests: []string{"image_digest_1"},
						RepoTags:    []string{"registry.k8s.io/image1:v1.0.0", "k8s.gcr.io/image1:v1.0.0"},
						Size:        "1",
					},
					{
						ID:          "image_id_2",
						RepoDigests: []string{"image_digest_2"},
						RepoTags:    []string{"k8s.gcr.io/image2:v1.0.0"},
						Size:        "1",
					},
				},
			},
			expected: []cruntime.ListImage{
				{
					ID:          "image_id_1",
					RepoDigests: []string{"image_digest_1"},
					RepoTags:    []string{"k8s.gcr.io/image1:v1.0.0", "registry.k8s.io/image1:v1.0.0"},
					Size:        "1",
				},
				{
					ID:          "image_id_2",
					RepoDigests: []string{"image_digest_2"},
					RepoTags:    []string{"k8s.gcr.io/image2:v1.0.0", "registry.k8s.io/image2:v1.0.0"},
					Size:        "1",
				},
			},
		},

		{
			description: "normal case",
			images: [][]cruntime.ListImage{
				{
					{
						ID:          "image_id_1",
						RepoDigests: []string{"image_digest_1"},
						RepoTags:    []string{"k8s.gcr.io/image1:v1.0.0"},
						Size:        "1",
					},
					{
						ID:          "image_id_2",
						RepoDigests: []string{"image_digest_2"},
						RepoTags:    []string{"registry.k8s.io/image2:v1.0.0"},
						Size:        "1",
					},
				},
				{
					{
						ID:          "image_id_3",
						RepoDigests: []string{"image_digest_3"},
						RepoTags:    []string{"registry.k8s.io/image3:v1.0.0"},
						Size:        "1",
					},
					{
						ID:          "image_id_4",
						RepoDigests: []string{"image_digest_4"},
						RepoTags:    []string{"k8s.gcr.io/image4:v1.0.0"},
						Size:        "1",
					},
				},
			},
			expected: []cruntime.ListImage{
				{
					ID:          "image_id_1",
					RepoDigests: []string{"image_digest_1"},
					RepoTags:    []string{"k8s.gcr.io/image1:v1.0.0"},
					Size:        "1",
				},
				{
					ID:          "image_id_2",
					RepoDigests: []string{"image_digest_2"},
					RepoTags:    []string{"registry.k8s.io/image2:v1.0.0"},
					Size:        "1",
				},
				{
					ID:          "image_id_3",
					RepoDigests: []string{"image_digest_3"},
					RepoTags:    []string{"registry.k8s.io/image3:v1.0.0"},
					Size:        "1",
				},
				{
					ID:          "image_id_4",
					RepoDigests: []string{"image_digest_4"},
					RepoTags:    []string{"k8s.gcr.io/image4:v1.0.0"},
					Size:        "1",
				},
			},
		},
	}

	for _, tc := range testCases {
		got := mergeImageLists(tc.images)
		sort.Slice(got, func(i, j int) bool {
			return got[i].ID < got[j].ID
		})
		for _, img := range got {
			sort.Slice(img.RepoTags, func(i, j int) bool {
				return img.RepoTags[i] < img.RepoTags[j]
			})
		}
		if ok := reflect.DeepEqual(got, tc.expected); !ok {
			t.Errorf("%s:\nmergeImageLists() = %+v;\nwant %+v", tc.description, got, tc.expected)
		}
	}
}
