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
	// Icon is an alternative icon
	Icon rune
	// OmitNewline omits a newline at the end of a message.
	OmitNewline bool
}

// styles is a map of style name to style struct
// For consistency, ensure that emojis added render with the same width across platforms.
var styles = map[StyleEnum]style{
	Empty:         {Prefix: "", LowPrefix: ""},
	Happy:         {Prefix: "😄  ", Icon: ''},
	SuccessType:   {Prefix: "✅  ", Icon: ''},
	FailureType:   {Prefix: "❌  ", Icon: ''},
	Conflict:      {Prefix: "💥  ", Icon: '', LowPrefix: lowWarning},
	FatalType:     {Prefix: "💣  ", Icon: 'ﮏ', LowPrefix: lowError},
	Notice:        {Prefix: "📌  ", Icon: ''},
	Ready:         {Prefix: "🏄  ", Icon: ''},
	Running:       {Prefix: "🏃  ", Icon: 'ﰌ'},
	Provisioning:  {Prefix: "🌱  ", Icon: ''},
	Restarting:    {Prefix: "🔄  ", Icon: 'ﰇ'},
	Reconfiguring: {Prefix: "📯  ", Icon: ''},
	Stopping:      {Prefix: "✋  ", Icon: 'ﭥ'},
	Stopped:       {Prefix: "🛑  ", Icon: 'ﭦ'},
	WarningType:   {Prefix: "⚠️  ", Icon: '', LowPrefix: lowWarning},
	Waiting:       {Prefix: "⌛  ", Icon: ''},
	Usage:         {Prefix: "💡  ", Icon: ''},
	Launch:        {Prefix: "🚀  ", Icon: ''},
	Sad:           {Prefix: "😿  ", Icon: ''},
	ThumbsUp:      {Prefix: "👍  ", Icon: ''},
	Option:        {Prefix: "    ▪ ", LowPrefix: lowIndent}, // Indented bullet
	Command:       {Prefix: "    ▪ ", LowPrefix: lowIndent}, // Indented bullet
	LogEntry:      {Prefix: "    "},                         // Indent
	Crushed:       {Prefix: "💔  ", Icon: ''},
	URL:           {Prefix: "👉  ", Icon: '', LowPrefix: lowIndent},
	Documentation: {Prefix: "📘  ", Icon: ''},
	Issues:        {Prefix: "⁉️   ", Icon: ''},
	Issue:         {Prefix: "    ▪ ", LowPrefix: lowIndent}, // Indented bullet
	Check:         {Prefix: "✅  ", Icon: ''},
	Celebration:   {Prefix: "🎉  ", Icon: '六'},
	Workaround:    {Prefix: "👉  ", Icon: '', LowPrefix: lowIndent},
	Sparkle:       {Prefix: "✨  ", Icon: ''},

	// Specialized purpose styles
	ISODownload:      {Prefix: "💿  ", Icon: '﫭'},
	FileDownload:     {Prefix: "💾  ", Icon: ''},
	Caching:          {Prefix: "🤹  ", Icon: ''},
	StartingVM:       {Prefix: "🔥  ", Icon: ''},
	StartingNone:     {Prefix: "🤹  ", Icon: ''},
	Provisioner:      {Prefix: "ℹ️   ", Icon: ''},
	Resetting:        {Prefix: "🔄  ", Icon: 'ﮦ'},
	DeletingHost:     {Prefix: "🔥  ", Icon: ''},
	Copying:          {Prefix: "✨  ", Icon: ''},
	Connectivity:     {Prefix: "📶  ", Icon: ''},
	Internet:         {Prefix: "🌐  ", Icon: ''},
	Mounting:         {Prefix: "📁  ", Icon: ''},
	Celebrate:        {Prefix: "🎉  ", Icon: '六'},
	ContainerRuntime: {Prefix: "🎁  ", Icon: ''},
	Docker:           {Prefix: "🐳  ", Icon: ''},
	CRIO:             {Prefix: "🎁  ", Icon: 'ﰕ'}, // This should be a snow-flake, but the emoji has a strange width on macOS
	Containerd:       {Prefix: "📦  ", Icon: ''},
	Permissions:      {Prefix: "🔑  ", Icon: '廬'},
	Enabling:         {Prefix: "🔌  ", Icon: ''},
	Shutdown:         {Prefix: "🛑  ", Icon: ''},
	Pulling:          {Prefix: "🚜  ", Icon: ''},
	Verifying:        {Prefix: "🤔  ", Icon: ''},
	VerifyingNoLine:  {Prefix: "🤔  ", Icon: '', OmitNewline: true},
	Kubectl:          {Prefix: "💗  ", Icon: ''},
	Meh:              {Prefix: "🙄  ", Icon: '', LowPrefix: lowWarning},
	Embarrassed:      {Prefix: "🤦  ", Icon: 'ﮙ', LowPrefix: lowWarning},
	Tip:              {Prefix: "💡  ", Icon: 'ﯦ'},
	Unmount:          {Prefix: "🔥  ", Icon: ''},
	MountOptions:     {Prefix: "💾  ", Icon: ''},
	Fileserver:       {Prefix: "🚀  ", Icon: '歷', OmitNewline: true},
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
func applyStyle(style StyleEnum, useColor bool, useIcons bool, format string) string {
	format = translate.T(format)

	s, ok := styles[style]
	if !s.OmitNewline {
		format += "\n"
	}

	// Similar to CSS styles, if no style matches, output an unformatted string.
	if !ok {
		return format
	}
	if !useColor {
		return applyPrefix(lowPrefix(s), format)
	}
	if useIcons && s.Icon != 0 {
		return applyPrefix(string(s.Icon)+"  ", format)
	}
	return applyPrefix(s.Prefix, format)
}

func applyTemplateFormatting(style StyleEnum, useColor bool, useIcons bool, format string, a ...V) string {
	if a == nil {
		a = []V{{}}
	}
	format = applyStyle(style, useColor, useIcons, format)

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
