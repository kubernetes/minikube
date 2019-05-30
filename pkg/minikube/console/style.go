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

package console

import (
	"strings"

	"golang.org/x/text/message"
	"golang.org/x/text/number"
	"k8s.io/minikube/pkg/minikube/translate"
)

var (
	// lowBullet is a bullet-point prefix for low-fi mode
	lowBullet = "* "
	// lowBullet is an indented bullet-point prefix for low-fi mode
	lowIndent = "  - "
	// lowBullet is a warning prefix for low-fi mode
	lowWarning = "! "
	// lowBullet is an error prefix for low-fi mode
	lowError = "X "
)

// style describes how to stylize a message.
type style struct {
	// Prefix is a string to place in the beginning of a message
	Prefix string
	// LowPrefix is the 7-bit compatible prefix we fallback to for less-awesome terminals
	LowPrefix string
	// OmitNewline omits a newline at the end of a message.
	OmitNewline bool
}

// styles is a map of style name to style struct
// For consistency, ensure that emojis added render with the same width across platforms.
var styles = map[StyleEnum]style{
	Happy:         {Prefix: "😄  "},
	SuccessType:   {Prefix: "✅  "},
	FailureType:   {Prefix: "❌  "},
	Conflict:      {Prefix: "💥  ", LowPrefix: lowWarning},
	FatalType:     {Prefix: "💣  ", LowPrefix: lowError},
	Notice:        {Prefix: "📌  "},
	Ready:         {Prefix: "🏄  "},
	Running:       {Prefix: "🏃  "},
	Provisioning:  {Prefix: "🌱  "},
	Restarting:    {Prefix: "🔄  "},
	Reconfiguring: {Prefix: "📯  "},
	Stopping:      {Prefix: "✋  "},
	Stopped:       {Prefix: "🛑  "},
	WarningType:   {Prefix: "⚠️  ", LowPrefix: lowWarning},
	Waiting:       {Prefix: "⌛  "},
	WaitingPods:   {Prefix: "⌛  ", OmitNewline: true},
	Usage:         {Prefix: "💡  "},
	Launch:        {Prefix: "🚀  "},
	Sad:           {Prefix: "😿  "},
	ThumbsUp:      {Prefix: "👍  "},
	Option:        {Prefix: "    ▪ ", LowPrefix: lowIndent}, // Indented bullet
	Command:       {Prefix: "    ▪ ", LowPrefix: lowIndent}, // Indented bullet
	LogEntry:      {Prefix: "    "},                         // Indent
	Crushed:       {Prefix: "💔  "},
	URL:           {Prefix: "👉  ", LowPrefix: lowIndent},
	Documentation: {Prefix: "📘  "},
	Issues:        {Prefix: "⁉️   "},
	Issue:         {Prefix: "    ▪ ", LowPrefix: lowIndent}, // Indented bullet
	Check:         {Prefix: "✔️  "},

	// Specialized purpose styles
	ISODownload:      {Prefix: "💿  "},
	FileDownload:     {Prefix: "💾  "},
	Caching:          {Prefix: "🤹  "},
	StartingVM:       {Prefix: "🔥  "},
	StartingNone:     {Prefix: "🤹  "},
	Resetting:        {Prefix: "🔄  "},
	DeletingHost:     {Prefix: "🔥  "},
	Copying:          {Prefix: "✨  "},
	Connectivity:     {Prefix: "📶  "},
	Internet:         {Prefix: "🌐  "},
	Mounting:         {Prefix: "📁  "},
	Celebrate:        {Prefix: "🎉  "},
	ContainerRuntime: {Prefix: "🎁  "},
	Docker:           {Prefix: "🐳  "},
	CRIO:             {Prefix: "🎁  "}, // This should be a snow-flake, but the emoji has a strange width on macOS
	Containerd:       {Prefix: "📦  "},
	Permissions:      {Prefix: "🔑  "},
	Enabling:         {Prefix: "🔌  "},
	Shutdown:         {Prefix: "🛑  "},
	Pulling:          {Prefix: "🚜  "},
	Verifying:        {Prefix: "🤔  "},
	VerifyingNoLine:  {Prefix: "🤔  ", OmitNewline: true},
	Kubectl:          {Prefix: "💗  "},
	Meh:              {Prefix: "🙄  ", LowPrefix: lowWarning},
	Embarrassed:      {Prefix: "🤦  ", LowPrefix: lowWarning},
	Tip:              {Prefix: "💡  "},
	Unmount:          {Prefix: "🔥  "},
	MountOptions:     {Prefix: "💾  "},
	Fileserver:       {Prefix: "🚀  ", OmitNewline: true},
}

// Add a prefix to a string
func applyPrefix(prefix, format string) string {
	if prefix == "" {
		return format
	}
	// TODO(tstromberg): Ensure compatibility with RTL languages.
	return prefix + format
}

// lowPrefix returns a 7-bit compatible prefix for a style
func lowPrefix(s style) string {
	if s.LowPrefix != "" {
		return s.LowPrefix
	}
	if strings.HasPrefix(s.Prefix, "  ") {
		return lowIndent
	}
	return lowBullet
}

// Apply styling to a format string
func applyStyle(style StyleEnum, useColor bool, format string, a ...interface{}) string {
	p := message.NewPrinter(translate.GetPreferredLanguage())
	for i, x := range a {
		if _, ok := x.(int); ok {
			a[i] = number.Decimal(x, number.NoSeparator())
		}
	}
	format = translate.Translate(format)
	out := p.Sprintf(format, a...)

	s, ok := styles[style]
	if !s.OmitNewline {
		out += "\n"
	}

	// Similar to CSS styles, if no style matches, output an unformatted string.
	if !ok {
		return p.Sprintf(format, a...)
	}

	if !useColor {
		return applyPrefix(lowPrefix(s), out)
	}
	return applyPrefix(s.Prefix, out)
}
