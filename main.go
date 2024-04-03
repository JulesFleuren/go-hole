package main

import (
	"log"
)

func main() {

	config, err := ConfigFromFile("/etc/gohole/config.json")
	config.updateFromEnvVar()
	if err != nil {
		log.Fatal("Error: Config file not found")
	}

	go runPrometheusServer(config)

	restartDNSServerChannel := make(chan struct{})

	go startDNSServer(restartDNSServerChannel, &config)

	runWebPageServer(restartDNSServerChannel, &config)
}
