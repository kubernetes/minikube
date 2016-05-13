package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/docker/machine/drivers/virtualbox"
	"github.com/docker/machine/libmachine"
	"github.com/docker/machine/libmachine/log"
)

func usage() {
	fmt.Println("Usage: go run main.go <example>\n" +
		"Available examples: create streaming.")
	os.Exit(1)
}

// Sample Virtualbox create independent of Machine CLI.
func create() {
	log.SetDebug(true)

	client := libmachine.NewClient("/tmp/automatic", "/tmp/automatic/certs")
	defer client.Close()

	hostName := "myfunhost"

	// Set some options on the provider...
	driver := virtualbox.NewDriver(hostName, "/tmp/automatic")
	driver.CPU = 2
	driver.Memory = 2048

	data, err := json.Marshal(driver)
	if err != nil {
		log.Error(err)
		return
	}

	h, err := client.NewHost("virtualbox", data)
	if err != nil {
		log.Error(err)
		return
	}

	h.HostOptions.EngineOptions.StorageDriver = "overlay"

	if err := client.Create(h); err != nil {
		log.Error(err)
		return
	}

	out, err := h.RunSSHCommand("df -h")
	if err != nil {
		log.Error(err)
		return
	}

	fmt.Printf("Results of your disk space query:\n%s\n", out)

	fmt.Println("Powering down machine now...")
	if err := h.Stop(); err != nil {
		log.Error(err)
		return
	}
}

// Streaming the output of an SSH session in virtualbox.
func streaming() {
	log.SetDebug(true)

	client := libmachine.NewClient("/tmp/automatic", "/tmp/automatic/certs")
	defer client.Close()

	hostName := "myfunhost"

	// Set some options on the provider...
	driver := virtualbox.NewDriver(hostName, "/tmp/automatic")
	data, err := json.Marshal(driver)
	if err != nil {
		log.Error(err)
		return
	}

	h, err := client.NewHost("virtualbox", data)
	if err != nil {
		log.Error(err)
		return
	}

	if err := client.Create(h); err != nil {
		log.Error(err)
		return
	}

	h.HostOptions.EngineOptions.StorageDriver = "overlay"

	sshClient, err := h.CreateSSHClient()
	if err != nil {
		log.Error(err)
		return
	}

	stdout, stderr, err := sshClient.Start("yes | head -n 10000")
	if err != nil {
		log.Error(err)
		return
	}
	defer func() {
		_ = stdout.Close()
		_ = stderr.Close()
	}()

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Error(err)
	}
	if err := sshClient.Wait(); err != nil {
		log.Error(err)
	}

	fmt.Println("Powering down machine now...")
	if err := h.Stop(); err != nil {
		log.Error(err)
		return
	}
}

func main() {
	if len(os.Args) != 2 {
		usage()
	}

	switch os.Args[1] {
	case "create":
		create()
	case "streaming":
		streaming()
	default:
		usage()
	}
}
