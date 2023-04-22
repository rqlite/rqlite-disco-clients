package dns

import (
	"net"
	"net/url"
	"reflect"
	"sort"
	"testing"
)

func Test_NewClient(t *testing.T) {
	client := New(nil)
	if client == nil {
		t.Fatalf("failed to create default client")
	}
}

func Test_ClientStats(t *testing.T) {
	client := New(nil)
	stats, err := client.Stats()
	if err != nil {
		t.Fatalf("failed to get stats: %s", err.Error())
	}
	if stats == nil {
		t.Fatalf("nil stats returned")
	}
}

func Test_ClientLookupSingle(t *testing.T) {
	client := New(nil)

	lookupFn := func(host string) ([]net.IP, error) {
		if exp, got := "rqlite", host; exp != got {
			t.Fatalf("incorrect host resolved, exp %s, got %s", exp, got)
		}
		return []net.IP{net.IPv4(8, 8, 8, 8)}, nil
	}
	client.lookupFn = lookupFn

	addrs, err := client.Lookup()
	if err != nil {
		t.Fatalf("failed to lookup host: %s", err.Error())
	}
	if !reflect.DeepEqual(addrs, []string{"8.8.8.8:4001"}) {
		t.Fatalf("failed to get correct address: %s", addrs)
	}
}

func Test_ClientLookupURLsSingle(t *testing.T) {
	client := New(nil)

	lookupFn := func(host string) ([]net.IP, error) {
		if exp, got := "rqlite", host; exp != got {
			t.Fatalf("incorrect host resolved, exp %s, got %s", exp, got)
		}
		return []net.IP{net.IPv4(8, 8, 8, 8)}, nil
	}
	client.lookupFn = lookupFn

	urls, err := client.LookupURLs()
	if err != nil {
		t.Fatalf("failed to lookup host: %s", err.Error())
	}

	expURLs := []*url.URL{
		{Scheme: "http", Host: "8.8.8.8:4001"},
		{Scheme: "https", Host: "8.8.8.8:4001"},
		{Scheme: "raft", Host: "8.8.8.8:4001"},
	}
	sortURLs(urls)
	sortURLs(expURLs)

	if !reflect.DeepEqual(urls, expURLs) {
		t.Fatalf("failed to get correct address: got %s, exp %s", urls, expURLs)
	}
}

func Test_ClientLookupDouble(t *testing.T) {
	client := New(nil)
	client.name = "qux"
	client.port = 8080

	lookupFn := func(host string) ([]net.IP, error) {
		if exp, got := client.name, host; exp != got {
			t.Fatalf("incorrect host resolved, exp %s, got %s", exp, got)
		}
		return []net.IP{net.IPv4(1, 2, 3, 4), net.IPv4(5, 6, 7, 8)}, nil
	}
	client.lookupFn = lookupFn

	addrs, err := client.Lookup()
	if err != nil {
		t.Fatalf("failed to lookup host: %s", err.Error())
	}
	if !reflect.DeepEqual(addrs, []string{"1.2.3.4:8080", "5.6.7.8:8080"}) {
		t.Fatalf("failed to get correct address: %s", addrs)
	}
}

func Test_ClientLookupURLsDouble(t *testing.T) {
	client := New(nil)
	client.name = "qux"
	client.port = 8080

	lookupFn := func(host string) ([]net.IP, error) {
		if exp, got := client.name, host; exp != got {
			t.Fatalf("incorrect host resolved, exp %s, got %s", exp, got)
		}
		return []net.IP{net.IPv4(1, 2, 3, 4), net.IPv4(5, 6, 7, 8)}, nil
	}

	client.lookupFn = lookupFn

	urls, err := client.LookupURLs()
	if err != nil {
		t.Fatalf("failed to lookup host: %s", err.Error())
	}

	expURLs := []*url.URL{
		{Scheme: "http", Host: "1.2.3.4:8080"},
		{Scheme: "https", Host: "1.2.3.4:8080"},
		{Scheme: "raft", Host: "1.2.3.4:8080"},
		{Scheme: "http", Host: "5.6.7.8:8080"},
		{Scheme: "https", Host: "5.6.7.8:8080"},
		{Scheme: "raft", Host: "5.6.7.8:8080"},
	}
	sortURLs(urls)
	sortURLs(expURLs)

	if !reflect.DeepEqual(urls, expURLs) {
		t.Fatalf("failed to get correct address: got %s, exp %s", urls, expURLs)
	}
}

func Test_ClientLookupLocalhost(t *testing.T) {
	client := New(nil)
	client.name = "localhost"
	client.port = 8080

	addrs, err := client.Lookup()
	if err != nil {
		t.Fatalf("failed to lookup host: %s", err.Error())
	}

	// At least one address should be IPv4, testing that an actual lookup
	// took place.
	for i := range addrs {
		if addrs[i] == "127.0.0.1:8080" {
			return
		}
	}
	t.Fatalf("failed to get local address %s", addrs)
}

// sort a slice of URLs by their string representation
func sortURLs(urls []*url.URL) {
	sort.Slice(urls, func(i, j int) bool {
		return urls[i].String() < urls[j].String()
	})
}
