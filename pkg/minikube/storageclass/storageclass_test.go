package storageclass

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	v1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	storagev1 "k8s.io/client-go/kubernetes/typed/storage/v1"
	testing_client "k8s.io/client-go/testing"
)

var orgGetStoragev1 func() (storagev1.StorageV1Interface, error)

func init() {
	orgGetStoragev1 = getStoragev1
}

func mockGetStoragev1() (storagev1.StorageV1Interface, error) {
	client := fake.Clientset{Fake: testing_client.Fake{}}
	s := client.StorageV1()
	return mockStorageV1Interface{s}, nil
}

func (mockStorageV1Interface) StorageClasses() storagev1.StorageClassInterface {
	return mockStorageClassInterface{}
}

type mockStorageV1Interface struct {
	storagev1.StorageV1Interface
}

type mockStorageClassInterface struct {
	storagev1.StorageClassInterface
}

func (mockStorageClassInterface) Get(name string, options metav1.GetOptions) (*v1.StorageClass, error) {
	if name == "no-such-class" {
		return nil, fmt.Errorf("mocked error. No such class")
	}
	sc := v1.StorageClass{Provisioner: name}
	return &sc, nil
}

func (mockStorageClassInterface) List(opts metav1.ListOptions) (*v1.StorageClassList, error) {
	scl := v1.StorageClassList{}
	return &scl, nil
}

func (mockStorageClassInterface) Update(v *v1.StorageClass) (*v1.StorageClass, error) {
	sc := v1.StorageClass{}
	return &sc, nil
}

func TestSetDefaultStorageClassNoFake(t *testing.T) {
	originalEnv := os.Getenv("KUBECONFIG")
	defer func() {
		err := os.Setenv("KUBECONFIG", originalEnv)
		if err != nil {
			t.Fatalf("Error reverting env KUBECONFIG to its original value. Got err (%s)", err)
		}
	}()
	mockK8sConfig := `apiVersion: v1
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
	var tests = []struct {
		description    string
		kubeconfigPath string
		config         string
		class          string
		err            bool
	}{
		{
			description:    "list storage class error",
			kubeconfigPath: "/tmp/kube_config",
			config:         mockK8sConfig,
			class:          "standard",
			err:            true,
		},
		{
			description:    "broken config",
			kubeconfigPath: "/tmp/kube_config",
			config:         "this**is&&not: yaml::valid: file",
			class:          "standard",
			err:            true,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			mockK8sConfigByte := []byte(test.config)
			mockK8sConfigPath := test.kubeconfigPath
			err := ioutil.WriteFile(mockK8sConfigPath, mockK8sConfigByte, 0644)
			defer func() { os.Remove(mockK8sConfigPath) }()
			if err != nil {
				t.Fatalf("Unexpected error when writing to file %v. Error: %v", test.kubeconfigPath, err)
			}
			os.Setenv("KUBECONFIG", mockK8sConfigPath)

			err = SetDefaultStorageClass(test.class)
			if err != nil && !test.err {
				t.Fatalf("Unexpected err: %v for test: %v", err, test.description)
			}
			if err == nil && test.err {
				t.Fatalf("Expected err for test: %v", test.description)
			}
		})
	}
}

func TestSetDefaultStorageClassFakeStorage(t *testing.T) {
	getStoragev1 = mockGetStoragev1
	defer func() { getStoragev1 = orgGetStoragev1 }()
	var tests = []struct {
		description string
		name        string
		class       string
		err         bool
	}{
		{
			description: "list storage class error",
			class:       "standard",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			err := SetDefaultStorageClass(test.name)
			if err != nil && !test.err {
				t.Fatalf("Unexpected err: %v for test: %v", err, test.description)
			}
			if err == nil && test.err {
				t.Fatalf("Expected err for test: %v", test.description)
			}
		})
	}
}

func TestSetDisableDefaultStorageClassFakeStorage(t *testing.T) {
	getStoragev1 = mockGetStoragev1
	defer func() { getStoragev1 = orgGetStoragev1 }()
	var tests = []struct {
		class string
		err   bool
	}{
		{
			class: "no-such-class",
			err:   true,
		},
		{
			class: "default",
		},
	}

	for _, test := range tests {
		t.Run(test.class, func(t *testing.T) {
			err := DisableDefaultStorageClass(test.class)
			if err != nil && !test.err {
				t.Fatalf("Unexpected err: %v for test: %v", err, test.class)
			}
			if err == nil && test.err {
				t.Fatalf("Expected err for test: %v", test.class)
			}
		})
	}
}
