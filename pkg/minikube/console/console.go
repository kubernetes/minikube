// Package console provides a mechanism for sending localized, stylized output to the console.
package console

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/golang/glog"
	isatty "github.com/mattn/go-isatty"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// By design, this package uses global references to language and output objects, in preference
// to passing a console object throughout the code base. Typical usage is:
//
// console.SetOutFile(os.Stdout)
// console.Out("Starting up!")
// console.OutStyle("status-change", "Configuring things")

// console.SetErrFile(os.Stderr)
// console.Fatal("Oh no, everything failed.")

var (
	// outFile is where Out* functions send output to. Set using SetOutFile()
	outFile fdWriter
	// errFile is where Err* functions send output to. Set using SetErrFile()
	errFile fdWriter
	// preferredLanguage is the default language messages will be output in
	preferredLanguage = language.AmericanEnglish
	// our default language
	defaultLanguage = language.AmericanEnglish
	// ignoreTTYCheck ignores the result of the TTY check (for testing!)
	ignoreTTYCheck = false
	// useColor is whether or not color output should be used, updated by Set*Writer.
	useColor = false
)

// fdWriter is the subset of file.File that implements io.Writer and Fd()
type fdWriter interface {
	io.Writer
	Fd() uintptr
}

// OutStyle writes a stylized and formatted message to stdout
func OutStyle(style, format string, a ...interface{}) error {
	OutStyle, err := applyStyle(style, useColor, fmt.Sprintf(format, a...))
	if err != nil {
		// Try anyways
		if err := Out(OutStyle); err != nil {
			glog.Errorf("Out failed: %v", err)
		}
		return err
	}
	return Out(OutStyle)
}

// Out writes a basic formatted string to stdout
func Out(format string, a ...interface{}) error {
	p := message.NewPrinter(preferredLanguage)
	if outFile == nil {
		return fmt.Errorf("No output file has been set")
	}
	_, err := p.Fprintf(outFile, format, a...)
	return err
}

// OutLn writes a basic formatted string with a newline to stdout
func OutLn(format string, a ...interface{}) error {
	return Out(format+"\n", a...)
}

// ErrStyle writes a stylized and formatted error message to stderr
func ErrStyle(style, format string, a ...interface{}) error {
	format, err := applyStyle(style, useColor, fmt.Sprintf(format, a...))
	if err != nil {
		// Try anyways.
		if err := Err(format); err != nil {
			glog.Errorf("Err failed: %v", err)
		}
		return err
	}
	return Err(format)
}

// Err writes a basic formatted string to stderr
func Err(format string, a ...interface{}) error {
	p := message.NewPrinter(preferredLanguage)
	if errFile == nil {
		return fmt.Errorf("No error output file has been set")
	}
	_, err := p.Fprintf(errFile, format, a...)
	return err
}

// ErrLn writes a basic formatted string with a newline to stderr
func ErrLn(format string, a ...interface{}) error {
	return Err(format+"\n", a...)
}

// Success is a shortcut for writing a styled success message to stdout
func Success(format string, a ...interface{}) error {
	return OutStyle("success", format, a...)
}

// Fatal is a shortcut for writing a styled fatal message to stderr
func Fatal(format string, a ...interface{}) error {
	return ErrStyle("fatal", format, a...)
}

// Warning is a shortcut for writing a styled warning message to stderr
func Warning(format string, a ...interface{}) error {
	return ErrStyle("warning", format, a...)
}

// Failure is a shortcut for writing a styled failure message to stderr
func Failure(format string, a ...interface{}) error {
	return ErrStyle("failure", format, a...)
}

// SetLanguageTag configures which language future messages should use.
func SetLanguageTag(l language.Tag) {
	glog.Infof("Setting Language to %s ...", l)
	preferredLanguage = l
}

// SetLanguage configures which language future messages should use, based on a LANG string.
func SetLanguage(s string) error {
	if s == "" || s == "C" {
		SetLanguageTag(defaultLanguage)
		return nil
	}
	// Ignore encoding preferences: we always output utf8. Handles "de_DE.utf8"
	parts := strings.Split(s, ".")
	l, err := language.Parse(parts[0])
	if err != nil {
		return err
	}
	SetLanguageTag(l)
	return nil
}

// SetOutFile configures which writer standard output goes to.
func SetOutFile(w fdWriter) {
	glog.Infof("Setting OutFile to %v (fd=%d) ...", w, w.Fd())
	outFile = w
	useColor = wantsColor(w.Fd())
}

// SetErrFile configures which writer error output goes to.
func SetErrFile(w fdWriter) {
	glog.Infof("Setting ErrFile to %v (fd=%d)...", w, w.Fd())
	errFile = w
	useColor = wantsColor(w.Fd())
}

// wantsColor determines if the user might want colorized output.
func wantsColor(fd uintptr) bool {
	// As in: term-256color
	if !strings.Contains(os.Getenv("TERM"), "color") {
		glog.Infof("TERM does not appear to support color")
		return false
	}
	// Allow boring people to continue to be boring people.
	if os.Getenv("MINIKUBE_IS_BORING") == "1" {
		glog.Infof("minikube is boring.")
		return false
	}
	if ignoreTTYCheck {
		return true
	}
	isT := isatty.IsTerminal(fd)
	glog.Infof("IsTerminal(%d) = %v", fd, isT)
	return isT
}
