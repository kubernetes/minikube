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
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
)

func TestEntry(t *testing.T) {
	c := "start"
	a := "--alsologtostderr"
	p := "profile1"
	u := "user1"
	st := time.Now()
	stFormatted := st.Format(constants.TimeFormat)
	et := time.Now()
	etFormatted := et.Format(constants.TimeFormat)

	// save current profile in case something depends on it
	oldProfile := viper.GetString(config.ProfileName)
	viper.Set(config.ProfileName, p)
	e := newEntry(c, a, u, st, et)
	viper.Set(config.ProfileName, oldProfile)

	t.Run("TestNewEntry", func(t *testing.T) {
		d := e.Data

		tests := []struct {
			key  string
			want string
		}{
			{"command", c},
			{"args", a},
			{"profile", p},
			{"user", u},
			{"startTime", stFormatted},
			{"endTime", etFormatted},
		}

		for _, tt := range tests {
			got := d[tt.key]

			if got != tt.want {
				t.Errorf("Data[%q] = %s; want %s", tt.key, got, tt.want)
			}
		}
	})

	t.Run("TestType", func(t *testing.T) {
		got := e.Type()
		want := "io.k8s.sigs.minikube.audit"

		if got != want {
			t.Errorf("Type() = %s; want %s", got, want)
		}
	})

	t.Run("TestToField", func(t *testing.T) {
		got := e.toFields()
		gotString := strings.Join(got, ",")
		want := []string{c, a, p, u, stFormatted, etFormatted}
		wantString := strings.Join(want, ",")

		if gotString != wantString {
			t.Errorf("toFields() = %s; want %s", gotString, wantString)
		}
	})
}
