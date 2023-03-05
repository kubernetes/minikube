package cmd

import (
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/detect"
	"k8s.io/minikube/pkg/minikube/download"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/reason"
	"k8s.io/minikube/pkg/minikube/sudominikube"
)

var sudominikubeCmd = &cobra.Command{
	Use:   "sudominikube",
	Short: "commands related with installation of sudominikube",
	Long:  "commands related with installation of sudominikube",
}

var (
	downloadPath string
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "download and install sudominikube",
	Long:  "download and install sudominikube at /opt/minikube/bin, sudo password will be required.",
	Run: func(cmd *cobra.Command, args []string) {
		location := downloadPath
		if downloadPath == "" {
			l, err := os.MkdirTemp("/tmp", "sudominikube")
			if err != nil {
				exit.Error(reason.Usage, "creating folder for download", err)
			}
			location = l
		}
		// downloading sudominikube
		err := download.SudoMinikube(location, runtime.GOOS, detect.EffectiveArch())
		if err != nil {
			exit.Error(reason.SudoMinkubeInstall, "error when downloading sudominikube", err)
		}
		// install minikube
		err = sudominikube.InstallSudoMinikube(location)
		if err != nil {
			exit.Error(reason.SudoMinkubeInstall, "error when installing sudominikube", err)
		}

	},
}

func init() {
	installCmd.Flags().StringVar(&downloadPath, "download-path", "", "location of download")
	sudominikubeCmd.AddCommand(installCmd)
}
