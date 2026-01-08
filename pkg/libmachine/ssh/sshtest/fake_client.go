package sshtest

import "io"

type CmdResult struct {
	Out string
	Err error
}

type FakeClient struct {
	ActivatedShell []string
	Outputs        map[string]CmdResult
}

func (fsc *FakeClient) Output(command string) (string, error) {
	outerr := fsc.Outputs[command]
	return outerr.Out, outerr.Err
}

func (fsc *FakeClient) Shell(args ...string) error {
	fsc.ActivatedShell = args
	return nil
}

func (fsc *FakeClient) Start(command string) (io.ReadCloser, io.ReadCloser, error) {
	return nil, nil, nil
}

func (fsc *FakeClient) Wait() error {
	return nil
}
