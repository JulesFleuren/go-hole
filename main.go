package main

import (
	"log"
)

var (
	config Config
)

func main() {

	config, err := ConfigFromFile("/etc/gohole/config.json")
	config.updateFromEnvVar()
	if err != nil {
		log.Fatal("Error: Config file not found")
	}

	go runPrometheusServer()

	restartDNSServerChannel := make(chan struct{})

	go startDNSServer(restartDNSServerChannel)

	runWebPageServer(restartDNSServerChannel)
}
