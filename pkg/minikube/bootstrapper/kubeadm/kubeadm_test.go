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
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/cruntime"
	"k8s.io/minikube/pkg/util"
)

func TestGenerateConfig(t *testing.T) {
	tests := []struct {
		description string
		cfg         config.KubernetesConfig
		expectedCfg string
		shouldErr   bool
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
			description: "extra args, v1.14.0",
			cfg: config.KubernetesConfig{
				NodeIP:            "192.168.1.101",
				KubernetesVersion: "v1.14.0-beta1",
				NodeName:          "extra-args-minikube-114",
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
				},
			},
			expectedCfg: `apiVersion: kubeadm.k8s.io/v1beta1
kind: InitConfiguration
localAPIEndpoint:
  advertiseAddress: 192.168.1.101
  bindPort: 8443
bootstrapTokens:
- groups:
  - system:bootstrappers:kubeadm:default-node-token
  ttl: 24h0m0s
  usages:
  - signing
  - authentication
nodeRegistration:
  criSocket: /var/run/dockershim.sock
  name: extra-args-minikube-114
  taints: []
---
apiVersion: kubeadm.k8s.io/v1beta1
kind: ClusterConfiguration
apiServer:
  extraArgs:
    enable-admission-plugins: "Initializers,NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota"fail-no-swap: "true"
controllerManager:
  extraArgs:
    kube-api-burst: "32"
certificatesDir: /var/lib/minikube/certs/
clusterName: kubernetes
controlPlaneEndpoint: localhost:8443
dns:
  type: CoreDNS
etcd:
  local:
    dataDir: /data/minikube
kubernetesVersion: v1.14.0-beta1
networking:
  dnsDomain: cluster.local
  podSubnet: ""
  serviceSubnet: 10.96.0.0/12
---
apiVersion: kubelet.config.k8s.io/v1beta1
kind: KubeletConfiguration
imageGCHighThresholdPercent: 100
evictionHard:
  nodefs.available: "0%"
  nodefs.inodesFree: "0%"
  imagefs.available: "0%"
`,
		}, {
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
	}

	for _, test := range tests {
		runtime, err := cruntime.New(cruntime.Config{Type: "docker"})
		if err != nil {
			t.Fatalf("runtime: %v", err)
		}

		t.Run(test.description, func(t *testing.T) {
			got, err := generateConfig(test.cfg, runtime)
			if err != nil && !test.shouldErr {
				t.Errorf("got unexpected error generating config: %v", err)
				return
			}
			if err == nil && test.shouldErr {
				t.Errorf("expected error but got none, config: %s", got)
				return
			}

			// cmp.Diff doesn't present diffs of multi-line text well
			gotSplit := strings.Split(got, "\n")
			wantSplit := strings.Split(test.expectedCfg, "\n")
			if diff := cmp.Diff(gotSplit, wantSplit); diff != "" {
				t.Errorf("unexpected diff: (-want +got)\n%s\ngot: %s\n", diff, got)
			}
		})
	}
}
