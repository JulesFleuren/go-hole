package main

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
)

type Config struct {
	UpstreamDNS        string   `json:"UpstreamDNS"`
	UpstreamTlsSrvName string   `json:"UpstreamTlsSrvName"`
	BlocklistSources   []string `json:"BlocklistSources"`

	DNSPort        string `json:"DNSPort,omitempty"`
	PrometheusPort string `json:"PrometheusPort,omitempty"`
	Debug          bool   `json:"Debug,omitempty"`

	AdminUsernameHash []byte `json:"AdminUsernameHash,omitempty"`
	AdminPasswordHash []byte `json:"AdminPasswordHash,omitempty"`
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

	val, ok = os.LookupEnv("ADMIN_USR_HASH")
	if ok {
		hash, err := base64.StdEncoding.DecodeString(val)
		if err != nil {
			c.AdminUsernameHash = hash
		}
	}

	val, ok = os.LookupEnv("ADMIN_PWD_HASH")
	if ok {
		hash, err := base64.StdEncoding.DecodeString(val)
		if err != nil {
			c.AdminPasswordHash = hash
		}
	}

	return nil
}

func (c *Config) updateFromHttpRequest(r *http.Request) error {
	configFromRequest := Config{}
	err := json.NewDecoder(r.Body).Decode(&configFromRequest)
	if err != nil {
		return err
	}

	c.UpstreamDNS = configFromRequest.UpstreamDNS
	c.UpstreamTlsSrvName = configFromRequest.UpstreamTlsSrvName
	c.BlocklistSources = configFromRequest.BlocklistSources
	return nil
}

func (c *Config) JsonForHttpRequest() ([]byte, error) {
	configForRequest := Config{
		UpstreamDNS:        c.UpstreamDNS,
		UpstreamTlsSrvName: c.UpstreamTlsSrvName,
		BlocklistSources:   c.BlocklistSources,
	}
	return json.Marshal(configForRequest)
}

func (c *Config) writeToFile(path string) error {
	content, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return err
	}
	err = os.WriteFile(path, content, 0644)
	return err
}
