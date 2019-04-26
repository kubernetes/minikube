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
	"fmt"
	"strings"

	"golang.org/x/text/message"
	"golang.org/x/text/number"
)

var (
	defaultLowPrefix       = "-   "
	defautlLowIndentPrefix = "    - "
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
var styles = map[string]style{
	"happy":         {Prefix: "😄  ", LowPrefix: "o   "},
	"success":       {Prefix: "✅  "},
	"failure":       {Prefix: "❌  ", LowPrefix: "X   "},
	"conflict":      {Prefix: "💥  ", LowPrefix: "x   "},
	"fatal":         {Prefix: "💣  ", LowPrefix: "!   "},
	"notice":        {Prefix: "📌  ", LowPrefix: "*   "},
	"ready":         {Prefix: "🏄  ", LowPrefix: "=   "},
	"running":       {Prefix: "🏃  ", LowPrefix: ":   "},
	"provisioning":  {Prefix: "🌱  ", LowPrefix: ">   "},
	"restarting":    {Prefix: "🔄  ", LowPrefix: ":   "},
	"reconfiguring": {Prefix: "📯  ", LowPrefix: ":   "},
	"stopping":      {Prefix: "✋  ", LowPrefix: ":   "},
	"stopped":       {Prefix: "🛑  "},
	"warning":       {Prefix: "⚠️  ", LowPrefix: "!   "},
	"waiting":       {Prefix: "⌛  ", LowPrefix: ":   "},
	"waiting-pods":  {Prefix: "⌛  ", LowPrefix: ":   ", OmitNewline: true},
	"usage":         {Prefix: "💡  "},
	"launch":        {Prefix: "🚀  "},
	"sad":           {Prefix: "😿  ", LowPrefix: "*   "},
	"thumbs-up":     {Prefix: "👍  "},
	"option":        {Prefix: "    ▪ "}, // Indented bullet
	"command":       {Prefix: "    ▪ "}, // Indented bullet
	"log-entry":     {Prefix: "    "},   // Indent
	"crushed":       {Prefix: "💔  "},
	"url":           {Prefix: "👉  "},
	"documentation": {Prefix: "📘  "},
	"issues":        {Prefix: "⁉️   "},
	"issue":         {Prefix: "    ▪ "}, // Indented bullet
	"check":         {Prefix: "✔️  "},

	// Specialized purpose styles
	"iso-download":      {Prefix: "💿  ", LowPrefix: "@   "},
	"file-download":     {Prefix: "💾  ", LowPrefix: "@   "},
	"caching":           {Prefix: "🤹  ", LowPrefix: "$   "},
	"starting-vm":       {Prefix: "🔥  ", LowPrefix: ">   "},
	"starting-none":     {Prefix: "🤹  ", LowPrefix: ">   "},
	"resetting":         {Prefix: "🔄  ", LowPrefix: "#   "},
	"deleting-host":     {Prefix: "🔥  ", LowPrefix: "x   "},
	"copying":           {Prefix: "✨  "},
	"connectivity":      {Prefix: "📶  "},
	"internet":          {Prefix: "🌐  ", LowPrefix: "o   "},
	"mounting":          {Prefix: "📁  "},
	"celebrate":         {Prefix: "🎉  "},
	"container-runtime": {Prefix: "🎁  "},
	"Docker":            {Prefix: "🐳  "},
	"CRI-O":             {Prefix: "🎁  "}, // This should be a snow-flake, but the emoji has a strange width on macOS
	"containerd":        {Prefix: "📦  "},
	"permissions":       {Prefix: "🔑  "},
	"enabling":          {Prefix: "🔌  "},
	"shutdown":          {Prefix: "🛑  "},
	"pulling":           {Prefix: "🚜  "},
	"verifying":         {Prefix: "🤔  "},
	"verifying-noline":  {Prefix: "🤔  ", OmitNewline: true},
	"kubectl":           {Prefix: "💗  ", LowPrefix: "+   "},
	"meh":               {Prefix: "🙄  ", LowPrefix: "?   "},
	"embarrassed":       {Prefix: "🤦  ", LowPrefix: "*   "},
	"tip":               {Prefix: "💡  ", LowPrefix: "i   "},
	"unmount":           {Prefix: "🔥  ", LowPrefix: "x   "},
	"mount-options":     {Prefix: "💾  ", LowPrefix: "o   "},
	"fileserver":        {Prefix: "🚀  ", LowPrefix: "@   ", OmitNewline: true},
}

// Add a prefix to a string
func applyPrefix(prefix, format string) string {
	if prefix == "" {
		return format
	}
	// TODO(tstromberg): Ensure compatibility with RTL languages.
	return prefix + format
}

func hasStyle(style string) bool {
	_, exists := styles[style]
	return exists
}

// lowPrefix returns a 7-bit compatible prefix for a style
func lowPrefix(s style) string {
	if s.LowPrefix != "" {
		return s.LowPrefix
	}
	if strings.HasPrefix(s.Prefix, "  ") {
		return defautlLowIndentPrefix
	}
	return defaultLowPrefix
}

// Apply styling to a format string
func applyStyle(style string, useColor bool, format string, a ...interface{}) (string, error) {
	p := message.NewPrinter(preferredLanguage)
	for i, x := range a {
		if _, ok := x.(int); ok {
			a[i] = number.Decimal(x, number.NoSeparator())
		}
	}
	out := p.Sprintf(format, a...)

	s, ok := styles[style]
	if !s.OmitNewline {
		out += "\n"
	}

	// Similar to CSS styles, if no style matches, output an unformatted string.
	if !ok {
		return p.Sprintf(format, a...), fmt.Errorf("unknown style: %q", style)
	}

	if !useColor {
		return applyPrefix(lowPrefix(s), out), nil
	}
	return applyPrefix(s.Prefix, out), nil
}
