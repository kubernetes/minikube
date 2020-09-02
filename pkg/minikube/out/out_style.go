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

// Package out provides a mechanism for sending localized, stylized output to the console.
package out

import (
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/minikube/translate"
)

// Add a prefix to a string
func applyPrefix(prefix, format string) string {
	if prefix == "" {
		return format
	}
	return prefix + format
}

// applyStyle translates the given string if necessary then adds any appropriate style prefix.
func applyStyle(st style.Enum, useColor bool, format string) string {
	format = translate.T(format)

	s, ok := style.Config[st]
	if !s.OmitNewline {
		format += "\n"
	}

	// Similar to CSS styles, if no style matches, output an unformatted string.
	if !ok || JSON {
		return format
	}

	if !useColor {
		return applyPrefix(style.LowPrefix(s), format)
	}
	return applyPrefix(s.Prefix, format)
}

// stylized applies formatting to the provided template
func stylized(st style.Enum, useColor bool, format string, a ...V) string {
	if a == nil {
		a = []V{{}}
	}
	format = applyStyle(st, useColor, format)
	return Fmt(format, a...)
}
