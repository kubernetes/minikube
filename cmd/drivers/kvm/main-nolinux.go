// +build !linux

package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println(
		"this driver was built on a non-linux machine, so it is " +
			"unavailable. Please re-build minikube on a linux machine to enable " +
			"it.",
	)
	os.Exit(1)
}
