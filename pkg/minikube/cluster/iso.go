package cluster

import (
	"k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/driver"
)

// CacheISO downloads and caches ISO.
func CacheISO(cfg config.MachineConfig) error {
	if driver.BareMetal(cfg.VMDriver) {
		return nil
	}
	return cfg.Downloader.CacheMinikubeISOFromURL(cfg.MinikubeISO)
}
