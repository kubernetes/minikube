/*
Copyright 2018 The Kubernetes Authors All rights reserved.

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

package tunnel

import (
	"fmt"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	typed_core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
)

// requestSender is an interface exposed for testing what requests are sent through the k8s REST client
type requestSender interface {
	send(request *rest.Request) ([]byte, error)
}

// patchConverter is an interface exposed for testing what patches are sent through the k8s REST client
type patchConverter interface {
	convert(restClient rest.Interface, patch *Patch) *rest.Request
}

// LoadBalancerEmulator is the main struct for emulating the loadbalancer behavior. it sets the ingress to the cluster IP
type LoadBalancerEmulator struct {
	coreV1Client   typed_core.CoreV1Interface
	requestSender  requestSender
	patchConverter patchConverter
}

// PatchServices will update all load balancer services
func (l *LoadBalancerEmulator) PatchServices() ([]string, error) {
	return l.applyOnLBServices(l.updateService)
}

// PatchServiceIP will patch the given service and ip
func (l *LoadBalancerEmulator) PatchServiceIP(restClient rest.Interface, svc core.Service, ip string) error {
	// TODO: do not ignore result
	_, err := l.updateServiceIP(restClient, svc, ip)
	return err
}

// Cleanup will clean up all load balancer services
func (l *LoadBalancerEmulator) Cleanup() ([]string, error) {
	return l.applyOnLBServices(l.cleanupService)
}

func (l *LoadBalancerEmulator) applyOnLBServices(action func(restClient rest.Interface, svc core.Service) ([]byte, error)) ([]string, error) {
	services := l.coreV1Client.Services("")
	serviceList, err := services.List(meta.ListOptions{})
	if err != nil {
		return nil, err
	}
	restClient := l.coreV1Client.RESTClient()

	var managedServices []string

	for _, svc := range serviceList.Items {
		if svc.Spec.Type != "LoadBalancer" {
			klog.V(3).Infof("%s is not type LoadBalancer, skipping.", svc.Name)
			continue
		}
		klog.Infof("%s is type LoadBalancer.", svc.Name)
		managedServices = append(managedServices, svc.Name)
		result, err := action(restClient, svc)
		if err != nil {
			klog.Errorf("%s", result)
			klog.Errorf("error patching service %s/%s: %s", svc.Namespace, svc.Name, err)
			continue
		}

	}
	return managedServices, nil
}

func (l *LoadBalancerEmulator) updateService(restClient rest.Interface, svc core.Service) ([]byte, error) {
	clusterIP := svc.Spec.ClusterIP
	ingresses := svc.Status.LoadBalancer.Ingress
	if len(ingresses) == 1 && ingresses[0].IP == clusterIP {
		return nil, nil
	}
	return l.updateServiceIP(restClient, svc, clusterIP)
}

func (l *LoadBalancerEmulator) updateServiceIP(restClient rest.Interface, svc core.Service, ip string) ([]byte, error) {
	if len(ip) == 0 {
		return nil, nil
	}
	klog.V(3).Infof("[%s] setting ClusterIP as the LoadBalancer Ingress", svc.Name)
	jsonPatch := fmt.Sprintf(`[{"op": "add", "path": "/status/loadBalancer/ingress", "value":  [ { "ip": "%s" } ] }]`, ip)
	patch := &Patch{
		Type:         types.JSONPatchType,
		ResourceName: svc.Name,
		NameSpaceSet: true,
		NameSpace:    svc.Namespace,
		Subresource:  "status",
		Resource:     "services",
		BodyContent:  jsonPatch,
	}
	request := l.patchConverter.convert(restClient, patch)
	result, err := l.requestSender.send(request)
	if err != nil {
		klog.Errorf("error patching %s with IP %s: %s", svc.Name, ip, err)
	} else {
		klog.Infof("Patched %s with IP %s", svc.Name, ip)
	}
	return result, err
}

func (l *LoadBalancerEmulator) cleanupService(restClient rest.Interface, svc core.Service) ([]byte, error) {
	ingresses := svc.Status.LoadBalancer.Ingress
	if len(ingresses) == 0 {
		return nil, nil
	}
	klog.V(3).Infof("[%s] cleanup: unset load balancer ingress", svc.Name)
	jsonPatch := `[{"op": "remove", "path": "/status/loadBalancer/ingress" }]`
	patch := &Patch{
		Type:         types.JSONPatchType,
		ResourceName: svc.Name,
		NameSpaceSet: true,
		NameSpace:    svc.Namespace,
		Subresource:  "status",
		Resource:     "services",
		BodyContent:  jsonPatch,
	}
	request := l.patchConverter.convert(restClient, patch)
	result, err := l.requestSender.send(request)
	klog.Infof("Removed load balancer ingress from %s.", svc.Name)
	return result, err

}

// NewLoadBalancerEmulator creates a new LoadBalancerEmulator
func NewLoadBalancerEmulator(corev1Client typed_core.CoreV1Interface) LoadBalancerEmulator {
	return LoadBalancerEmulator{
		coreV1Client:   corev1Client,
		requestSender:  &defaultRequestSender{},
		patchConverter: &defaultPatchConverter{},
	}
}

type defaultPatchConverter struct{}

func (c *defaultPatchConverter) convert(restClient rest.Interface, patch *Patch) *rest.Request {
	request := restClient.Patch(patch.Type)
	request.Name(patch.ResourceName)
	request.Resource(patch.Resource)
	request.SubResource(patch.Subresource)
	if patch.NameSpaceSet {
		request.Namespace(patch.NameSpace)
	}
	request.Body([]byte(patch.BodyContent))
	return request
}

type defaultRequestSender struct{}

func (r *defaultRequestSender) send(request *rest.Request) ([]byte, error) {
	return request.Do().Raw()
}
