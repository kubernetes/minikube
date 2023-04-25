// +build !windows

package commands

import (
	"k8s.io/minikube/pkg/libmachine/libmachine"
)

func cmdScp(c CommandLine, api libmachine.API) error {
	args := c.Args()
	if len(args) != 2 {
		c.ShowHelp()
		return errWrongNumberArguments
	}

	src := args[0]
	dest := args[1]

	hostInfoLoader := &storeHostInfoLoader{api}

	cmd, err := getScpCmd(src, dest, c.Bool("recursive"), c.Bool("delta"), c.Bool("quiet"), hostInfoLoader)
	if err != nil {
		return err
	}

	return runCmdWithStdIo(*cmd)
}
