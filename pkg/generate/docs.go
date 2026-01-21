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
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	"k8s.io/minikube/pkg/minikube/out"
)

// Docs generates docs for minikube command
func Docs(root *cobra.Command, path string, testPath string, codePath string) error {
	cmds := root.Commands()
	for _, c := range cmds {
		if c.Hidden {
			klog.Infof("Skipping generating doc for %s as it's a hidden command", c.Name())
			continue
		}
		contents, err := DocForCommand(c)
		if err != nil {
			return fmt.Errorf("generating doc for %s: %w", c.Name(), err)
		}
		if err := saveDocForCommand(c, []byte(contents), path); err != nil {
			return fmt.Errorf("saving doc for %s: %w", c.Name(), err)
		}
	}
	err := TestDocs(testPath, "test/integration")
	if err != nil {
		return fmt.Errorf("failed to generate test docs: %w", err)
	}

	return ErrorCodes(codePath, []string{"pkg/minikube/reason/reason.go", "pkg/minikube/reason/exitcodes.go"})
}

// DocForCommand returns the specific doc for that command
func DocForCommand(command *cobra.Command) (string, error) {
	buf := bytes.NewBuffer([]byte{})
	if err := generateTitle(command, buf); err != nil {
		return "", fmt.Errorf("generating title: %w", err)
	}
	if err := rewriteLogFile(); err != nil {
		return "", fmt.Errorf("rewriting log_file: %w", err)
	}
	if err := rewriteFlags(command); err != nil {
		return "", fmt.Errorf("rewriting flags: %w", err)
	}
	if err := writeSubcommands(command, buf); err != nil {
		return "", fmt.Errorf("writing subcommands: %w", err)
	}
	return buf.String(), nil
}

// GenMarkdown creates markdown output.
func GenMarkdown(cmd *cobra.Command, w io.Writer) error {
	return GenMarkdownCustom(cmd, w)
}

// GenMarkdownCustom creates custom markdown output.
func GenMarkdownCustom(cmd *cobra.Command, w io.Writer) error {
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
		fmt.Fprintf(buf, "```shell\n%s\n```\n\n", cmd.UseLine())
	}

	if len(cmd.Aliases) > 0 {
		buf.WriteString("### Aliases\n\n")
		fmt.Fprintf(buf, "%s\n\n", cmd.Aliases)
	}

	if len(cmd.Example) > 0 {
		buf.WriteString("### Examples\n\n")
		fmt.Fprintf(buf, "```\n%s\n```\n\n", cmd.Example)
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
		return fmt.Errorf("getting markdown custom: %w", err)
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
	return os.WriteFile(fp, contents, 0o644)
}
