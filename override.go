package main

import (
	"log"
	"net"

	"github.com/miekg/dns"
)

type Override struct {
	Type   string `json:"Type"`
	Domain string `json:"Domain"`
	Ip     string `json:"Ip"`
}

func addOverridesToCache(config *Config, cache *Cache) {
	for _, override := range config.Overrides {
		// create a mock request for this domain
		req := new(dns.Msg)

		// create a response that will be added to the cache
		res := new(dns.Msg)

		if override.Ip == "" {
			if override.Type == "A" {
				req.SetQuestion(dns.Fqdn(override.Domain), dns.TypeA)
			} else if override.Type == "AAAA" {
				req.SetQuestion(dns.Fqdn(override.Domain), dns.TypeAAAA)
			} else {
				continue
			}
			// create a NXDOMAIN response
			res.SetRcode(req, dns.RcodeNameError)
			cache.Set(&req.Question[0], res, NoExpiration)
		} else {

			// create the answer and add it to the response
			switch override.Type {
			case "A":
				recordType := dns.TypeA

				record := new(dns.A)
				record.Hdr = dns.RR_Header{
					Name:   dns.Fqdn(override.Domain),
					Rrtype: recordType,
					Class:  dns.ClassINET,
					Ttl:    3600,
				}
				record.A = net.ParseIP(override.Ip)
				if record.A == nil {
					log.Printf("Not able to parse override IP: %s\n", override.Ip)
					continue
				}

				req.SetQuestion(dns.Fqdn(override.Domain), recordType)
				res.SetReply(req)
				res.Answer = []dns.RR{record}

				// add record to cache with no expiration
				cache.Set(&req.Question[0], res, NoExpiration)

			case "AAAA":
				recordType := dns.TypeAAAA

				record := new(dns.AAAA)
				record.Hdr = dns.RR_Header{
					Name:   dns.Fqdn(override.Domain),
					Rrtype: recordType,
					Class:  dns.ClassINET,
					Ttl:    3600,
				}
				record.AAAA = net.ParseIP(override.Ip)
				if record.AAAA == nil {
					log.Printf("Not able to parse override IP: %s\n", override.Ip)
					continue
				}

			default:
				log.Printf("Unsupported override type: %s\n", override.Type)
				continue
			}

		}
	}
	log.Printf("Succesfully added %v overrides\n", cache.c.ItemCount())
}
