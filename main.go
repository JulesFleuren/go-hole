package main

import (
	"log"

	"github.com/miekg/dns"
)

func main() {

	config, err := ConfigFromFile("config.json")
	config.updateFromEnvVar()
	if err != nil {
		log.Fatal("Error: Config file not found")
	}

	go runPrometheusServer(config)

	restartDNSServerChannel := make(chan struct{})

	go restartDNSServer(restartDNSServerChannel, &config)

	runWebPageServer(restartDNSServerChannel, &config)
}

func restartDNSServer(channel chan struct{}, config *Config) {
	for {
		//start (new) DNS server
		server := createDNSServer(*config)

		go func(srv *dns.Server) {
			err := server.ListenAndServe()
			if err != nil {
				log.Fatal(err)
			}
		}(server)

		// wait for signal to restart server
		<-channel

		// We received an interrupt signal, shut down.
		log.Println("Restarting DNS server")
		if err := server.Shutdown(); err != nil {
			// Error from closing listeners, or context timeout:
			log.Fatalf("Error shutting down DNS Server: %v", err)
		}
	}
}
