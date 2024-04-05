package main

import (
	"log"
	"net"

	"github.com/miekg/dns"
)

type Override struct {
	Domain string `json:"Domain"`
	Ip     string `json:"Ip"`
}

func addOverridesToCache(config *Config, cache *Cache) {
	for _, override := range config.Overrides {
		// create a mock request for this domain
		req := new(dns.Msg)
		req.SetQuestion(dns.Fqdn(override.Domain), dns.TypeA)

		res := new(dns.Msg)
		if override.Ip == "" {
			// create a NXDOMAIN response
			res.SetRcode(req, dns.RcodeNameError)
		} else {
			// create a response with the right ip address
			record := new(dns.A)
			record.Hdr = dns.RR_Header{
				Name:   dns.Fqdn(override.Domain),
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    3600,
			}
			record.A = net.ParseIP(override.Ip)
			if record.A == nil {
				log.Printf("Not able to parse override IP: %s\n", override.Ip)
				continue
			}

			res.SetReply(req)
			res.Answer = []dns.RR{record}
		}
		cache.Set(&req.Question[0], res, NoExpiration)
	}
	log.Printf("Succesfully added %v overrides\n", cache.c.ItemCount())
}
