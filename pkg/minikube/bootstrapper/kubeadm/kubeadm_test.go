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

	"k8s.io/minikube/pkg/minikube/config"
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
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			actualCfg, err := generateConfig(test.cfg)
			if err != nil && !test.shouldErr {
				t.Errorf("got unexpected error generating config: %s", err)
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
		})
	}
}
