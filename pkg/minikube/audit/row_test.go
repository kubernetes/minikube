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

	"k8s.io/minikube/pkg/minikube/constants"
)

func TestRow(t *testing.T) {
	c := "start"
	a := "--alsologtostderr"
	p := "profile1"
	u := "user1"
	v := "v0.17.1"
	st := time.Now()
	stFormatted := st.Format(constants.TimeFormat)
	et := time.Now()
	etFormatted := et.Format(constants.TimeFormat)

	r := newRow(c, a, u, v, st, et, p)

	t.Run("TestNewRow", func(t *testing.T) {
		tests := []struct {
			key  string
			got  string
			want string
		}{
			{"command", r.command, c},
			{"args", r.args, a},
			{"profile", r.profile, p},
			{"user", r.user, u},
			{"version", r.version, v},
			{"startTime", r.startTime, stFormatted},
			{"endTime", r.endTime, etFormatted},
		}

		for _, tt := range tests {
			if tt.got != tt.want {
				t.Errorf("row.%s = %s; want %s", tt.key, tt.got, tt.want)
			}
		}
	})

	t.Run("TestType", func(t *testing.T) {
		got := r.Type()
		want := "io.k8s.sigs.minikube.audit"

		if got != want {
			t.Errorf("Type() = %s; want %s", got, want)
		}
	})

	t.Run("TestToMap", func(t *testing.T) {
		m := r.toMap()

		tests := []struct {
			key  string
			want string
		}{
			{"command", c},
			{"args", a},
			{"profile", p},
			{"user", u},
			{"version", v},
			{"startTime", stFormatted},
			{"endTime", etFormatted},
		}

		for _, tt := range tests {
			got := m[tt.key]
			if got != tt.want {
				t.Errorf("map[%s] = %s; want %s", tt.key, got, tt.want)
			}
		}
	})

	t.Run("TestToField", func(t *testing.T) {
		got := r.toFields()
		gotString := strings.Join(got, ",")
		want := []string{c, a, p, u, v, stFormatted, etFormatted}
		wantString := strings.Join(want, ",")

		if gotString != wantString {
			t.Errorf("toFields() = %s; want %s", gotString, wantString)
		}
	})

	t.Run("TestAssignFields", func(t *testing.T) {
		l := fmt.Sprintf(`{"data":{"args":"%s","command":"%s","endTime":"%s","profile":"%s","startTime":"%s","user":"%s","version":"v0.17.1"},"datacontenttype":"application/json","id":"bc6ec9d4-0d08-4b57-ac3b-db8d67774768","source":"https://minikube.sigs.k8s.io/","specversion":"1.0","type":"io.k8s.sigs.minikube.audit"}`, a, c, etFormatted, p, stFormatted, u)

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
			{"profile", r.profile, p},
			{"user", r.user, u},
			{"version", r.version, v},
			{"startTime", r.startTime, stFormatted},
			{"endTime", r.endTime, etFormatted},
		}

		for _, tt := range tests {
			if tt.got != tt.want {
				t.Errorf("singleEntry.%s = %s; want %s", tt.key, tt.got, tt.want)

			}
		}
	})
}
