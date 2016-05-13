package commandstest

import (
	"bytes"
	"io"

	"os"
)

var (
	stdout *os.File
)

func init() {
	stdout = os.Stdout
}

type StdoutGetter interface {
	Output() string
	Stop()
}

type stdoutCapturer struct {
	stdout *os.File
	output chan string
}

func NewStdoutGetter() StdoutGetter {
	r, w, _ := os.Pipe()
	os.Stdout = w

	output := make(chan string)
	go func() {
		var testOutput bytes.Buffer
		io.Copy(&testOutput, r)
		output <- testOutput.String()
	}()

	return &stdoutCapturer{
		stdout: w,
		output: output,
	}
}

func (c *stdoutCapturer) Output() string {
	c.stdout.Close()
	text := <-c.output
	close(c.output)
	return text
}

func (c *stdoutCapturer) Stop() {
	os.Stdout = stdout
}
