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
var styles = map[string]style{
	"happy":         {Prefix: "😄  "},
	"success":       {Prefix: "✅  "},
	"failure":       {Prefix: "❌  "},
	"conflict":      {Prefix: "💥  ", LowPrefix: lowWarning},
	"fatal":         {Prefix: "💣  ", LowPrefix: lowError},
	"notice":        {Prefix: "📌  "},
	"ready":         {Prefix: "🏄  "},
	"running":       {Prefix: "🏃  "},
	"provisioning":  {Prefix: "🌱  "},
	"restarting":    {Prefix: "🔄  "},
	"reconfiguring": {Prefix: "📯  "},
	"stopping":      {Prefix: "✋  "},
	"stopped":       {Prefix: "🛑  "},
	"warning":       {Prefix: "⚠️  ", LowPrefix: lowWarning},
	"waiting":       {Prefix: "⌛  "},
	"waiting-pods":  {Prefix: "⌛  ", OmitNewline: true},
	"usage":         {Prefix: "💡  "},
	"launch":        {Prefix: "🚀  "},
	"sad":           {Prefix: "😿  "},
	"thumbs-up":     {Prefix: "👍  "},
	"option":        {Prefix: "    ▪ ", LowPrefix: lowIndent}, // Indented bullet
	"command":       {Prefix: "    ▪ ", LowPrefix: lowIndent}, // Indented bullet
	"log-entry":     {Prefix: "    "},                         // Indent
	"crushed":       {Prefix: "💔  "},
	"url":           {Prefix: "👉  ", LowPrefix: lowIndent},
	"documentation": {Prefix: "📘  "},
	"issues":        {Prefix: "⁉️   "},
	"issue":         {Prefix: "    ▪ ", LowPrefix: lowIndent}, // Indented bullet
	"check":         {Prefix: "✔️  "},

	// Specialized purpose styles
	"iso-download":      {Prefix: "💿  "},
	"file-download":     {Prefix: "💾  "},
	"caching":           {Prefix: "🤹  "},
	"starting-vm":       {Prefix: "🔥  "},
	"starting-none":     {Prefix: "🤹  "},
	"resetting":         {Prefix: "🔄  "},
	"deleting-host":     {Prefix: "🔥  "},
	"copying":           {Prefix: "✨  "},
	"connectivity":      {Prefix: "📶  "},
	"internet":          {Prefix: "🌐  "},
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
	"kubectl":           {Prefix: "💗  "},
	"meh":               {Prefix: "🙄  ", LowPrefix: lowWarning},
	"embarrassed":       {Prefix: "🤦  ", LowPrefix: lowWarning},
	"tip":               {Prefix: "💡  "},
	"unmount":           {Prefix: "🔥  "},
	"mount-options":     {Prefix: "💾  "},
	"fileserver":        {Prefix: "🚀  ", OmitNewline: true},
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
		return lowIndent
	}
	return lowBullet
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
