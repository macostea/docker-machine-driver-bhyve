package main

import (
	"fmt"
	"github.com/macostea/docker-machine-driver-bhyve/pkg/bhyve"
)

func main() {
	bHyveDriver := bhyve.NewDriver("bhyve-boot2docker", "/tmp/storePath")
	if err := bHyveDriver.Create(); err != nil {
		fmt.Printf("Error starting BHyve driver %v", err)
	}
}
