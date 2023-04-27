package runner

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
