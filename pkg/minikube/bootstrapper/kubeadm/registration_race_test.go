/*
Copyright 2026 The Kubernetes Authors All rights reserved.

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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"slices"
	"sync/atomic"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/minikube/pkg/minikube/command"
	"k8s.io/minikube/pkg/minikube/config"
)

// registrationRunner models the Node read required by kubectl label.
type registrationRunner struct {
	*command.FakeCommandRunner
	nodes    coreclient.NodeInterface
	labels   int
	nodeName string
}

func (r *registrationRunner) RunCmd(cmd *exec.Cmd) (*command.RunResult, error) {
	rr := &command.RunResult{Args: cmd.Args}
	if !slices.Contains(cmd.Args, "label") {
		return rr, fmt.Errorf("unexpected command: %s", rr.Command())
	}
	r.labels++
	// kubectl label first reads the named Node before applying its patch.
	if _, err := r.nodes.Get(context.Background(), r.nodeName, metav1.GetOptions{}); err != nil {
		rr.ExitCode = 1
		return rr, err
	}
	return rr, nil
}

// TestLabelAndUntaintNodeWaitsForWorkerRegistration reproduces a join that
// completes before the worker Node becomes visible. The first GET returns 404;
// the next GET publishes the Node, avoiding timing-dependent sleeps or a VM.
func TestLabelAndUntaintNodeWaitsForWorkerRegistration(t *testing.T) {
	const nodeName = "registration-race-m02"
	var reads atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v1/nodes/"+nodeName {
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if reads.Add(1) == 1 {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(&metav1.Status{
				TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "Status"},
				Status:   metav1.StatusFailure, Reason: metav1.StatusReasonNotFound, Code: http.StatusNotFound,
				Message: fmt.Sprintf("nodes %q not found", nodeName),
				Details: &metav1.StatusDetails{Kind: "nodes", Name: nodeName},
			})
			return
		}
		_ = json.NewEncoder(w).Encode(&corev1.Node{
			TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Node"},
			ObjectMeta: metav1.ObjectMeta{Name: nodeName},
		})
	}))
	defer server.Close()
	client, err := kubernetes.NewForConfig(&rest.Config{Host: server.URL, Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	runner := &registrationRunner{
		FakeCommandRunner: command.NewFakeCommandRunner(), nodes: client.CoreV1().Nodes(), nodeName: nodeName,
	}
	worker := config.Node{Name: "m02"}
	cfg := config.ClusterConfig{
		Name: "registration-race", Driver: "qemu2", Network: "socket_vmnet",
		Nodes: []config.Node{{Name: "m01", ControlPlane: true, IP: "127.0.0.1", Port: 8443}, worker},
	}
	bs := &Bootstrapper{c: runner, k8sClient: client}
	if err := bs.LabelAndUntaintNode(cfg, worker); err != nil {
		t.Fatalf("worker labeling raced registration: %v", err)
	}
	if reads.Load() != 3 || runner.labels != 1 {
		t.Fatalf("got %d API reads and %d label commands; want 3 and 1", reads.Load(), runner.labels)
	}
}
