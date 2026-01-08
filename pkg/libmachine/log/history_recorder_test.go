package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRecording(t *testing.T) {
	recorder := NewHistoryRecorder()
	recorder.Record("foo")
	recorder.Record("bar")
	recorder.Record("qix")
	assert.Equal(t, recorder.History(), []string{"foo", "bar", "qix"})
}

func TestFormattedRecording(t *testing.T) {
	recorder := NewHistoryRecorder()
	recorder.Recordf("%s, %s and %s", "foo", "bar", "qix")
	assert.Equal(t, recorder.History()[0], "foo, bar and qix")
}
