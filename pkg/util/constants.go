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

package util

// These constants are used by both minikube and localkube
const (
	APIServerPort             = 8443
	DefaultLocalkubeDirectory = "/var/lib/localkube"
	DefaultCertPath           = DefaultLocalkubeDirectory + "/certs/"
	DefaultKubeConfigPath     = DefaultLocalkubeDirectory + "/kubeconfig"
	DefaultServiceClusterIP   = "10.0.0.1"
	DefaultDNSDomain          = "cluster.local"
	DefaultDNSIP              = "10.0.0.10"
	DefaultInsecureRegistry   = "10.0.0.0/24"
)

func GetAlternateDNS(domain string) []string {
	return []string{"kubernetes.default.svc." + domain, "kubernetes.default.svc", "kubernetes.default", "kubernetes", "localhost"}
}
