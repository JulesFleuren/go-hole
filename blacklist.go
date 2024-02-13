package main

import (
	"sort"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/willf/bloom"
)

var (
	blacklistPath = "./data/blacklist.txt"

	blacklistHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "blacklist",
			Name:      "lookup_duration_seconds",
			Help:      "Duration of a domain lookup in the blacklist.",
			Buckets:   []float64{1e-6, 1.75e-6, 2.5e-6, 3.75e-6, 5e-6, 6.25e-6, 7.5e-6, 8.75e-6, 1e-5},
		},
		[]string{"bloom_filter", "array"},
	)
)

// Blacklist represents a set of domains to block.
// Blocked domains serve ads, tracking, malware, etc.
type Blacklist struct {
	filter *bloom.BloomFilter
	array  []string
}

// Size returns the number of domains in the blacklist.
func (blacklist *Blacklist) Size() int {
	return len(blacklist.array)
}

// Contains checks if the given domain belongs to the blacklist:
// the method returns true if the domain is present, false otherwise.
func (blacklist *Blacklist) Contains(domain string) bool {
	start := time.Now()

	// check the bloom filter first: it either says "definitely no present" or "maybe present"
	lower := strings.ToLower(domain)
	possiblyPresent := blacklist.filter.TestString(lower)
	if possiblyPresent {

		// the domain might be present... we need to manually check the list
		index := sort.SearchStrings(blacklist.array, lower)
		present := index < len(blacklist.array) && blacklist.array[index] == lower

		// collect metrics
		duration := time.Since(start).Seconds()
		if present {
			blacklistHistogram.WithLabelValues("maybe", "present").Observe(duration)
		} else {
			blacklistHistogram.WithLabelValues("maybe", "absent").Observe(duration)
		}

		return present
	}

	// collect metrics
	duration := time.Since(start).Seconds()
	blacklistHistogram.WithLabelValues("absent", "absent").Observe(duration)

	// if here, the domain is not present at all
	return false
}
