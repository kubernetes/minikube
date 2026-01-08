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

package log

import "io"

type MachineLogger interface {
	SetDebug(debug bool)

	SetOutWriter(io.Writer)
	SetErrWriter(io.Writer)

	Debug(args ...interface{})
	Debugf(fmtString string, args ...interface{})

	Error(args ...interface{})
	Errorf(fmtString string, args ...interface{})

	Info(args ...interface{})
	Infof(fmtString string, args ...interface{})

	Warn(args ...interface{})
	Warnf(fmtString string, args ...interface{})

	History() []string
}
