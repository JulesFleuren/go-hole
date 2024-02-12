package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func indexHandler(w http.ResponseWriter, r *http.Request, config *Config, stopDNSServer chan struct{}) {
	switch r.Method {
	case "GET":
		http.ServeFile(w, r, "./static/index.html")
	case "POST":
		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}

		// process urls
		sources := strings.Split(r.FormValue("blocklists"), "\n")

		// trim whitespace and remove empty urls
		strippedSources := make([]string, 0, len(sources))
		for _, source := range sources {
			if s := strings.TrimSpace(source); s != "" {
				strippedSources = append(strippedSources, s)
			}
		}

		config.BlocklistSources = strippedSources
		config.writeToFile("config.json")

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		resp := make(map[string]string)
		resp["message"] = "Status OK"
		jsonResp, err := json.Marshal(resp)
		if err != nil {
			log.Fatalf("Error happened in JSON marshal. Err: %s", err)
		}
		w.Write(jsonResp)

		// Send signal to restart dns server
		stopDNSServer <- struct{}{}
		return
	default:
		fmt.Fprintf(w, "Only GET and POST methods are supported.")
	}
}

func configHandler(w http.ResponseWriter, r *http.Request, config Config) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	jsonResp, err := json.Marshal(config)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)
}

func runWebPageServer(stopDNSServer chan struct{}, config *Config) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		indexHandler(w, r, config, stopDNSServer)
	})

	http.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		configHandler(w, r, *config)
	})

	http.ListenAndServe("localhost:8080", nil)
}
