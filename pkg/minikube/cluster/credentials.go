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

package cluster

import (
	"net"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/util"
)

var (
	// This is the internalIP , the API server and other components communicate on.
	internalIP = net.ParseIP(util.DefaultServiceClusterIP)
)

func GenerateCerts(caCert, caKey, pub, priv string, ip net.IP, name string, dnsDomain string) error {
	if !(util.CanReadFile(caCert) && util.CanReadFile(caKey)) {
		if err := util.GenerateCACert(caCert, caKey, name); err != nil {
			return errors.Wrap(err, "Error generating certificate")
		}
	}

	ips := []net.IP{ip, internalIP}
	if err := util.GenerateSignedCert(pub, priv, ips, util.GetAlternateDNS(dnsDomain), caCert, caKey); err != nil {
		return errors.Wrap(err, "Error generating signed cert")
	}
	return nil
}
