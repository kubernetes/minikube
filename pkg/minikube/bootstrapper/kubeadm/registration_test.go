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
	"errors"
	"fmt"
	"testing"
	"time"

	core "k8s.io/api/core/v1"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
)

// TestWaitForRegisteredNode exercises delayed registration with ordered API outcomes.
func TestWaitForRegisteredNode(t *testing.T) {
	for name, initialErr := range map[string]error{
		"missing":        apierr.NewNotFound(schema.GroupResource{Resource: "nodes"}, "synthetic-worker"),
		"timeout":        apierr.NewTimeoutError("synthetic timeout", 1),
		"server timeout": apierr.NewServerTimeout(schema.GroupResource{Resource: "nodes"}, "get", 1),
		"unavailable":    apierr.NewServiceUnavailable("synthetic pressure"),
		"rate limited":   apierr.NewTooManyRequests("synthetic pressure", 1),
	} {
		t.Run(name, func(t *testing.T) {
			c := fake.NewClientset()
			calls := 0
			c.PrependReactor("get", "nodes", func(action clienttesting.Action) (bool, runtime.Object, error) {
				get := action.(clienttesting.GetAction)
				if get.GetName() != "synthetic-worker" {
					return true, nil, fmt.Errorf("unexpected worker name")
				}
				calls++
				if calls == 1 {
					return true, nil, initialErr
				}
				return true, &core.Node{ObjectMeta: meta.ObjectMeta{Name: get.GetName()}}, nil
			})
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			if err := waitForRegisteredNode(ctx, c.CoreV1().Nodes(), "synthetic-worker", time.Nanosecond); err != nil {
				t.Fatal(err)
			}
			if calls != 2 {
				t.Fatalf("got %d reads, want transient failure then registered", calls)
			}
		})
	}
}

// TestWaitForRegisteredNodeRejectsPermanentErrors preserves authorization failures.
func TestWaitForRegisteredNodeRejectsPermanentErrors(t *testing.T) {
	c := fake.NewClientset()
	calls := 0
	c.PrependReactor("get", "nodes", func(clienttesting.Action) (bool, runtime.Object, error) {
		calls++
		return true, nil, apierr.NewForbidden(schema.GroupResource{Resource: "nodes"}, "synthetic-worker", errors.New("synthetic denial"))
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	err := waitForRegisteredNode(ctx, c.CoreV1().Nodes(), "synthetic-worker", time.Nanosecond)
	if !apierr.IsForbidden(err) || calls != 1 {
		t.Fatalf("got err=%v calls=%d, want one permanent denial", err, calls)
	}
}

// TestWaitForRegisteredNodeCancellation preserves the caller's finite ownership.
func TestWaitForRegisteredNodeCancellation(t *testing.T) {
	c := fake.NewClientset()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := waitForRegisteredNode(ctx, c.CoreV1().Nodes(), "synthetic-worker", time.Nanosecond); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want cancellation", err)
	}
}
