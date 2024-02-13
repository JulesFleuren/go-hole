package main

import (
	"bufio"
	"container/list"
	"errors"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/willf/bloom"
)

var (
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
	isNotUrlRe   = regexp.MustCompile(`[^a-zA-Z0-9\._\- ]`)
	bloomfilterK = uint(5)
	bloomfilterP = float64(0.01)
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

func LoadBlacklistFromSources(sources []string) *Blacklist {
	domains, size := combineDomainsFromSources(sources)

	// allocate the data structure of optimal Size
	n := float64(len(domains))
	m := uint(math.Ceil((n * math.Log(bloomfilterP)) / math.Log(1/math.Pow(2, math.Log(2)))))
	blacklist := Blacklist{
		filter: bloom.New(m, bloomfilterK),
		array:  domains,
	}

	for _, d := range domains {
		blacklist.filter.AddString(d)
	}

	fmt.Printf("Loaded %d domains. Size of array: %.2f MB, size of bloom filter: %.2f MB\n", len(domains), float64(size)/float64(1e6), float64(m)/float64(8e6))
	return &blacklist
}

// Loads domains from all sources and returns them in an alphabetically sorted list. Duplicates are not possible.
// Memorysize is an estimate of the amount of memory required in bytes
func combineDomainsFromSources(sources []string) (domainsArray []string, memorySize int) {
	domains := list.New()
	for _, source := range sources {
		newDomains, err := getDomains(source)

		if err != nil {
			fmt.Printf("Could not load from %s, continuing with next source. Error: %s\n", source, err.Error())
			continue
		}

		// make sure domains is not empty
		if domains.Len() == 0 {
			domains.PushFront(newDomains[0])
		}
		currentDomain := domains.Front()

	newDomainLoop:
		for _, newDomain := range newDomains {
			// Loop through linked list until the we find an element that should come after newDomain
			for newDomain > string(currentDomain.Value.(string)) {
				currentDomain = currentDomain.Next()

				// If we are at the end of the list, append newDomain to list and continue to next newDomain
				if currentDomain == nil {
					domains.PushBack(newDomain)
					currentDomain = domains.Back()
					continue newDomainLoop
				}
			}
			// If newDomain is already in list, continue, else insert in list
			if newDomain == currentDomain.Value.(string) {
				continue
			} else {
				currentDomain = domains.InsertBefore(newDomain, currentDomain)
			}
		}
	}

	domainsArray = make([]string, 0, domains.Len())
	memorySize = 0
	for e := domains.Front(); e != nil; e = e.Next() {
		domain := e.Value.(string)
		// memorySize += len(domain)
		domainsArray = append(domainsArray, domain)
	}
	// size of the pointers
	memorySize += len(domainsArray) * 16
	return domainsArray, memorySize
}

// Loads a source, filters out all domains and returns them in alphabetically sorted order. Duplicates are possible
func getDomains(source string) ([]string, error) {

	resp, err := http.Get(source)
	if err != nil {
		return []string{}, err
	}
	defer resp.Body.Close()

	// rough estimate for number of lines: 25 bytes per line
	urls := make([]string, 0, resp.ContentLength/25)

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if isNotUrlRe.MatchString(line) {
			continue
		}

		lastWord := line[strings.LastIndex(line, " ")+1:]

		if lastWord == "" {
			continue
		}

		url := strings.ToLower(lastWord)

		urls = append(urls, url)
	}

	if err := scanner.Err(); err != nil {
		return []string{}, err
	}

	slices.Sort(urls)

	if len(urls) == 0 {
		return []string{}, errors.New("no domains found in source")
	}

	return urls, nil
}
