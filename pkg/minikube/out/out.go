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
	"bytes"
	"fmt"
	"html"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Delta456/box-cli-maker/v2"
	"github.com/briandowns/spinner"
	"github.com/mattn/go-isatty"

	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/localpath"
	"k8s.io/minikube/pkg/minikube/out/register"
	"k8s.io/minikube/pkg/minikube/style"
	"k8s.io/minikube/pkg/minikube/translate"
)

// By design, this package uses global references to language and output objects, in preference
// to passing a console object throughout the code base. Typical usage is:
//
// out.SetOutFile(os.Stdout)
// out.String("Starting up!")
// out.Step(style.StatusChange, "Configuring things")

// out.SetErrFile(os.Stderr)
// out.Fatal("Oh no, everything failed.")

// NOTE: If you do not want colorized output, set MINIKUBE_IN_STYLE=false in your environment.

var (
	// silent will disable all output, if called from a script. Set using SetSilent()
	silent bool
	// outFile is where Out* functions send output to. Set using SetOutFile()
	outFile fdWriter
	// errFile is where Err* functions send output to. Set using SetErrFile()
	errFile fdWriter
	// useColor is whether or not color output should be used, updated by Set*Writer.
	useColor = false
	// OverrideEnv is the environment variable used to override color/emoji usage
	OverrideEnv = "MINIKUBE_IN_STYLE"
	// JSON is whether or not we should output stdout in JSON format. Set using SetJSON()
	JSON = false
	// spin is spinner showed at starting minikube
	spin = spinner.New(spinner.CharSets[style.SpinnerCharacter], 100*time.Millisecond)
)

// MaxLogEntries controls the number of log entries to show for each source
const (
	MaxLogEntries = 3
)

// fdWriter is the subset of file.File that implements io.Writer and Fd()
type fdWriter interface {
	io.Writer
	Fd() uintptr
}

// V is a convenience wrapper for templating, it represents the variable key/value pair.
type V map[string]interface{}

// Step writes a stylized and templated message to stdout
func Step(st style.Enum, format string, a ...V) {
	if st == style.Option {
		Infof(format, a...)
		return
	}
	outStyled, _ := stylized(st, useColor, format, a...)
	if JSON {
		register.PrintStep(outStyled)
		return
	}
	register.RecordStep(outStyled)
	Styled(st, format, a...)
}

// Styled writes a stylized and templated message to stdout
func Styled(st style.Enum, format string, a ...V) {
	if JSON || st == style.Option {
		Infof(format, a...)
		return
	}
	outStyled, spinner := stylized(st, useColor, format, a...)
	if spinner {
		spinnerString(outStyled)
	} else {
		String(outStyled)
	}
}

func boxedCommon(printFunc func(format string, a ...interface{}), format string, a ...V) {
	box := box.New(box.Config{Py: 1, Px: 4, Type: "Round"})
	if useColor {
		box.Config.Color = "Red"
	}
	str := Sprintf(style.None, format, a...)
	printFunc(box.String("", strings.TrimSpace(str)))
}

// Boxed writes a stylized and templated message in a box to stdout
func Boxed(format string, a ...V) {
	boxedCommon(String, format, a...)
}

// BoxedErr writes a stylized and templated message in a box to stderr
func BoxedErr(format string, a ...V) {
	boxedCommon(Err, format, a...)
}

// Sprintf is used for returning the string (doesn't write anything)
func Sprintf(st style.Enum, format string, a ...V) string {
	outStyled, _ := stylized(st, useColor, format, a...)
	return outStyled
}

// Infof is used for informational logs (options, env variables, etc)
func Infof(format string, a ...V) {
	outStyled, _ := stylized(style.Option, useColor, format, a...)
	if JSON {
		register.PrintInfo(outStyled)
		return
	}
	String(outStyled)
}

// String writes a basic formatted string to stdout
func String(format string, a ...interface{}) {
	// Flush log buffer so that output order makes sense
	klog.Flush()

	if silent {
		klog.Infof(format, a...)
		return
	}

	if outFile == nil {
		klog.Warningf("[unset outFile]: %s", fmt.Sprintf(format, a...))
		return
	}
	klog.Infof(format, a...)
	// if spin is active from a previous step, it will stop spinner displaying
	if spin.Active() {
		spin.Stop()
	}

	Output(outFile, format, a...)
}

// Output writes a basic formatted string
func Output(file fdWriter, format string, a ...interface{}) {
	_, err := fmt.Fprintf(file, format, a...)
	if err != nil {
		klog.Errorf("Fprintf failed: %v", err)
	}
}

// spinnerString writes a basic formatted string to stdout with spinner character
func spinnerString(format string, a ...interface{}) {
	// Flush log buffer so that output order makes sense
	klog.Flush()

	if outFile == nil {
		klog.Warningf("[unset outFile]: %s", fmt.Sprintf(format, a...))
		return
	}

	klog.Infof(format, a...)
	// if spin is active from a previous step, it will stop spinner displaying
	if spin.Active() {
		spin.Stop()
	}
	Output(outFile, format, a...)
	// Start spinning at the end of the printed line
	spin.Start()
}

// Ln writes a basic formatted string with a newline to stdout
func Ln(format string, a ...interface{}) {
	if JSON {
		klog.Warningf("please use out.T to log steps in JSON")
		return
	}
	String(format+"\n", a...)
}

// ErrT writes a stylized and templated error message to stderr
func ErrT(st style.Enum, format string, a ...V) {
	errStyled, _ := stylized(st, useColor, format, a...)
	Err(errStyled)
}

// Err writes a basic formatted string to stderr
func Err(format string, a ...interface{}) {
	if JSON {
		register.PrintError(format)
		return
	}
	register.RecordError(format)

	if errFile == nil {
		klog.Errorf("[unset errFile]: %s", fmt.Sprintf(format, a...))
		return
	}

	klog.Warningf(format, a...)

	// if spin is active from a previous step, it will stop spinner displaying
	if spin.Active() {
		spin.Stop()
	}
	Output(errFile, format, a...)
}

// ErrLn writes a basic formatted string with a newline to stderr
func ErrLn(format string, a ...interface{}) {
	Err(format+"\n", a...)
}

// SuccessT is a shortcut for writing a templated success message to stdout
func SuccessT(format string, a ...V) {
	Step(style.Success, format, a...)
}

// FatalT is a shortcut for writing a templated fatal message to stderr
func FatalT(format string, a ...V) {
	ErrT(style.Fatal, format, a...)
}

// WarningT is a shortcut for writing a templated warning message to stderr
func WarningT(format string, a ...V) {
	if JSON {
		// if spin is active from a previous step, it will stop spinner displaying
		if spin.Active() {
			spin.Stop()
		}
		st, _ := stylized(style.Warning, useColor, format, a...)
		register.PrintWarning(st)
		return
	}
	ErrT(style.Warning, format, a...)
}

// FailureT is a shortcut for writing a templated failure message to stderr
func FailureT(format string, a ...V) {
	ErrT(style.Failure, format, a...)
}

// IsTerminal returns whether we have a terminal or not
func IsTerminal(w fdWriter) bool {
	return isatty.IsTerminal(w.Fd())
}

// SetSilent configures whether output is disabled or not
func SetSilent(q bool) {
	klog.Infof("Setting silent to %v", q)
	silent = q
}

// SetOutFile configures which writer standard output goes to.
func SetOutFile(w fdWriter) {
	klog.Infof("Setting OutFile to fd %d ...", w.Fd())
	outFile = w
	useColor = wantsColor(w)
}

// SetJSON configures printing to STDOUT in JSON
func SetJSON(j bool) {
	klog.Infof("Setting JSON to %v", j)
	JSON = j
}

// SetErrFile configures which writer error output goes to.
func SetErrFile(w fdWriter) {
	klog.Infof("Setting ErrFile to fd %d...", w.Fd())
	errFile = w
	useColor = wantsColor(w)
}

// wantsColor determines if the user might want colorized output.
func wantsColor(w fdWriter) bool {
	// First process the environment: we allow users to force colors on or off.
	//
	// MINIKUBE_IN_STYLE=[1, T, true, TRUE]
	// MINIKUBE_IN_STYLE=[0, f, false, FALSE]
	//
	// If unset, we try to automatically determine suitability from the environment.
	val := os.Getenv(OverrideEnv)
	if val != "" {
		klog.Infof("%s=%q\n", OverrideEnv, os.Getenv(OverrideEnv))
		override, err := strconv.ParseBool(val)
		if err != nil {
			// That's OK, we will just fall-back to automatic detection.
			klog.Errorf("ParseBool(%s): %v", OverrideEnv, err)
		} else {
			return override
		}
	}

	// New Windows Terminal
	if os.Getenv("WT_SESSION") != "" {
		return true
	}

	term := os.Getenv("TERM")
	colorTerm := os.Getenv("COLORTERM")
	// Example: term-256color
	if !strings.Contains(term, "color") && !strings.Contains(colorTerm, "truecolor") && !strings.Contains(colorTerm, "24bit") && !strings.Contains(colorTerm, "yes") {
		klog.Infof("TERM=%s,COLORTERM=%s, which probably does not support color", term, colorTerm)
		return false
	}

	isT := IsTerminal(w)
	klog.Infof("isatty.IsTerminal(%d) = %v\n", w.Fd(), isT)
	return isT
}

// LogEntries outputs an error along with any important log entries.
func LogEntries(msg string, err error, entries map[string][]string) {
	displayError(msg, err)

	for name, lines := range entries {
		Step(style.Failure, "Problems detected in {{.entry}}:", V{"entry": name})
		if len(lines) > MaxLogEntries {
			lines = lines[:MaxLogEntries]
		}
		for _, l := range lines {
			Step(style.LogEntry, l)
		}
	}
}

// displayError prints the error and displays the standard minikube error messaging
func displayError(msg string, err error) {
	klog.Warningf(fmt.Sprintf("%s: %v", msg, err))
	if JSON {
		FatalT("{{.msg}}: {{.err}}", V{"msg": translate.T(msg), "err": err})
		return
	}
	// use Warning because Error will display a duplicate message to stderr
	ErrT(style.Empty, "")
	FatalT("{{.msg}}: {{.err}}", V{"msg": translate.T(msg), "err": err})
	ErrT(style.Empty, "")
	displayGitHubIssueMessage()
}

func latestLogFilePath() (string, error) {
	if len(os.Args) < 2 {
		return "", fmt.Errorf("unable to detect command")
	}
	cmd := os.Args[1]
	if cmd == "start" {
		return localpath.LastStartLog(), nil
	}

	tmpdir := os.TempDir()
	files, err := ioutil.ReadDir(tmpdir)
	if err != nil {
		return "", fmt.Errorf("failed to get list of files in tempdir: %v", err)
	}
	var lastModName string
	var lastModTime time.Time
	for _, file := range files {
		if !strings.Contains(file.Name(), "minikube_") {
			continue
		}
		if !lastModTime.IsZero() && lastModTime.After(file.ModTime()) {
			continue
		}
		lastModName = file.Name()
		lastModTime = file.ModTime()
	}
	fullPath := filepath.Join(tmpdir, lastModName)

	return fullPath, nil
}

func displayGitHubIssueMessage() {
	logPath, err := latestLogFilePath()
	if err != nil {
		klog.Warningf("failed to diplay GitHub issue message: %v", err)
	}

	msg := Sprintf(style.Sad, "If the above advice does not help, please let us know:")
	msg += Sprintf(style.URL, "https://github.com/kubernetes/minikube/issues/new/choose\n")
	msg += Sprintf(style.Empty, "Please attach the following file to the GitHub issue:")
	msg += Sprintf(style.Empty, "- {{.logPath}}", V{"logPath": logPath})

	BoxedErr(msg)
}

// applyTmpl applies formatting
func applyTmpl(format string, a ...V) string {
	if len(a) == 0 {
		klog.Warningf("no arguments passed for %q - returning raw string", format)
		return format
	}

	var buf bytes.Buffer
	t, err := template.New(format).Parse(format)
	if err != nil {
		klog.Errorf("unable to parse %q: %v - returning raw string.", format, err)
		return format
	}
	err = t.Execute(&buf, a[0])
	if err != nil {
		klog.Errorf("unable to execute %s: %v - returning raw string.", format, err)
		return format
	}
	out := buf.String()

	// Return quotes back to normal
	out = html.UnescapeString(out)

	// escape any outstanding '%' signs so that they don't get interpreted
	// as a formatting directive down the line
	out = strings.ReplaceAll(out, "%", "%%")
	// avoid doubling up in case this function is called multiple times
	out = strings.ReplaceAll(out, "%%%%", "%%")
	return out
}

// Fmt applies formatting and translation
func Fmt(format string, a ...V) string {
	format = translate.T(format)
	if len(a) == 0 {
		return format
	}
	return applyTmpl(format, a...)
}
