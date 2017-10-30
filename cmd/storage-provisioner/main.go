package main

import (
	"k8s.io/minikube/cmd/localkube/cmd"
	"sync"
)

func main() {
	localkubeServer := cmd.NewLocalkubeServer()
	storageProvisionerServer := localkubeServer.NewStorageProvisionerServer()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		storageProvisionerServer.Start()
	}()
	wg.Wait()
}
