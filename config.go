package main

import (
	"encoding/json"
	"net/http"
	"os"
)

type Config struct {
	DNSPort            string
	PrometheusPort     string
	UpstreamDNS        string
	UpstreamTlsSrvName string
	Debug              bool

	BlocklistSources []string
}

func ConfigFromFile(path string) (Config, error) {
	file, err := os.ReadFile(path) // For read access.
	if err != nil {
		return Config{}, err
	}
	config := Config{}
	json.Unmarshal(file, &config)
	return config, nil
}

func (c *Config) updateFromEnvVar() error {
	val, ok := os.LookupEnv("DNS_PORT")
	if ok {
		c.DNSPort = val
	}

	val, ok = os.LookupEnv("PROMETHEUS_PORT")
	if ok {
		c.PrometheusPort = val
	}

	val, ok = os.LookupEnv("UPSTREAM_DNS")
	if ok {
		c.UpstreamDNS = val
	}

	val, ok = os.LookupEnv("UPSTREAM_TLS_SRVNAME")
	if ok {
		c.UpstreamTlsSrvName = val
	}

	val, ok = os.LookupEnv("DEBUG")
	if ok {
		if val == "true" {
			c.Debug = true
		}
	}
	return nil
}

func (c *Config) updateFromHttpRequest(r *http.Request) error {
	newConfig := Config{}
	err := json.NewDecoder(r.Body).Decode(&newConfig)
	if err != nil {
		return err
	}

	// Only these fields can be changed with an http request
	if newConfig.UpstreamDNS != "" {
		c.UpstreamDNS = newConfig.UpstreamDNS
	}
	if newConfig.UpstreamDNS != "" {
		c.UpstreamTlsSrvName = newConfig.UpstreamTlsSrvName
	}
	if newConfig.UpstreamDNS != "" {
		c.BlocklistSources = newConfig.BlocklistSources
	}
	return nil
}

func (c *Config) writeToFile(path string) error {
	content, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return err
	}
	err = os.WriteFile(path, content, 0644)
	return err
}
