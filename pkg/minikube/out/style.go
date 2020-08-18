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

package out

import (
	"bytes"
	"strings"
	"text/template"

	"github.com/golang/glog"
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
	Celebration:   {Prefix: "🎉  "},
	Check:         {Prefix: "✅  "},
	Command:       {Prefix: "    ▪ ", LowPrefix: lowIndent}, // Indented bullet
	Conflict:      {Prefix: "💥  ", LowPrefix: lowWarning},
	Confused:      {Prefix: "😕  "},
	Deleted:       {Prefix: "💀  "},
	Documentation: {Prefix: "📘  "},
	Empty:         {Prefix: "", LowPrefix: ""},
	FailureType:   {Prefix: "❌  "},
	FatalType:     {Prefix: "💣  ", LowPrefix: lowError},
	Happy:         {Prefix: "😄  "},
	Issue:         {Prefix: "    ▪ ", LowPrefix: lowIndent}, // Indented bullet
	Issues:        {Prefix: "⁉️   "},
	Launch:        {Prefix: "🚀  "},
	LogEntry:      {Prefix: "    "}, // Indent
	New:           {Prefix: "🆕  "},
	Notice:        {Prefix: "📌  "},
	Option:        {Prefix: "    ▪ ", LowPrefix: lowIndent}, // Indented bullet
	Pause:         {Prefix: "⏸️  "},
	Provisioning:  {Prefix: "🌱  "},
	Ready:         {Prefix: "🏄  "},
	Restarting:    {Prefix: "🔄  "},
	Running:       {Prefix: "🏃  "},
	Sad:           {Prefix: "😿  "},
	Shrug:         {Prefix: "🤷  "},
	Sparkle:       {Prefix: "✨  "},
	Stopped:       {Prefix: "🛑  "},
	Stopping:      {Prefix: "✋  "},
	SuccessType:   {Prefix: "✅  "},
	ThumbsDown:    {Prefix: "👎  "},
	ThumbsUp:      {Prefix: "👍  "},
	Unpause:       {Prefix: "⏯️  "},
	URL:           {Prefix: "👉  ", LowPrefix: lowIndent},
	Usage:         {Prefix: "💡  "},
	Waiting:       {Prefix: "⌛  "},
	Warning:       {Prefix: "❗  ", LowPrefix: lowWarning},
	Workaround:    {Prefix: "👉  ", LowPrefix: lowIndent},

	// Specialized purpose styles
	AddonDisable:     {Prefix: "🌑  "},
	AddonEnable:      {Prefix: "🌟  "},
	Caching:          {Prefix: "🤹  "},
	Celebrate:        {Prefix: "🎉  "},
	Connectivity:     {Prefix: "📶  "},
	Containerd:       {Prefix: "📦  "},
	ContainerRuntime: {Prefix: "🎁  "},
	Copying:          {Prefix: "✨  "},
	CRIO:             {Prefix: "🎁  "}, // This should be a snow-flake, but the emoji has a strange width on macOS
	DeletingHost:     {Prefix: "🔥  "},
	Docker:           {Prefix: "🐳  "},
	DryRun:           {Prefix: "🌵  "},
	Embarrassed:      {Prefix: "🤦  ", LowPrefix: lowWarning},
	Enabling:         {Prefix: "🔌  "},
	FileDownload:     {Prefix: "💾  "},
	Fileserver:       {Prefix: "🚀  ", OmitNewline: true},
	HealthCheck:      {Prefix: "🔎  "},
	Internet:         {Prefix: "🌐  "},
	ISODownload:      {Prefix: "💿  "},
	Kubectl:          {Prefix: "💗  "},
	Meh:              {Prefix: "🙄  ", LowPrefix: lowWarning},
	Mounting:         {Prefix: "📁  "},
	MountOptions:     {Prefix: "💾  "},
	Permissions:      {Prefix: "🔑  "},
	Provisioner:      {Prefix: "ℹ️  "},
	Pulling:          {Prefix: "🚜  "},
	Resetting:        {Prefix: "🔄  "},
	Shutdown:         {Prefix: "🛑  "},
	StartingNone:     {Prefix: "🤹  "},
	StartingVM:       {Prefix: "🔥  "},
	Tip:              {Prefix: "💡  "},
	Unmount:          {Prefix: "🔥  "},
	VerifyingNoLine:  {Prefix: "🤔  ", OmitNewline: true},
	Verifying:        {Prefix: "🤔  "},
	CNI:              {Prefix: "🔗  "},
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

// applyStyle translates the given string if necessary then adds any appropriate style prefix.
func applyStyle(style StyleEnum, useColor bool, format string) string {
	format = translate.T(format)

	s, ok := styles[style]
	if !s.OmitNewline {
		format += "\n"
	}

	// Similar to CSS styles, if no style matches, output an unformatted string.
	if !ok || JSON {
		return format
	}

	if !useColor {
		return applyPrefix(lowPrefix(s), format)
	}
	return applyPrefix(s.Prefix, format)
}

// ApplyTemplateFormatting applies formatting to the provided template
func ApplyTemplateFormatting(style StyleEnum, useColor bool, format string, a ...V) string {
	if a == nil {
		a = []V{{}}
	}
	format = applyStyle(style, useColor, format)

	var buf bytes.Buffer
	t, err := template.New(format).Parse(format)
	if err != nil {
		glog.Errorf("unable to parse %q: %v - returning raw string.", format, err)
		return format
	}
	err = t.Execute(&buf, a[0])
	if err != nil {
		glog.Errorf("unable to execute %s: %v - returning raw string.", format, err)
		return format
	}
	outStyled := buf.String()

	// escape any outstanding '%' signs so that they don't get interpreted
	// as a formatting directive down the line
	outStyled = strings.Replace(outStyled, "%", "%%", -1)

	return outStyled
}
