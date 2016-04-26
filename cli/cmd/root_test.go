package cmd

import (
	"bufio"
	"io/ioutil"
	"log"
	"net"
	"os"
	"testing"

	"github.com/docker/machine/libmachine/drivers/plugin/localbinary"
	"github.com/kubernetes/minikube/cli/constants"
	"github.com/spf13/cobra"
)

func makeTempDir() string {
	tempDir, err := ioutil.TempDir("", "minipath")
	if err != nil {
		log.Fatal(err)
	}
	constants.Minipath = tempDir
	return tempDir
}

func runCommand(f func(*cobra.Command, []string)) {
	cmd := cobra.Command{}
	var args []string
	f(&cmd, args)
}

func TestPreRunDirectories(t *testing.T) {
	// Make sure we create the required directories.
	tempDir := makeTempDir()
	defer os.RemoveAll(tempDir)

	runCommand(RootCmd.PersistentPreRun)

	for _, dir := range dirs {
		_, err := os.Stat(dir)
		if os.IsNotExist(err) {
			t.Fatalf("Directory %s does not exist.", dir)
		}
	}
}

func TestPreRunNotDriver(t *testing.T) {
	tempDir := makeTempDir()
	defer os.RemoveAll(tempDir)
	runCommand(RootCmd.PersistentPreRun)
	if !localbinary.CurrentBinaryIsDockerMachine {
		t.Fatal("CurrentBinaryIsDockerMachine not set. This will break driver initialization.")
	}
}

func TestRunDriver(t *testing.T) {
	// This test is a bit complicated. It verifies that when the root command is
	// called with the proper environment variables, we setup the libmachine driver.

	tempDir := makeTempDir()
	defer os.RemoveAll(tempDir)
	os.Setenv(localbinary.PluginEnvKey, localbinary.PluginEnvVal)
	os.Setenv(localbinary.PluginEnvDriverName, "virtualbox")

	// Capture stdout and reset it later.
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = old
	}()

	// Run the command asynchronously. It should listen on a port for connections.
	go runCommand(RootCmd.Run)

	// The command will write out what port it's listening on over stdout.
	reader := bufio.NewReader(r)
	addr, _, err := reader.ReadLine()
	if err != nil {
		t.Fatal("Failed to read address over stdout.")
	}
	os.Stdout = old

	// Now that we got the port, make sure we can connect.
	if _, err := net.Dial("tcp", string(addr)); err != nil {
		t.Fatal("Driver not listening.")
	}
}
