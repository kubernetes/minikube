/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

package audit

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/run"
)

func TestRow(t *testing.T) {
	options := &run.CommandOptions{ProfileName: constants.DefaultClusterName}

	c := "start"
	a := "--alsologtostderr"
	u := "user1"
	v := "v0.17.1"
	st := time.Now()
	stFormatted := st.Format(constants.TimeFormat)
	et := time.Now()
	etFormatted := et.Format(constants.TimeFormat)
	id := uuid.New().String()

	r := newRow(c, a, u, v, st, id, options)
	r.endTime = etFormatted

	t.Run("NewRow", func(t *testing.T) {
		tests := []struct {
			key  string
			got  string
			want string
		}{
			{"command", r.command, c},
			{"args", r.args, a},
			{"profile", r.profile, options.ProfileName},
			{"user", r.user, u},
			{"version", r.version, v},
			{"startTime", r.startTime, stFormatted},
			{"id", r.id, id},
		}

		for _, tt := range tests {
			if tt.got != tt.want {
				t.Errorf("row.%s = %s; want %s", tt.key, tt.got, tt.want)
			}
		}
	})

	t.Run("Type", func(t *testing.T) {
		got := r.Type()
		want := "io.k8s.sigs.minikube.audit"

		if got != want {
			t.Errorf("Type() = %s; want %s", got, want)
		}
	})

	t.Run("toMap", func(t *testing.T) {
		m := r.toMap()

		tests := []struct {
			key  string
			want string
		}{
			{"command", c},
			{"args", a},
			{"profile", options.ProfileName},
			{"user", u},
			{"version", v},
			{"startTime", stFormatted},
			{"id", id},
		}

		for _, tt := range tests {
			got := m[tt.key]
			if got != tt.want {
				t.Errorf("map[%s] = %s; want %s", tt.key, got, tt.want)
			}
		}
	})

	t.Run("toFields", func(t *testing.T) {
		got := r.toFields()
		gotString := strings.Join(got, ",")
		want := []string{c, a, options.ProfileName, u, v, stFormatted, etFormatted}
		wantString := strings.Join(want, ",")

		if gotString != wantString {
			t.Errorf("toFields() = %s; want %s", gotString, wantString)
		}
	})

	t.Run("assignFields", func(t *testing.T) {
		l := fmt.Sprintf(
			`{"data":{"args":"%s","command":"%s","id":"%s","profile":"%s","startTime":"%s","user":"%s","version":"v0.17.1"},"datacontenttype":"application/json","id":"bc6ec9d4-0d08-4b57-ac3b-db8d67774768","source":"https://minikube.sigs.k8s.io/","specversion":"1.0","type":"io.k8s.sigs.minikube.audit"}`,
			a, c, id, options.ProfileName, stFormatted, u)

		r := &row{}
		if err := json.Unmarshal([]byte(l), r); err != nil {
			t.Fatalf("failed to unmarshal log: %v", err)
		}

		r.assignFields()

		tests := []struct {
			key  string
			got  string
			want string
		}{
			{"command", r.command, c},
			{"args", r.args, a},
			{"profile", r.profile, options.ProfileName},
			{"user", r.user, u},
			{"version", r.version, v},
			{"startTime", r.startTime, stFormatted},
			{"id", r.id, id},
		}

		for _, tt := range tests {
			if tt.got != tt.want {
				t.Errorf("singleEntry.%s = %s; want %s", tt.key, tt.got, tt.want)

			}
		}
	})
}
