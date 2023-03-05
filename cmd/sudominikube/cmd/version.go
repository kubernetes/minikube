package cmd

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/out"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/version"
)

var (
	versionOutput string
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of minikube",
	Long:  `Print the version of minikube.`,
	Run: func(command *cobra.Command, args []string) {
		minikubeVersion := version.GetVersion()
		gitCommitID := version.GetGitCommitID()
		data := map[string]interface{}{
			"minikubeVersion": minikubeVersion,
			"commit":          gitCommitID,
		}
		switch versionOutput {
		case "":
			out.Ln("minikube version: %v", minikubeVersion)
			if gitCommitID != "" {
				out.Ln("commit: %v", gitCommitID)
			}
		case "json":
			json, err := json.Marshal(data)
			if err != nil {
				exit.Error(reason.InternalJSONMarshal, "version json failure", err)
			}
			out.Ln(string(json))
		case "yaml":
			yaml, err := yaml.Marshal(data)
			if err != nil {
				exit.Error(reason.InternalYamlMarshal, "version yaml failure", err)
			}
			out.Ln(string(yaml))
		default:
			exit.Message(reason.InternalOutputUsage, "error: --output must be 'yaml' or 'json'")
		}
	},
}

func init() {
	versionCmd.Flags().StringVarP(&versionOutput, "output", "o", "", "One of 'yaml' or 'json'.")
}
