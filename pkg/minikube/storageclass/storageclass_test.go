package storageclass

import (
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

var orgGetStoragev1 func() (storagev1.StorageV1Interface, error)
var orgMockStorageProvisioner, mockK8sConfig string

func init() {
	orgMockStorageProvisioner = "foo"
	orgGetStoragev1 = getStoragev1
	mockK8sConfig = `apiVersion: v1
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

var mockStorageProvisioner string

func (mockStorageClassInterface) List(opts metav1.ListOptions) (*v1.StorageClassList, error) {
	scl := v1.StorageClassList{}
	scl.Items = append(scl.Items, v1.StorageClass{Provisioner: mockStorageProvisioner})
	return &scl, nil
}

func (mockStorageClassInterface) Update(sc *v1.StorageClass) (*v1.StorageClass, error) {
	if strings.HasPrefix(sc.Provisioner, "bad") {
		return nil, fmt.Errorf("bad provisioner")
	}
	s := v1.StorageClass{}
	return &s, nil
}

func TestSetDefaultStorageClassNoFake(t *testing.T) {
	originalEnv := os.Getenv("KUBECONFIG")
	defer func() {
		err := os.Setenv("KUBECONFIG", originalEnv)
		if err != nil {
			t.Fatalf("Error reverting env KUBECONFIG to its original value. Got err (%s)", err)
		}
	}()
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
			if err := setK8SConfig(test.config, test.kubeconfigPath); err != nil {
				t.Fatalf(err.Error())
			}
			defer func() { os.Remove(test.kubeconfigPath) }()

			err := SetDefaultStorageClass(test.class)
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
		provisioner string
		err         bool
	}{
		{
			description: "ok",
			provisioner: "foo",
		},
		{
			description: "bad provisioner",
			provisioner: "bad",
			err:         true,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			mockStorageProvisioner = test.provisioner
			defer func() { mockStorageProvisioner = orgMockStorageProvisioner }()
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
			class: "bad-class-provisioner",
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

func TestDisableDefaultStorageClassNoFake(t *testing.T) {
	originalEnv := os.Getenv("KUBECONFIG")
	defer func() {
		err := os.Setenv("KUBECONFIG", originalEnv)
		if err != nil {
			t.Fatalf("Error reverting env KUBECONFIG to its original value. Got err (%s)", err)
		}
	}()
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
			if err := setK8SConfig(test.config, test.kubeconfigPath); err != nil {
				t.Fatalf(err.Error())
			}
			defer func() { os.Remove(test.kubeconfigPath) }()

			err := DisableDefaultStorageClass(test.class)
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
