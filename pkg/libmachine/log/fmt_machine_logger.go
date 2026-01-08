package log

import (
	"fmt"
	"io"
	"os"
)

type FmtMachineLogger struct {
	outWriter io.Writer
	errWriter io.Writer
	debug     bool
	history   *HistoryRecorder
}

// NewFmtMachineLogger creates a MachineLogger implementation used by the drivers
func NewFmtMachineLogger() MachineLogger {
	return &FmtMachineLogger{
		outWriter: os.Stdout,
		errWriter: os.Stderr,
		debug:     false,
		history:   NewHistoryRecorder(),
	}
}

func (ml *FmtMachineLogger) SetDebug(debug bool) {
	ml.debug = debug
}

func (ml *FmtMachineLogger) SetOutWriter(out io.Writer) {
	ml.outWriter = out
}

func (ml *FmtMachineLogger) SetErrWriter(err io.Writer) {
	ml.errWriter = err
}

func (ml *FmtMachineLogger) Debug(args ...interface{}) {
	ml.history.Record(args...)
	if ml.debug {
		_, _ = fmt.Fprintln(ml.errWriter, args...)
	}
}

func (ml *FmtMachineLogger) Debugf(fmtString string, args ...interface{}) {
	ml.history.Recordf(fmtString, args...)
	if ml.debug {
		_, _ = fmt.Fprintf(ml.errWriter, fmtString+"\n", args...)
	}
}

func (ml *FmtMachineLogger) Error(args ...interface{}) {
	ml.history.Record(args...)
	_, _ = fmt.Fprintln(ml.errWriter, args...)
}

func (ml *FmtMachineLogger) Errorf(fmtString string, args ...interface{}) {
	ml.history.Recordf(fmtString, args...)
	_, _ = fmt.Fprintf(ml.errWriter, fmtString+"\n", args...)
}

func (ml *FmtMachineLogger) Info(args ...interface{}) {
	ml.history.Record(args...)
	_, _ = fmt.Fprintln(ml.outWriter, args...)
}

func (ml *FmtMachineLogger) Infof(fmtString string, args ...interface{}) {
	ml.history.Recordf(fmtString, args...)
	_, _ = fmt.Fprintf(ml.outWriter, fmtString+"\n", args...)
}

func (ml *FmtMachineLogger) Warn(args ...interface{}) {
	ml.history.Record(args...)
	_, _ = fmt.Fprintln(ml.outWriter, args...)
}

func (ml *FmtMachineLogger) Warnf(fmtString string, args ...interface{}) {
	ml.history.Recordf(fmtString, args...)
	_, _ = fmt.Fprintf(ml.outWriter, fmtString+"\n", args...)
}

func (ml *FmtMachineLogger) History() []string {
	return ml.history.records
}
