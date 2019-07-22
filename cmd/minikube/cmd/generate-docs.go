package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/out"
	"os"
)

var path string

// generateDocs represents the generate-docs command
var generateDocs = &cobra.Command{
	Use:   "generate-docs",
	Short: "Populates the specified folder with documentation in markdown about minikube",
	Long:  "Populates the specified folder with documentation in markdown about minikube",
	Example: "minikube generate-docs --path <FOLDER_PATH>",
	Run: func(cmd *cobra.Command, args []string) {

		// if directory does not exist
		docsPath, err := os.Stat(path)
		if err != nil || !docsPath.IsDir() {
			exit.UsageT("Unable to generate the documentation. Please ensure that the path specified is a directory, exists & you have permission to write to it.")
		}

		// generate docs
		if err := doc.GenMarkdownTree(RootCmd,path); err != nil {
			exit.WithError("Unable to generate docs", err)
		}
		out.T(out.Documentation,"Docs have been saved at - {{.path}}",out.V{"path":path})
	},

}

func init() {
	generateDocs.Flags().StringVar(&path,"path","","The path on the file system where the docs in markdown need to be saved")
	RootCmd.AddCommand(generateDocs)
}