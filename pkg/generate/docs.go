/*
Copyright 2020 The Kubernetes Authors All rights reserved.

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

package generate

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/out"
)

// Docs generates docs for minikube command
func Docs(root *cobra.Command, path string) error {
	cmds := root.Commands()
	for _, c := range cmds {
		if c.Hidden {
			klog.Infof("Skipping generating doc for %s as it's a hidden command", c.Name())
			continue
		}
		contents, err := DocForCommand(c)
		if err != nil {
			return errors.Wrapf(err, "generating doc for %s", c.Name())
		}
		if err := saveDocForCommand(c, []byte(contents), path); err != nil {
			return errors.Wrapf(err, "saving doc for %s", c.Name())
		}
	}
	return nil
}

// DocForCommand returns the specific doc for that command
func DocForCommand(command *cobra.Command) (string, error) {
	buf := bytes.NewBuffer([]byte{})
	if err := generateTitle(command, buf); err != nil {
		return "", errors.Wrap(err, "generating title")
	}
	if err := rewriteFlags(command); err != nil {
		return "", errors.Wrap(err, "rewriting flags")
	}
	if err := writeSubcommands(command, buf); err != nil {
		return "", errors.Wrap(err, "writing subcommands")
	}
	return buf.String(), nil
}

// GenMarkdown creates markdown output.
func GenMarkdown(cmd *cobra.Command, w io.Writer) error {
	return GenMarkdownCustom(cmd, w, func(s string) string { return s })
}

// GenMarkdownCustom creates custom markdown output.
func GenMarkdownCustom(cmd *cobra.Command, w io.Writer, linkHandler func(string) string) error {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()

	buf := new(bytes.Buffer)
	name := cmd.CommandPath()

	short := cmd.Short
	long := cmd.Long
	if len(long) == 0 {
		long = short
	}

	buf.WriteString("## " + name + "\n\n")
	buf.WriteString(short + "\n\n")
	buf.WriteString("### Synopsis\n\n")
	buf.WriteString(long + "\n\n")

	if cmd.Runnable() {
		buf.WriteString(fmt.Sprintf("```shell\n%s\n```\n\n", cmd.UseLine()))
	}

	if len(cmd.Example) > 0 {
		buf.WriteString("### Examples\n\n")
		buf.WriteString(fmt.Sprintf("```\n%s\n```\n\n", cmd.Example))
	}

	if err := printOptions(buf, cmd); err != nil {
		return err
	}
	_, err := buf.WriteTo(w)
	return err
}

func printOptions(buf *bytes.Buffer, cmd *cobra.Command) error {
	flags := cmd.NonInheritedFlags()
	flags.SetOutput(buf)
	if flags.HasAvailableFlags() {
		buf.WriteString("### Options\n\n```\n")
		flags.PrintDefaults()
		buf.WriteString("```\n\n")
	}

	parentFlags := cmd.InheritedFlags()
	parentFlags.SetOutput(buf)
	if parentFlags.HasAvailableFlags() {
		buf.WriteString("### Options inherited from parent commands\n\n```\n")
		parentFlags.PrintDefaults()
		buf.WriteString("```\n\n")
	}
	return nil
}

// writeSubcommands recursively appends all subcommands to the doc
func writeSubcommands(command *cobra.Command, w io.Writer) error {
	if err := GenMarkdown(command, w); err != nil {
		return errors.Wrapf(err, "getting markdown custom")
	}
	if !command.HasSubCommands() {
		return nil
	}
	subCommands := command.Commands()
	for _, sc := range subCommands {
		if err := writeSubcommands(sc, w); err != nil {
			return err
		}
	}
	return nil
}

func generateTitle(command *cobra.Command, w io.Writer) error {
	date := time.Now().Format("2006-01-02")
	title := out.Fmt(title, out.V{"Command": command.Name(), "Description": command.Short, "Date": date})
	_, err := w.Write([]byte(title))
	return err
}

func saveDocForCommand(command *cobra.Command, contents []byte, path string) error {
	fp := filepath.Join(path, fmt.Sprintf("%s.md", command.Name()))
	if err := os.Remove(fp); err != nil {
		klog.Warningf("error removing %s", fp)
	}
	return ioutil.WriteFile(fp, contents, 0o644)
}
