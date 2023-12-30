package dns

import (
	"net"
	"os"
	"reflect"
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
	if exp, got := 1, len(addrs); exp != got {
		t.Fatalf("wrong number of addresses returned, exp %d, got %d", exp, got)
	}
	if !reflect.DeepEqual(addrs, []string{"8.8.8.8:4001"}) {
		t.Fatalf("failed to get correct address: %s", addrs)
	}
}

func Test_ClientLookupSingle_Env(t *testing.T) {
	os.Setenv(DNSOverrideEnv, "1.2.3.4:4001")
	defer os.Unsetenv(DNSOverrideEnv)

	client := New(nil)
	addrs, err := client.Lookup()
	if err != nil {
		t.Fatalf("failed to lookup host: %s", err.Error())
	}
	if exp, got := 1, len(addrs); exp != got {
		t.Fatalf("wrong number of addresses returned, exp %d, got %d", exp, got)
	}
	if !reflect.DeepEqual(addrs, []string{"1.2.3.4:4001"}) {
		t.Fatalf("failed to get correct address: %s", addrs)
	}
}

func Test_ClientLookupSingleIPv6(t *testing.T) {
	client := New(nil)
	lookupFn := func(host string) ([]net.IP, error) {
		if exp, got := "rqlite", host; exp != got {
			t.Fatalf("incorrect host resolved, exp %s, got %s", exp, got)
		}
		return []net.IP{net.ParseIP("2001:db8::68")}, nil
	}
	client.lookupFn = lookupFn

	addrs, err := client.Lookup()
	if err != nil {
		t.Fatalf("failed to lookup host: %s", err.Error())
	}
	if exp, got := 1, len(addrs); exp != got {
		t.Fatalf("wrong number of addresses returned, exp %d, got %d", exp, got)
	}
	if !reflect.DeepEqual(addrs, []string{"[2001:db8::68]:4001"}) {
		t.Fatalf("failed to get correct address: %s", addrs)
	}
}

func Test_ClientLookupSingleWithPort(t *testing.T) {
	client := NewWithPort(nil, 5001)
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
	if exp, got := 1, len(addrs); exp != got {
		t.Fatalf("wrong number of addresses returned, exp %d, got %d", exp, got)
	}
	if !reflect.DeepEqual(addrs, []string{"8.8.8.8:5001"}) {
		t.Fatalf("failed to get correct address: %s", addrs)
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
	if exp, got := 2, len(addrs); exp != got {
		t.Fatalf("wrong number of addresses returned, exp %d, got %d", exp, got)
	}
	if !reflect.DeepEqual(addrs, []string{"1.2.3.4:8080", "5.6.7.8:8080"}) {
		t.Fatalf("failed to get correct address: %s", addrs)
	}
}

func Test_ClientLookupDouble_Env(t *testing.T) {
	os.Setenv(DNSOverrideEnv, "1.2.3.4:8080,5.6.7.8:8080")
	defer os.Unsetenv(DNSOverrideEnv)

	client := New(nil)
	addrs, err := client.Lookup()
	if err != nil {
		t.Fatalf("failed to lookup host: %s", err.Error())
	}
	if exp, got := 2, len(addrs); exp != got {
		t.Fatalf("wrong number of addresses returned, exp %d, got %d", exp, got)
	}
	if !reflect.DeepEqual(addrs, []string{"1.2.3.4:8080", "5.6.7.8:8080"}) {
		t.Fatalf("failed to get correct address: %s", addrs)
	}
}

func Test_ClientLookupDouble_EnvError(t *testing.T) {
	os.Setenv(DNSOverrideEnv, "1.2.3.4:8080,5.6.7.8")
	defer os.Unsetenv(DNSOverrideEnv)

	client := New(nil)
	_, err := client.Lookup()
	if err == nil {
		t.Fatalf("expected error due to bad address")
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
