package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func configController(w http.ResponseWriter, r *http.Request, config *Config, stopDNSServer chan struct{}) {
	switch r.Method {
	case "GET":

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		jsonResp, err := config.JsonForHttpRequest()
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

func basicAuth(config Config, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the username and password from the request
		// Authorization header. If no Authentication header is present
		// or the header value is invalid, then the 'ok' return value
		// will be false.

		// This function is a modified code snippet from
		// https://www.alexedwards.net/blog/basic-authentication-in-go
		// (c) Copyright 2013-2024 Alex Edwards (MIT license)

		username, password, ok := r.BasicAuth()
		if ok {
			usernameMatch := (bcrypt.CompareHashAndPassword(config.AdminUsernameHash, []byte(username)) == nil)
			passwordMatch := (bcrypt.CompareHashAndPassword(config.AdminPasswordHash, []byte(password)) == nil)

			if usernameMatch && passwordMatch {
				next.ServeHTTP(w, r)
				return
			}
		}

		// If the Authentication header is not present, is invalid, or the
		// username or password is wrong, then set a WWW-Authenticate
		// header to inform the client that we expect them to use basic
		// authentication and send a 401 Unauthorized response.
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}

func runWebPageServer(stopDNSServer chan struct{}, config *Config) {
	http.Handle("/", http.FileServer(http.Dir("./static")))

	http.HandleFunc("/config", basicAuth(*config, func(w http.ResponseWriter, r *http.Request) {
		configController(w, r, config, stopDNSServer)
	}))

	err := http.ListenAndServe(":8080", nil)
	fmt.Println(err)
}
