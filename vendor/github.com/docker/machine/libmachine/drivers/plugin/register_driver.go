package plugin

import (
	"fmt"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"time"

	"github.com/docker/machine/libmachine/drivers"
	"github.com/docker/machine/libmachine/drivers/plugin/localbinary"
	"github.com/docker/machine/libmachine/drivers/rpc"
	"github.com/docker/machine/libmachine/log"
	"github.com/docker/machine/libmachine/version"
)

var (
	heartbeatTimeout = 10 * time.Second
)

func RegisterDriver(d drivers.Driver) {
	if os.Getenv(localbinary.PluginEnvKey) != localbinary.PluginEnvVal {
		fmt.Fprintf(os.Stderr, `This is a Docker Machine plugin binary.
Plugin binaries are not intended to be invoked directly.
Please use this plugin through the main 'docker-machine' binary.
(API version: %d)
`, version.APIVersion)
		os.Exit(1)
	}

	log.SetDebug(true)
	os.Setenv("MACHINE_DEBUG", "1")

	rpcd := rpcdriver.NewRPCServerDriver(d)
	rpc.RegisterName(rpcdriver.RPCServiceNameV0, rpcd)
	rpc.RegisterName(rpcdriver.RPCServiceNameV1, rpcd)
	rpc.HandleHTTP()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading RPC server: %s\n", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println(listener.Addr())

	go http.Serve(listener, nil)

	for {
		select {
		case <-rpcd.CloseCh:
			log.Debug("Closing plugin on server side")
			os.Exit(0)
		case <-rpcd.HeartbeatCh:
			continue
		case <-time.After(heartbeatTimeout):
			// TODO: Add heartbeat retry logic
			os.Exit(1)
		}
	}
}
