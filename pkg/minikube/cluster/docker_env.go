package cluster

import (
	"fmt"
	"net"

	"github.com/docker/machine/libmachine"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/drivers/kic"
	"k8s.io/minikube/pkg/drivers/kic/oci"
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/driver"
	"k8s.io/minikube/pkg/minikube/localpath"
)

// GetHostDockerEnv gets the necessary docker env variables to allow the use of docker through minikube's vm
func GetHostDockerEnv(api libmachine.API) (map[string]string, error) {
	pName := viper.GetString(config.MachineProfile)
	host, err := CheckIfHostExistsAndLoad(api, pName)
	if err != nil {
		return nil, errors.Wrap(err, "Error checking that api exists and loading it")
	}

	ip := kic.DefaultBindIPV4
	if !driver.IsKIC(host.Driver.DriverName()) { // kic externally accessible ip is different that node ip
		ip, err = host.Driver.GetIP()
		if err != nil {
			return nil, errors.Wrap(err, "Error getting ip from host")
		}

	}

	tcpPrefix := "tcp://"
	port := constants.DockerDaemonPort
	if driver.IsKIC(host.Driver.DriverName()) { // for kic we need to find out what port docker allocated during creation
		port, err = oci.HostPortBinding(host.Driver.DriverName(), pName, constants.DockerDaemonPort)
		if err != nil {
			return nil, errors.Wrapf(err, "get hostbind port for %d", constants.DockerDaemonPort)
		}
	}

	envMap := map[string]string{
		"DOCKER_TLS_VERIFY": "1",
		"DOCKER_HOST":       tcpPrefix + net.JoinHostPort(ip, fmt.Sprint(port)),
		"DOCKER_CERT_PATH":  localpath.MakeMiniPath("certs"),
	}
	return envMap, nil
}
