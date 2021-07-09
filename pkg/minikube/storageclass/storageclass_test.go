/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package storageclass

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	v1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	storagev1 "k8s.io/client-go/kubernetes/typed/storage/v1"
	testing_client "k8s.io/client-go/testing"
)

type mockStorageV1InterfaceOk struct {
	storagev1.StorageV1Interface
}
type mockStorageV1InterfaceListErr struct {
	storagev1.StorageV1Interface
}
type mockStorageV1InterfaceWithBadItem struct {
	storagev1.StorageV1Interface
}
type mockStorageClassInterfaceOk struct {
	storagev1.StorageClassInterface
}
type mockStorageClassInterfaceListErr struct {
	storagev1.StorageClassInterface
}
type mockStorageClassInterfaceWithBadItem struct {
	storagev1.StorageClassInterface
}

func testStoragev1Ok() (storagev1.StorageV1Interface, error) {
	client := fake.Clientset{Fake: testing_client.Fake{}}
	return mockStorageV1InterfaceOk{client.StorageV1()}, nil
}
func testStoragev1ListErr() (storagev1.StorageV1Interface, error) {
	client := fake.Clientset{Fake: testing_client.Fake{}}
	return mockStorageV1InterfaceListErr{client.StorageV1()}, nil
}
func testStoragev1WithBadItem() (storagev1.StorageV1Interface, error) {
	client := fake.Clientset{Fake: testing_client.Fake{}}
	return mockStorageV1InterfaceWithBadItem{client.StorageV1()}, nil
}

func (mockStorageV1InterfaceOk) StorageClasses() storagev1.StorageClassInterface {
	return mockStorageClassInterfaceOk{}
}

func (mockStorageV1InterfaceListErr) StorageClasses() storagev1.StorageClassInterface {
	return mockStorageClassInterfaceListErr{}
}

func (mockStorageV1InterfaceWithBadItem) StorageClasses() storagev1.StorageClassInterface {
	return mockStorageClassInterfaceWithBadItem{}
}

func (mockStorageClassInterfaceOk) Get(ctx context.Context, name string, options metav1.GetOptions) (*v1.StorageClass, error) {
	if strings.HasPrefix(name, "bad-class") {
		return nil, fmt.Errorf("mocked error. No such class")
	}
	sc := v1.StorageClass{Provisioner: name}
	return &sc, nil
}

func (m mockStorageClassInterfaceOk) List(ctx context.Context, opts metav1.ListOptions) (*v1.StorageClassList, error) {
	scl := v1.StorageClassList{}
	sc := v1.StorageClass{Provisioner: "standard"}
	scl.Items = append(scl.Items, sc)
	return &scl, nil
}

func (m mockStorageClassInterfaceWithBadItem) List(ctx context.Context, opts metav1.ListOptions) (*v1.StorageClassList, error) {
	scl := v1.StorageClassList{}
	sc := v1.StorageClass{Provisioner: "bad", ObjectMeta: metav1.ObjectMeta{Name: "standard"}}
	scl.Items = append(scl.Items, sc)
	return &scl, nil
}
func (mockStorageClassInterfaceListErr) List(ctx context.Context, opts metav1.ListOptions) (*v1.StorageClassList, error) {
	return nil, fmt.Errorf("mocked list error")
}

func (mockStorageClassInterfaceOk) Update(ctx context.Context, sc *v1.StorageClass, opts metav1.UpdateOptions) (*v1.StorageClass, error) {
	if strings.HasPrefix(sc.Provisioner, "bad") {
		return nil, fmt.Errorf("bad provisioner")
	}
	return &v1.StorageClass{}, nil
}

func (mockStorageClassInterfaceWithBadItem) Update(ctx context.Context, sc *v1.StorageClass, opts metav1.UpdateOptions) (*v1.StorageClass, error) {
	if strings.HasPrefix(sc.Provisioner, "bad") {
		return nil, fmt.Errorf("bad provisioner")
	}
	return &v1.StorageClass{}, nil
}

func TestDisableDefaultStorageClass(t *testing.T) {
	var tests = []struct {
		description string
		class       string
		err         bool
		sv1Fixture  func() (storagev1.StorageV1Interface, error)
	}{
		{
			description: "ok",
			class:       "standard",
			sv1Fixture:  testStoragev1Ok,
		},
		{
			description: "no such class",
			class:       "bad-class",
			err:         true,
			sv1Fixture:  testStoragev1Ok,
		},
		{
			description: "bad existing class",
			class:       "bad-existing-class",
			err:         true,
			sv1Fixture:  testStoragev1Ok,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			sv1, _ := test.sv1Fixture()
			err := DisableDefaultStorageClass(sv1, test.class)
			if err != nil && !test.err {
				t.Fatalf("Unexpected err: %v for test: %v", err, test.description)
			}
			if err == nil && test.err {
				t.Fatalf("Expected err for test: %v", test.description)
			}
		})
	}
}

func TestSetDefaultStorageClass(t *testing.T) {
	var tests = []struct {
		description string
		class       string
		err         bool
		sv1Fixture  func() (storagev1.StorageV1Interface, error)
	}{
		{
			description: "ok (no fail)",
			class:       "standard",
			sv1Fixture:  testStoragev1Ok,
		},
		{
			description: "ok (failed annotation)",
			class:       "standard",
			sv1Fixture:  testStoragev1WithBadItem,
			err:         true,
		},

		{
			description: "list error",
			class:       "standard",
			sv1Fixture:  testStoragev1ListErr,
			err:         true,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			sv1, _ := test.sv1Fixture()

			err := SetDefaultStorageClass(sv1, test.class)
			if err != nil && !test.err {
				t.Fatalf("Unexpected err: %v for test: %v", err, test.description)
			}
			if err == nil && test.err {
				t.Fatalf("Expected err for test: %v", test.description)
			}
		})
	}
}

var mockK8sConfig = `apiVersion: v1
clusters:
- cluster:
    server: https://example.com:443
  name: minikube
contexts:
- context:
    cluster: minikube
    user: minikube
  name: minikube
current-context: minikube
kind: Config
preferences: {}
users:
- name: minikube
`

func TestGetStoragev1(t *testing.T) {
	var tests = []struct {
		description string
		config      string
		err         bool
	}{
		{
			description: "ok",
			config:      mockK8sConfig,
		},
		{
			description: "no valid config",
			config:      "this is not valid config",
			err:         true,
		},
	}
	configFile, err := ioutil.TempFile("/tmp", "")
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer os.Remove(configFile.Name())
	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			if err := setK8SConfig(test.config, configFile.Name()); err != nil {
				t.Fatalf(err.Error())
			}

			// context name is hardcoded by mockK8sConfig
			_, err = GetStoragev1("minikube")
			if err != nil && !test.err {
				t.Fatalf("Unexpected err: %v for test: %v", err, test.description)
			}
			if err == nil && test.err {
				t.Fatalf("Expected err for test: %v", test.description)
			}
		})
	}
}

func setK8SConfig(config, kubeconfigPath string) error {
	mockK8sConfigByte := []byte(config)
	mockK8sConfigPath := kubeconfigPath
	err := ioutil.WriteFile(mockK8sConfigPath, mockK8sConfigByte, 0644)
	if err != nil {
		return fmt.Errorf("Unexpected error when writing to file %v. Error: %v", kubeconfigPath, err)
	}
	os.Setenv("KUBECONFIG", mockK8sConfigPath)
	return nil
}
