package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
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
		var responseBody map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&responseBody)

		if err != nil {
			log.Printf("Error decoding request: %v", err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		blocklistsString, ok := responseBody["blocklists"].(string)
		if !ok {
			log.Printf(`Error: Cannot convert "blocklists" field to string`)
			http.Error(w, `Error: Cannot convert "blocklists" field to string`, http.StatusBadRequest)
			return
		}
		// process urls
		// TODO: what if type assertion fails?
		sources := strings.Split(blocklistsString, "\n")

		// trim whitespace and remove empty urls
		strippedSources := make([]string, 0, len(sources))
		for _, source := range sources {
			if s := strings.TrimSpace(source); s != "" {
				strippedSources = append(strippedSources, s)
			}
		}

		config.BlocklistSources = strippedSources
		config.writeToFile("config.json")

		resp := make(map[string]string)
		resp["message"] = "Status OK"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
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

	http.ListenAndServe(":8080", nil)
}
