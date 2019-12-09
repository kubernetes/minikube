/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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

package command

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestTeePrefix(t *testing.T) {
	var in bytes.Buffer
	var out bytes.Buffer
	var logged strings.Builder

	logSink := func(format string, args ...interface{}) {
		logged.WriteString("(" + fmt.Sprintf(format, args...) + ")")
	}

	// Simulate the primary use case: tee in the background. This also helps avoid I/O races.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		if err := teePrefix(":", &in, &out, logSink); err != nil {
			t.Errorf("teePrefix: %v", err)
		}
		wg.Done()
	}()

	in.Write([]byte("goo"))
	in.Write([]byte("\n"))
	in.Write([]byte("g\r\n\r\n"))
	in.Write([]byte("le"))
	wg.Wait()

	gotBytes := out.Bytes()
	wantBytes := []byte("goo\ng\r\n\r\nle")
	if !bytes.Equal(gotBytes, wantBytes) {
		t.Errorf("output=%q, want: %q", gotBytes, wantBytes)
	}

	gotLog := logged.String()
	wantLog := "(:goo)(:g)(:le)"
	if gotLog != wantLog {
		t.Errorf("log=%q, want: %q", gotLog, wantLog)
	}
}
