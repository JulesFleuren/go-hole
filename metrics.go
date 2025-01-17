package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// namespace for all metrics of the application
	namespace = "gohole"
)

// runPrometheusServer starts an HTTP server which exposes
// the application metrics in the Prometheus format.
func runPrometheusServer(config Config) {
	port := config.PrometheusPort
	if port == "0" {
		fmt.Printf("HTTP server with metrics has been DISABLED.\n")
		return
	}

	fmt.Printf("Starting HTTP server with metrics on TCP port %s...\n", port)
	server := &http.Server{Addr: "0.0.0.0:" + port}
	http.Handle("/metrics", promhttp.Handler())
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
