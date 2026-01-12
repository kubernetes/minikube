/*
Copyright 2022 The Kubernetes Authors All rights reserved.

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
