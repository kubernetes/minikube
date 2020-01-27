package metadata

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/minikube/pkg/kapi"
	"k8s.io/minikube/pkg/minikube/config"
)

func EnableOrDisable(name, val, profile string) error {
	enable, err := strconv.ParseBool(val)
	if err != nil {
		return errors.Wrapf(err, "parsing bool: %s", name)
	}
	if enable {
		return enableAddon()
	}
	return disableAddon()

}

func enableAddon() error {
	fmt.Println("updating configmap")
	if err := updateConfigmap(metadataCorefileConfigmap); err != nil {
		return err
	}
	fmt.Println("restarting core dns")
	if err := restartCoreDNS(); err != nil {
		return err
	}
	return nil
}

func disableAddon() error {
	fmt.Println("updating configmap")
	if err := updateConfigmap(originalCorefileConfigmap); err != nil {
		return err
	}
	fmt.Println("restarting core dns")
	if err := restartCoreDNS(); err != nil {
		return err
	}
	return nil
}

func restartCoreDNS() error {
	client, err := kapi.Client(viper.GetString(config.MachineProfile))
	if err != nil {
		return err
	}
	ns := "kube-system"
	pods, err := client.CoreV1().Pods(ns).List(metav1.ListOptions{})

	var coreDNSPods []string
	for _, p := range pods.Items {
		if !strings.Contains(p.GetName(), "coredns") {
			continue
		}
		coreDNSPods = append(coreDNSPods, p.GetName())
	}

	for _, p := range coreDNSPods {
		fmt.Println("Deleting", p)
		if err := client.CoreV1().Pods(ns).Delete(p, &metav1.DeleteOptions{}); err != nil {
			return err
		}
	}
	// Wait for deployment to be healthy again
	if err := kapi.WaitForDeploymentToStabilize(client, "kube-system", "coredns", 2*time.Minute); err != nil {
		return err
	}
	return nil
}
