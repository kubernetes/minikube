// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"bytes"
	"fmt"
	"go/format"
	"io"
	"maps"
	"os"
	"path"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) Infof(format string, a ...any) {
	g.output(os.Stdout, format, a...)
}

func (g *Generator) Warnf(format string, a ...any) {
	g.Errorf(format, a...)
}

func (g *Generator) Errorf(format string, a ...any) {
	g.output(os.Stderr, format, a...)
}

func (g *Generator) Fatalf(format string, a ...any) {
	g.Errorf(format, a...)
	os.Exit(1)
}

func (g *Generator) output(w io.Writer, format string, a ...any) {
	fmt.Fprintf(w, format, a...)
	fmt.Fprint(w, "\n")
}

type Destination interface {
	CreateDirectories() error
	Write() error
	WriteBytes(body []byte) error
	WriteTemplate(templateName, templateBody string, templateData any, funcMaps ...template.FuncMap) error
	WriteTemplateSet(templates *template.Template, templateData any) error
}

func (g *Generator) NewGoFileDestination(filename string) Destination {
	return &fileDestination{
		baseDestination: baseDestination{formatter: format.Source},
		filename:        filename,
	}
}

func (g *Generator) NewUnformattedFileDestination(filename string) Destination {
	return &fileDestination{
		filename: filename,
	}
}

type fileDestination struct {
	baseDestination
	append   bool
	filename string
}

func (d *fileDestination) CreateDirectories() error {
	const (
		perm os.FileMode = 0755
	)
	dirname := path.Dir(d.filename)
	err := os.MkdirAll(dirname, perm)

	if err != nil {
		return fmt.Errorf("creating target directory %s: %w", dirname, err)
	}

	return nil
}

func (d *fileDestination) Write() error {
	var flags int
	if d.append {
		flags = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	} else {
		flags = os.O_TRUNC | os.O_CREATE | os.O_WRONLY
	}
	f, err := os.OpenFile(d.filename, flags, 0644) //nolint:mnd // good protection for new files

	if err != nil {
		return fmt.Errorf("opening file (%s): %w", d.filename, err)
	}

	defer f.Close()

	_, err = f.WriteString(d.buffer.String())

	if err != nil {
		return fmt.Errorf("writing to file (%s): %w", d.filename, err)
	}

	return nil
}

type baseDestination struct {
	formatter func([]byte) ([]byte, error)
	buffer    strings.Builder
}

func (d *baseDestination) WriteBytes(body []byte) error {
	_, err := d.buffer.Write(body)
	return err
}

func (d *baseDestination) WriteTemplate(templateName, templateBody string, templateData any, funcMaps ...template.FuncMap) error {
	body, err := parseTemplate(templateName, templateBody, templateData, funcMaps...)

	if err != nil {
		return err
	}

	body, err = d.format(body)
	if err != nil {
		return err
	}

	return d.WriteBytes(body)
}

func parseTemplate(templateName, templateBody string, templateData any, funcMaps ...template.FuncMap) ([]byte, error) {
	funcMap := template.FuncMap{
		"FirstUpper": FirstUpper,
		// Title returns a string with the first character of each word as upper case.
		"Title": cases.Title(language.Und, cases.NoLower).String,
	}
	for _, v := range funcMaps {
		maps.Copy(funcMap, v) // Extras overwrite defaults.
	}
	tmpl, err := template.New(templateName).Funcs(funcMap).Parse(templateBody)

	if err != nil {
		return nil, fmt.Errorf("parsing function template: %w", err)
	}

	return executeTemplate(tmpl, templateData)
}

func executeTemplate(tmpl *template.Template, templateData any) ([]byte, error) {
	var buffer bytes.Buffer
	err := tmpl.Execute(&buffer, templateData)

	if err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}

	return buffer.Bytes(), nil
}

func (d *baseDestination) WriteTemplateSet(templates *template.Template, templateData any) error {
	body, err := executeTemplate(templates, templateData)
	if err != nil {
		return err
	}

	body, err = d.format(body)
	if err != nil {
		return err
	}

	return d.WriteBytes(body)
}

func (d *baseDestination) format(body []byte) ([]byte, error) {
	if d.formatter == nil {
		return body, nil
	}

	unformattedBody := body
	body, err := d.formatter(unformattedBody)
	if err != nil {
		return nil, fmt.Errorf("formatting parsed template:\n%s\n%w", unformattedBody, err)
	}

	return body, nil
}

// FirstUpper returns a string with the first character as upper case.
func FirstUpper(s string) string {
	if s == "" {
		return ""
	}
	r, n := utf8.DecodeRuneInString(s)
	return string(unicode.ToUpper(r)) + s[n:]
}
