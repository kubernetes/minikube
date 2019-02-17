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

package kubeadm

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/util"
)

func TestGenerateKubeletConfig(t *testing.T) {
	tests := []struct {
		description string
		cfg         config.KubernetesConfig
		expectedCfg string
		shouldErr   bool
	}{
		{
			description: "docker runtime",
			cfg: config.KubernetesConfig{
				NodeIP:            "192.168.1.100",
				KubernetesVersion: "v1.1.0",
				NodeName:          "minikube",
				ContainerRuntime:  "docker",
			},
			expectedCfg: `
[Unit]
Wants=docker.socket

[Service]
ExecStart=
ExecStart=/usr/bin/kubelet --allow-privileged=true --authorization-mode=Webhook --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --cadvisor-port=0 --cgroup-driver=cgroupfs --client-ca-file=/var/lib/minikube/certs/ca.crt --cluster-dns=10.96.0.10 --cluster-domain=cluster.local --container-runtime=docker --hostname-override=minikube --kubeconfig=/etc/kubernetes/kubelet.conf --pod-manifest-path=/etc/kubernetes/manifests --require-kubeconfig=true

[Install]
`,
		},
		{
			description: "cri runtime",
			cfg: config.KubernetesConfig{
				NodeIP:            "192.168.1.100",
				KubernetesVersion: "v1.1.0",
				NodeName:          "minikube",
				ContainerRuntime:  "cri-o",
			},
			expectedCfg: `
[Unit]
Wants=crio.service

[Service]
ExecStart=
ExecStart=/usr/bin/kubelet --allow-privileged=true --authorization-mode=Webhook --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --cadvisor-port=0 --cgroup-driver=cgroupfs --client-ca-file=/var/lib/minikube/certs/ca.crt --cluster-dns=10.96.0.10 --cluster-domain=cluster.local --container-runtime=remote --container-runtime-endpoint=/var/run/crio/crio.sock --hostname-override=minikube --image-service-endpoint=/var/run/crio/crio.sock --kubeconfig=/etc/kubernetes/kubelet.conf --pod-manifest-path=/etc/kubernetes/manifests --require-kubeconfig=true --runtime-request-timeout=15m

[Install]
`,
		},
		{
			description: "docker runtime with custom image repository",
			cfg: config.KubernetesConfig{
				NodeIP:            "192.168.1.100",
				KubernetesVersion: "v1.1.0",
				NodeName:          "minikube",
				ContainerRuntime:  "docker",
				ImageRepository:   "docker-proxy-image.io/google_containers",
			},
			expectedCfg: `
[Unit]
Wants=docker.socket

[Service]
ExecStart=
ExecStart=/usr/bin/kubelet --allow-privileged=true --authorization-mode=Webhook --bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --cadvisor-port=0 --cgroup-driver=cgroupfs --client-ca-file=/var/lib/minikube/certs/ca.crt --cluster-dns=10.96.0.10 --cluster-domain=cluster.local --container-runtime=docker --hostname-override=minikube --kubeconfig=/etc/kubernetes/kubelet.conf --pod-infra-container-image=docker-proxy-image.io/google_containers//pause-amd64:3.0 --pod-manifest-path=/etc/kubernetes/manifests --require-kubeconfig=true

[Install]
`,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			runtime, err := cruntime.New(cruntime.Config{Type: test.cfg.ContainerRuntime})
			if err != nil {
				t.Fatalf("runtime: %v", err)
			}

			actualCfg, err := NewKubeletConfig(test.cfg, runtime)
			if err != nil && !test.shouldErr {
				t.Errorf("got unexpected error generating config: %v", err)
				return
			}
			if err == nil && test.shouldErr {
				t.Errorf("expected error but got none, config: %s", actualCfg)
				return
			}
			if diff := cmp.Diff(test.expectedCfg, actualCfg); diff != "" {
				t.Errorf("actual config does not match expected.  (-want +got)\n%s", diff)
			}
		})
	}
}

func TestGenerateConfig(t *testing.T) {
	tests := []struct {
		description  string
		cfg          config.KubernetesConfig
		expectedCfg  string
		expectedOpts interface{}
		shouldErr    bool
	}{
		{
			description: "no extra args",
			cfg: config.KubernetesConfig{
				NodeIP:            "192.168.1.100",
				KubernetesVersion: "v1.10.0",
				NodeName:          "minikube",
			},
			expectedCfg: `apiVersion: kubeadm.k8s.io/v1alpha1
kind: MasterConfiguration
noTaintMaster: true
api:
  advertiseAddress: 192.168.1.100
  bindPort: 8443
  controlPlaneEndpoint: localhost
kubernetesVersion: v1.10.0
certificatesDir: /var/lib/minikube/certs/
networking:
  serviceSubnet: 10.96.0.0/12
etcd:
  dataDir: /data/minikube
nodeName: minikube
apiServerExtraArgs:
  admission-control: "Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota"
`,
			expectedOpts: struct {
				CertDir           string
				ServiceCIDR       string
				AdvertiseAddress  string
				APIServerPort     int
				KubernetesVersion string
				EtcdDataDir       string
				NodeName          string
				CRISocket         string
				ImageRepository   string
				ExtraArgs         []ComponentExtraArgs
				FeatureArgs       map[string]bool
				NoTaintMaster     bool
			}{
				CertDir:           util.DefaultCertPath,
				ServiceCIDR:       util.DefaultServiceCIDR,
				AdvertiseAddress:  "192.168.1.100",
				APIServerPort:     8443,
				KubernetesVersion: "v1.10.0",
				EtcdDataDir:       "/data/minikube",
				NodeName:          "minikube",
				ExtraArgs: []ComponentExtraArgs{
					{
						Component: "apiServerExtraArgs",
						Options: map[string]string{
							"admission-control": "Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota",
						},
					},
				},
				FeatureArgs:   map[string]bool{},
				NoTaintMaster: true},
		},
		{
			description: "extra args all components",
			cfg: config.KubernetesConfig{
				NodeIP:            "192.168.1.101",
				KubernetesVersion: "v1.10.0-alpha.0",
				NodeName:          "extra-args-minikube",
				ExtraOptions: util.ExtraOptionSlice{
					util.ExtraOption{
						Component: Apiserver,
						Key:       "fail-no-swap",
						Value:     "true",
					},
					util.ExtraOption{
						Component: ControllerManager,
						Key:       "kube-api-burst",
						Value:     "32",
					},
					util.ExtraOption{
						Component: Scheduler,
						Key:       "scheduler-name",
						Value:     "mini-scheduler",
					},
				},
			},
			expectedCfg: `apiVersion: kubeadm.k8s.io/v1alpha1
kind: MasterConfiguration
noTaintMaster: true
api:
  advertiseAddress: 192.168.1.101
  bindPort: 8443
  controlPlaneEndpoint: localhost
kubernetesVersion: v1.10.0-alpha.0
certificatesDir: /var/lib/minikube/certs/
networking:
  serviceSubnet: 10.96.0.0/12
etcd:
  dataDir: /data/minikube
nodeName: extra-args-minikube
apiServerExtraArgs:
  admission-control: "Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota"
  fail-no-swap: "true"
controllerManagerExtraArgs:
  kube-api-burst: "32"
schedulerExtraArgs:
  scheduler-name: "mini-scheduler"
`,
		},
		{
			description: "two extra args for one component",
			cfg: config.KubernetesConfig{
				NodeIP:            "192.168.1.101",
				KubernetesVersion: "v1.10.0-alpha.0",
				NodeName:          "extra-args-minikube",
				ExtraOptions: util.ExtraOptionSlice{
					util.ExtraOption{
						Component: Apiserver,
						Key:       "fail-no-swap",
						Value:     "true",
					},
					util.ExtraOption{
						Component: Apiserver,
						Key:       "kube-api-burst",
						Value:     "32",
					},
				},
			},
			expectedCfg: `apiVersion: kubeadm.k8s.io/v1alpha1
kind: MasterConfiguration
noTaintMaster: true
api:
  advertiseAddress: 192.168.1.101
  bindPort: 8443
  controlPlaneEndpoint: localhost
kubernetesVersion: v1.10.0-alpha.0
certificatesDir: /var/lib/minikube/certs/
networking:
  serviceSubnet: 10.96.0.0/12
etcd:
  dataDir: /data/minikube
nodeName: extra-args-minikube
apiServerExtraArgs:
  admission-control: "Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota"
  fail-no-swap: "true"
  kube-api-burst: "32"
`,
		},
		{
			description: "enable feature gates",
			cfg: config.KubernetesConfig{
				NodeIP:            "192.168.1.101",
				KubernetesVersion: "v1.10.0-alpha.0",
				NodeName:          "extra-args-minikube",
				FeatureGates:      "HugePages=true,OtherFeature=false",
			},
			expectedCfg: `apiVersion: kubeadm.k8s.io/v1alpha1
kind: MasterConfiguration
noTaintMaster: true
api:
  advertiseAddress: 192.168.1.101
  bindPort: 8443
  controlPlaneEndpoint: localhost
kubernetesVersion: v1.10.0-alpha.0
certificatesDir: /var/lib/minikube/certs/
networking:
  serviceSubnet: 10.96.0.0/12
etcd:
  dataDir: /data/minikube
nodeName: extra-args-minikube
apiServerExtraArgs:
  admission-control: "Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota"
  feature-gates: "HugePages=true,OtherFeature=false"
controllerManagerExtraArgs:
  feature-gates: "HugePages=true,OtherFeature=false"
schedulerExtraArgs:
  feature-gates: "HugePages=true,OtherFeature=false"
`,
		},
		{
			description: "enable feature gates and extra config",
			cfg: config.KubernetesConfig{
				NodeIP:            "192.168.1.101",
				KubernetesVersion: "v1.10.0-alpha.0",
				NodeName:          "extra-args-minikube",
				FeatureGates:      "HugePages=true,OtherFeature=false",
				ExtraOptions: util.ExtraOptionSlice{
					util.ExtraOption{
						Component: Apiserver,
						Key:       "fail-no-swap",
						Value:     "true",
					},
				},
			},
			expectedCfg: `apiVersion: kubeadm.k8s.io/v1alpha1
kind: MasterConfiguration
noTaintMaster: true
api:
  advertiseAddress: 192.168.1.101
  bindPort: 8443
  controlPlaneEndpoint: localhost
kubernetesVersion: v1.10.0-alpha.0
certificatesDir: /var/lib/minikube/certs/
networking:
  serviceSubnet: 10.96.0.0/12
etcd:
  dataDir: /data/minikube
nodeName: extra-args-minikube
apiServerExtraArgs:
  admission-control: "Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota"
  fail-no-swap: "true"
  feature-gates: "HugePages=true,OtherFeature=false"
controllerManagerExtraArgs:
  feature-gates: "HugePages=true,OtherFeature=false"
schedulerExtraArgs:
  feature-gates: "HugePages=true,OtherFeature=false"
`,
		},
		{
			// Unknown components should fail silently
			description: "unknown component",
			cfg: config.KubernetesConfig{
				NodeIP:            "192.168.1.101",
				KubernetesVersion: "v1.8.0-alpha.0",
				NodeName:          "extra-args-minikube",
				ExtraOptions: util.ExtraOptionSlice{
					util.ExtraOption{
						Component: "not-a-real-component",
						Key:       "killswitch",
						Value:     "true",
					},
				},
			},
			shouldErr: true,
		},
		{
			description: "custom api server port",
			cfg: config.KubernetesConfig{
				NodeIP:            "192.168.1.100",
				NodePort:          18443,
				KubernetesVersion: "v1.10.0",
				NodeName:          "minikube",
			},
			expectedCfg: `apiVersion: kubeadm.k8s.io/v1alpha1
kind: MasterConfiguration
noTaintMaster: true
api:
  advertiseAddress: 192.168.1.100
  bindPort: 18443
  controlPlaneEndpoint: localhost
kubernetesVersion: v1.10.0
certificatesDir: /var/lib/minikube/certs/
networking:
  serviceSubnet: 10.96.0.0/12
etcd:
  dataDir: /data/minikube
nodeName: minikube
apiServerExtraArgs:
  admission-control: "Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota"
`,
		},
		{
			description: "custom image repository",
			cfg: config.KubernetesConfig{
				NodeIP:            "192.168.1.100",
				KubernetesVersion: "v1.10.0",
				NodeName:          "minikube",
				ImageRepository:   "docker-proxy-image.io/google_containers",
			},
			expectedCfg: `apiVersion: kubeadm.k8s.io/v1alpha1
kind: MasterConfiguration
noTaintMaster: true
api:
  advertiseAddress: 192.168.1.100
  bindPort: 8443
  controlPlaneEndpoint: localhost
kubernetesVersion: v1.10.0
certificatesDir: /var/lib/minikube/certs/
networking:
  serviceSubnet: 10.96.0.0/12
etcd:
  dataDir: /data/minikube
nodeName: minikube
imageRepository: docker-proxy-image.io/google_containers
apiServerExtraArgs:
  admission-control: "Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota"
`,
			expectedOpts: struct {
				CertDir           string
				ServiceCIDR       string
				AdvertiseAddress  string
				APIServerPort     int
				KubernetesVersion string
				EtcdDataDir       string
				NodeName          string
				CRISocket         string
				ImageRepository   string
				ExtraArgs         []ComponentExtraArgs
				FeatureArgs       map[string]bool
				NoTaintMaster     bool
			}{
				CertDir:           util.DefaultCertPath,
				ServiceCIDR:       util.DefaultServiceCIDR,
				AdvertiseAddress:  "192.168.1.100",
				APIServerPort:     8443,
				KubernetesVersion: "v1.10.0",
				EtcdDataDir:       "/data/minikube",
				NodeName:          "minikube",
				ExtraArgs: []ComponentExtraArgs{
					{
						Component: "apiServerExtraArgs",
						Options: map[string]string{
							"admission-control": "Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota",
						},
					},
				},
				ImageRepository: "docker-proxy-image.io/google_containers",
				FeatureArgs:     map[string]bool{},
				NoTaintMaster:   true},
		},
	}

	for _, test := range tests {
		runtime, err := cruntime.New(cruntime.Config{Type: "docker"})
		if err != nil {
			t.Fatalf("runtime: %v", err)
		}

		t.Run(test.description, func(t *testing.T) {
			actualCfg, actualOpts, err := generateConfig(test.cfg, runtime)
			if err != nil && !test.shouldErr {
				t.Errorf("got unexpected error generating config: %v", err)
				return
			}
			if err == nil && test.shouldErr {
				t.Errorf("expected error but got none, config: %s", actualCfg)
				return
			}
			if actualCfg != test.expectedCfg {
				t.Errorf("actual config does not match expected.  actual:\n%sexpected:\n%s", actualCfg, test.expectedCfg)
				return
			}
			if test.expectedOpts != nil {
				if diff := cmp.Diff(test.expectedOpts, actualOpts); diff != "" {
					t.Errorf("opts differ: (-want +got)\n%s", diff)
				}
			}
		})
	}
}
