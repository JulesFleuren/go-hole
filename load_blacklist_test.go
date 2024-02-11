package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var (
	testResponse1 = `# This is a comment

0.0.0.0 a.com
0.0.0.0 e.com
0.0.0.0 b.com
`
	testResponse2 = `f.com
b.com
c.com
c.com
`
)

func TestGetDomains(t *testing.T) {
	ts1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, testResponse1)
	}))
	defer ts1.Close()

	domains, err := getDomains(ts1.URL)

	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(domains, []string{"a.com", "b.com", "e.com"}) {
		t.Error("Domains are wrong, got: ", domains, " expected: ", []string{"a.com", "b.com", "e.com"})
	}

	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, testResponse2)
	}))
	defer ts2.Close()

	domains, err = getDomains(ts2.URL)

	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(domains, []string{"b.com", "c.com", "c.com", "f.com"}) {
		t.Error("Domains are wrong, got: ", domains, " expected: ", []string{"b.com", "c.com", "c.com", "f.com"})
	}
}

func TestCombineDomainsFromSources(t *testing.T) {
	ts1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, testResponse1)
	}))
	defer ts1.Close()

	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, testResponse2)
	}))
	defer ts2.Close()

	domains, _ := combineDomainsFromSources([]string{ts1.URL, ts2.URL})

	if !reflect.DeepEqual(domains, []string{"a.com", "b.com", "c.com", "e.com", "f.com"}) {
		t.Error("Domains are wrong, got: ", domains, " expected: ", []string{"a.com", "b.com", "c.com", "e.com", "f.com"})
	}

	if len(domains) != cap(domains) {
		t.Error("Capacity of domains is bigger than length")
	}
}

func TestLoadBlackListFromSources(t *testing.T) {
	ts1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, testResponse1)
	}))
	defer ts1.Close()

	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, testResponse2)
	}))
	defer ts2.Close()

	blacklist := LoadBlacklistFromSources([]string{ts1.URL, ts2.URL})

	if blacklist.Contains(`d.com`) {
		t.Errorf("Domain d.com should not be blocked")
	}

	if !blacklist.Contains(`a.com`) {
		t.Errorf("Domain a.com should be blocked")
	}
}

func BenchmarkLoadBlacklistFromSources(b *testing.B) {
	for i := 0; i < b.N; i++ {
		LoadBlacklistFromSources([]string{`https://raw.githubusercontent.com/PolishFiltersTeam/KADhosts/master/KADhosts.txt`, `https://adaway.org/hosts.txt`})
	}
}

func BenchmarkGetDomains(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getDomains(`https://raw.githubusercontent.com/PolishFiltersTeam/KADhosts/master/KADhosts.txt`)
	}
}
