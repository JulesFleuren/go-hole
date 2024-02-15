package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func configController(w http.ResponseWriter, r *http.Request, config *Config, stopDNSServer chan struct{}) {
	switch r.Method {
	case "GET":

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		jsonResp, err := json.Marshal(config)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		}
		w.Write(jsonResp)

	case "POST":
		err := config.updateFromHttpRequest(r)
		if err != nil {
			log.Printf("Error decoding request: %v", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		config.writeToFile("config.json")

		resp := make(map[string]string)
		resp["message"] = "Status OK"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResp)

		// Send signal to restart dns server
		stopDNSServer <- struct{}{}
		return
	default:
		http.Error(w, "Only GET and POST methods are supported.", http.StatusMethodNotAllowed)
	}
}

func runWebPageServer(stopDNSServer chan struct{}, config *Config) {
	http.Handle("/", http.FileServer(http.Dir("./static")))

	http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		configController(w, r, config, stopDNSServer)
	})

	err := http.ListenAndServe(":8080", nil)
	fmt.Println(err)
}
