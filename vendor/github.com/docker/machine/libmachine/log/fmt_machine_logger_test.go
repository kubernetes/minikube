package log

import (
	"bufio"
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func captureOutput(testLogger MachineLogger, lambda func()) string {
	pipeReader, pipeWriter := io.Pipe()
	scanner := bufio.NewScanner(pipeReader)
	testLogger.SetOutWriter(pipeWriter)
	go lambda()
	scanner.Scan()
	return scanner.Text()
}

func captureError(testLogger MachineLogger, lambda func()) string {
	pipeReader, pipeWriter := io.Pipe()
	scanner := bufio.NewScanner(pipeReader)
	testLogger.SetErrWriter(pipeWriter)
	go lambda()
	scanner.Scan()
	return scanner.Text()
}

func TestSetDebugToTrue(t *testing.T) {
	testLogger := NewFmtMachineLogger().(*FmtMachineLogger)
	testLogger.SetDebug(true)
	assert.Equal(t, true, testLogger.debug)
}

func TestSetDebugToFalse(t *testing.T) {
	testLogger := NewFmtMachineLogger().(*FmtMachineLogger)
	testLogger.SetDebug(true)
	testLogger.SetDebug(false)
	assert.Equal(t, false, testLogger.debug)
}

func TestSetOut(t *testing.T) {
	testLogger := NewFmtMachineLogger().(*FmtMachineLogger)
	testLogger.SetOutWriter(ioutil.Discard)
	assert.Equal(t, ioutil.Discard, testLogger.outWriter)
}

func TestSetErr(t *testing.T) {
	testLogger := NewFmtMachineLogger().(*FmtMachineLogger)
	testLogger.SetErrWriter(ioutil.Discard)
	assert.Equal(t, ioutil.Discard, testLogger.errWriter)
}

func TestDebug(t *testing.T) {
	testLogger := NewFmtMachineLogger()
	testLogger.SetDebug(true)

	result := captureError(testLogger, func() { testLogger.Debug("debug") })

	assert.Equal(t, result, "debug")
}

func TestInfo(t *testing.T) {
	testLogger := NewFmtMachineLogger()

	result := captureOutput(testLogger, func() { testLogger.Info("info") })

	assert.Equal(t, result, "info")
}

func TestWarn(t *testing.T) {
	testLogger := NewFmtMachineLogger()

	result := captureOutput(testLogger, func() { testLogger.Warn("warn") })

	assert.Equal(t, result, "warn")
}

func TestError(t *testing.T) {
	testLogger := NewFmtMachineLogger()

	result := captureError(testLogger, func() { testLogger.Error("error") })

	assert.Equal(t, result, "error")
}

func TestEntriesAreCollected(t *testing.T) {
	testLogger := NewFmtMachineLogger()
	testLogger.Debug("debug")
	testLogger.Info("info")
	testLogger.Error("error")
	assert.Equal(t, 3, len(testLogger.History()))
	assert.Equal(t, "debug", testLogger.History()[0])
	assert.Equal(t, "info", testLogger.History()[1])
	assert.Equal(t, "error", testLogger.History()[2])
}
