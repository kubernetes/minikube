/*
Copyright 2016 The Kubernetes Authors All rights reserved.

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

package assets

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/tests"
)

// mapsEqual returns true if and only if `a` contains all the same pairs as `b`.
func mapsEqual(a, b map[string]string) bool {
	for aKey, aValue := range a {
		if bValue, ok := b[aKey]; !ok || aValue != bValue {
			return false
		}
	}

	for bKey := range b {
		if _, ok := a[bKey]; !ok {
			return false
		}
	}
	return true
}

func TestParseMapString(t *testing.T) {
	cases := map[string]map[string]string{
		"Ardvark=1,B=2,Cantaloupe=3":         {"Ardvark": "1", "B": "2", "Cantaloupe": "3"},
		"A=,B=2,C=":                          {"A": "", "B": "2", "C": ""},
		"":                                   {},
		"malformed,good=howdy,manyequals==,": {"good": "howdy"},
	}
	for actual, expected := range cases {
		if parsedMap := parseMapString(actual); !mapsEqual(parsedMap, expected) {
			t.Errorf("Parsed map from string \"%s\" differs from expected map: Actual: %v Expected: %v", actual, parsedMap, expected)
		}
	}
}

func TestMergeMaps(t *testing.T) {
	type TestCase struct {
		sourceMap   map[string]string
		overrideMap map[string]string
		expectedMap map[string]string
	}
	cases := []TestCase{
		{
			sourceMap:   map[string]string{"A": "1", "B": "2"},
			overrideMap: map[string]string{"B": "7", "C": "3"},
			expectedMap: map[string]string{"A": "1", "B": "7", "C": "3"},
		},
		{
			sourceMap:   map[string]string{"B": "7", "C": "3"},
			overrideMap: map[string]string{"A": "1", "B": "2"},
			expectedMap: map[string]string{"A": "1", "B": "2", "C": "3"},
		},
		{
			sourceMap:   map[string]string{"B": "7", "C": "3"},
			overrideMap: map[string]string{},
			expectedMap: map[string]string{"B": "7", "C": "3"},
		},
		{
			sourceMap:   map[string]string{},
			overrideMap: map[string]string{"B": "7", "C": "3"},
			expectedMap: map[string]string{"B": "7", "C": "3"},
		},
	}
	for _, test := range cases {
		if actualMap := mergeMaps(test.sourceMap, test.overrideMap); !mapsEqual(actualMap, test.expectedMap) {
			t.Errorf("Merging maps (source=%v, override=%v) differs from expected map: Actual: %v Expected: %v", test.sourceMap, test.overrideMap, actualMap, test.expectedMap)
		}
	}
}

func TestFilterKeySpace(t *testing.T) {
	type TestCase struct {
		keySpace    map[string]string
		targetMap   map[string]string
		expectedMap map[string]string
	}
	cases := []TestCase{
		{
			keySpace:    map[string]string{"A": "0", "B": ""},
			targetMap:   map[string]string{"B": "1", "C": "2", "D": "3"},
			expectedMap: map[string]string{"B": "1"},
		},
		{
			keySpace:    map[string]string{},
			targetMap:   map[string]string{"B": "1", "C": "2", "D": "3"},
			expectedMap: map[string]string{},
		},
		{
			keySpace:    map[string]string{"B": "1", "C": "2", "D": "3"},
			targetMap:   map[string]string{},
			expectedMap: map[string]string{},
		},
	}
	for _, test := range cases {
		if actualMap := filterKeySpace(test.keySpace, test.targetMap); !mapsEqual(actualMap, test.expectedMap) {
			t.Errorf("Filtering keyspace of map (keyspace=%v, target=%v) differs from expected map: Actual: %v Expected: %v", test.keySpace, test.targetMap, actualMap, test.expectedMap)
		}
	}
}

func TestOverrideDefautls(t *testing.T) {
	type TestCase struct {
		defaultMap  map[string]string
		overrideMap map[string]string
		expectedMap map[string]string
	}
	cases := []TestCase{
		{
			defaultMap:  map[string]string{"A": "1", "B": "2", "C": "3"},
			overrideMap: map[string]string{"B": "7", "C": "8"},
			expectedMap: map[string]string{"A": "1", "B": "7", "C": "8"},
		},
		{
			defaultMap:  map[string]string{"A": "1", "B": "2", "C": "3"},
			overrideMap: map[string]string{"B": "7", "D": "8", "E": "9"},
			expectedMap: map[string]string{"A": "1", "B": "7", "C": "3"},
		},
		{
			defaultMap:  map[string]string{"A": "1", "B": "2", "C": "3"},
			overrideMap: map[string]string{"B": "7", "D": "8", "E": "9"},
			expectedMap: map[string]string{"A": "1", "B": "7", "C": "3"},
		},
	}
	for _, test := range cases {
		if actualMap := overrideDefaults(test.defaultMap, test.overrideMap); !mapsEqual(actualMap, test.expectedMap) {
			t.Errorf("Override defaults (defaults=%v, overrides=%v) differs from expected map: Actual: %v Expected: %v", test.defaultMap, test.overrideMap, actualMap, test.expectedMap)
		}
	}
}

func TestSelectAndPersistImages(t *testing.T) {
	gcpAuth := Addons["gcp-auth"]
	gcpAuthImages := gcpAuth.Images

	// this test will write to ~/.minikube/profiles/minikube/config.json so need to create the file
	home := tests.MakeTempDir(t)
	profilePath := filepath.Join(home, "profiles", "minikube")
	if err := os.MkdirAll(profilePath, 0777); err != nil {
		t.Fatalf("failed to create profile directory: %v", err)
	}
	f, err := os.Create(filepath.Join(profilePath, "config.json"))
	if err != nil {
		t.Fatalf("failed to create config file: %v", err)
	}
	defer f.Close()

	type expected struct {
		numImages           int
		numRegistries       int
		numCustomImages     int
		numCustomRegistries int
	}

	test := func(t *testing.T, cc *config.ClusterConfig, e expected) (images, registries map[string]string) {
		images, registries, err := SelectAndPersistImages(gcpAuth, cc)
		if err != nil {
			t.Fatal(err)
		}
		if len(images) != e.numImages {
			t.Errorf("expected %d images but got %v", e.numImages, images)
		}
		if len(registries) != e.numRegistries {
			t.Errorf("expected %d registries but got %v", e.numRegistries, registries)
		}
		if len(cc.CustomAddonImages) != e.numCustomImages {
			t.Errorf("expected %d CustomAddonImages in config but got %+v", e.numCustomImages, cc.CustomAddonImages)
		}
		if len(cc.CustomAddonRegistries) != e.numCustomRegistries {
			t.Errorf("expected %d CustomAddonRegistries in config but got %+v", e.numCustomRegistries, cc.CustomAddonRegistries)
		}
		return images, registries
	}

	t.Run("NoCustomImage", func(t *testing.T) {
		cc := &config.ClusterConfig{}
		e := expected{numImages: 2}
		images, _ := test(t, cc, e)
		checkMatches(t, "KubeWebhookCertgen", images, gcpAuthImages)
		checkMatches(t, "GCPAuthWebhook", images, gcpAuthImages)
	})

	t.Run("ExistingCustomImage", func(t *testing.T) {
		cc := &config.ClusterConfig{
			CustomAddonImages: map[string]string{
				"GCPAuthWebhook": "test123",
			},
		}
		e := expected{numImages: 2, numCustomImages: 1}
		images, _ := test(t, cc, e)
		checkMatches(t, "KubeWebhookCertgen", images, gcpAuthImages)
		checkMatches(t, "GCPAuthWebhook", images, cc.CustomAddonImages)
	})

	t.Run("NewCustomImage", func(t *testing.T) {
		cc := &config.ClusterConfig{}
		e := expected{numImages: 2, numCustomImages: 1}
		addonImages := setAddonImages("GCPAuthWebhook", "test123")
		defer viper.Reset()
		images, _ := test(t, cc, e)
		checkMatches(t, "KubeWebhookCertgen", images, gcpAuthImages)
		checkMatches(t, "GCPAuthWebhook", images, addonImages)
	})

	t.Run("NewAndExistingCustomImages", func(t *testing.T) {
		cc := &config.ClusterConfig{
			CustomAddonImages: map[string]string{
				"GCPAuthWebhook": "test123",
			},
		}
		e := expected{numImages: 2, numCustomImages: 2}
		addonImages := setAddonImages("KubeWebhookCertgen", "test456")
		defer viper.Reset()
		images, _ := test(t, cc, e)
		checkMatches(t, "KubeWebhookCertgen", images, addonImages)
		checkMatches(t, "GCPAuthWebhook", images, cc.CustomAddonImages)
	})

	t.Run("NewOverwritesExistingCustomImage", func(t *testing.T) {
		cc := &config.ClusterConfig{
			CustomAddonImages: map[string]string{
				"GCPAuthWebhook": "test123",
			},
		}
		e := expected{numImages: 2, numCustomImages: 1}
		addonImages := setAddonImages("GCPAuthWebhook", "test456")
		defer viper.Reset()
		images, _ := test(t, cc, e)
		checkMatches(t, "KubeWebhookCertgen", images, gcpAuthImages)
		checkMatches(t, "GCPAuthWebhook", images, addonImages)
	})

	t.Run("NewUnrelatedCustomImageWithExistingCustomImage", func(t *testing.T) {
		cc := &config.ClusterConfig{
			CustomAddonImages: map[string]string{
				"GCPAuthWebhook": "test123",
			},
		}
		e := expected{numImages: 2, numCustomImages: 1}
		setAddonImages("IngressDNS", "test456")
		defer viper.Reset()
		images, _ := test(t, cc, e)
		checkMatches(t, "KubeWebhookCertgen", images, gcpAuthImages)
		checkMatches(t, "GCPAuthWebhook", images, cc.CustomAddonImages)
	})

	t.Run("NewCustomImageWithExistingUnrelatedCustomImage", func(t *testing.T) {
		cc := &config.ClusterConfig{
			CustomAddonImages: map[string]string{
				"IngressDNS": "test123",
			},
		}
		e := expected{numImages: 2, numCustomImages: 2}
		addonImages := setAddonImages("GCPAuthWebhook", "test456")
		defer viper.Reset()
		images, _ := test(t, cc, e)
		checkMatches(t, "KubeWebhookCertgen", images, gcpAuthImages)
		checkMatches(t, "GCPAuthWebhook", images, addonImages)
	})

	t.Run("ExistingCustomRegistry", func(t *testing.T) {
		cc := &config.ClusterConfig{
			CustomAddonRegistries: map[string]string{
				"GCPAuthWebhook": "test123",
			},
		}
		e := expected{numImages: 2, numRegistries: 1, numCustomRegistries: 1}
		_, registries := test(t, cc, e)
		checkMatches(t, "GCPAuthWebhook", registries, cc.CustomAddonRegistries)
	})

	t.Run("NewCustomRegistry", func(t *testing.T) {
		cc := &config.ClusterConfig{}
		e := expected{numImages: 2, numRegistries: 1, numCustomRegistries: 1}
		addonRegistries := setAddonRegistries("GCPAuthWebhook", "test123")
		defer viper.Reset()
		_, registries := test(t, cc, e)
		checkMatches(t, "GCPAuthWebhook", registries, addonRegistries)
	})

	t.Run("NewAndExistingCustomRegistries", func(t *testing.T) {
		cc := &config.ClusterConfig{
			CustomAddonRegistries: map[string]string{
				"GCPAuthWebhook": "test123",
			},
		}
		e := expected{numImages: 2, numRegistries: 2, numCustomRegistries: 2}
		addonRegistries := setAddonRegistries("KubeWebhookCertgen", "test456")
		defer viper.Reset()
		_, registries := test(t, cc, e)
		checkMatches(t, "GCPAuthWebhook", registries, cc.CustomAddonRegistries)
		checkMatches(t, "KubeWebhookCertgen", registries, addonRegistries)
	})

	t.Run("NewOverwritesExistingCustomRegistry", func(t *testing.T) {
		cc := &config.ClusterConfig{
			CustomAddonRegistries: map[string]string{
				"GCPAuthWebhook": "test123",
			},
		}
		e := expected{numImages: 2, numRegistries: 1, numCustomRegistries: 1}
		addonRegistries := setAddonRegistries("GCPAuthWebhook", "test456")
		defer viper.Reset()
		_, registries := test(t, cc, e)
		checkMatches(t, "GCPAuthWebhook", registries, addonRegistries)
	})

	t.Run("NewUnrelatedCustomRegistryWithExistingCustomRegistry", func(t *testing.T) {
		cc := &config.ClusterConfig{
			CustomAddonRegistries: map[string]string{
				"GCPAuthWebhook": "test123",
			},
		}
		e := expected{numImages: 2, numRegistries: 1, numCustomRegistries: 1}
		setAddonRegistries("IngressDNS", "test456")
		defer viper.Reset()
		_, registries := test(t, cc, e)
		checkMatches(t, "GCPAuthWebhook", registries, cc.CustomAddonRegistries)
	})

	t.Run("NewCustomRegistryWithExistingUnrelatedCustomRegistry", func(t *testing.T) {
		cc := &config.ClusterConfig{
			CustomAddonRegistries: map[string]string{
				"IngressDNS": "test123",
			},
		}
		e := expected{numImages: 2, numRegistries: 1, numCustomRegistries: 2}
		addonRegistries := setAddonRegistries("GCPAuthWebhook", "test456")
		defer viper.Reset()
		_, registries := test(t, cc, e)
		checkMatches(t, "GCPAuthWebhook", registries, addonRegistries)
	})
}

func setAddonImages(k, v string) map[string]string {
	return setFlag(config.AddonImages, k, v)
}

func setAddonRegistries(k, v string) map[string]string {
	return setFlag(config.AddonRegistries, k, v)
}

func setFlag(name, k, v string) map[string]string {
	viper.Set(name, fmt.Sprintf("%s=%s", k, v))
	return map[string]string{k: v}
}

func checkMatches(t *testing.T, name string, got, expected map[string]string) {
	if expected[name] != got[name] {
		t.Errorf("expected %q to be %q, but got %q", name, expected[name], got[name])
	}
}
