package dnssrv

import (
	"net"
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

	lookupSRVFn := func(service, proto, name string) (string, []*net.SRV, error) {
		if exp, got := "rqlite", service; exp != got {
			t.Fatalf("incorrect service resolved, exp %s, got %s", exp, got)
		}
		if exp, got := "rqlite", name; exp != got {
			t.Fatalf("incorrect name resolved, exp %s, got %s", exp, got)
		}
		return "", []*net.SRV{{
			Target:   "rqlite.node",
			Port:     1000,
			Priority: 1,
			Weight:   10,
		}}, nil
	}
	client.lookupSRVFn = lookupSRVFn

	lookupFn := func(host string) ([]net.IP, error) {
		if exp, got := "rqlite.node", host; exp != got {
			t.Fatalf("incorrect host resolved, exp %s, got %s", exp, got)
		}
		return []net.IP{net.IPv4(8, 8, 8, 8)}, nil
	}
	client.lookupFn = lookupFn

	addrs, err := client.Lookup()
	if err != nil {
		t.Fatalf("failed to lookup SRV record: %s", err.Error())
	}
	if exp, got := 1, len(addrs); exp != got {
		t.Fatalf("wrong number of addresses returned, exp %d, got %d", exp, got)
	}
	if !reflect.DeepEqual(addrs, []string{"8.8.8.8:1000"}) {
		t.Fatalf("failed to get correct address: %s", addrs)
	}
}

func Test_ClientLookupSingleIPv6(t *testing.T) {
	client := New(nil)

	lookupSRVFn := func(service, proto, name string) (string, []*net.SRV, error) {
		if exp, got := "rqlite", service; exp != got {
			t.Fatalf("incorrect service resolved, exp %s, got %s", exp, got)
		}
		if exp, got := "rqlite", name; exp != got {
			t.Fatalf("incorrect name resolved, exp %s, got %s", exp, got)
		}
		return "", []*net.SRV{{
			Target:   "rqlite.node",
			Port:     1000,
			Priority: 1,
			Weight:   10,
		}}, nil
	}
	client.lookupSRVFn = lookupSRVFn

	lookupFn := func(host string) ([]net.IP, error) {
		if exp, got := "rqlite.node", host; exp != got {
			t.Fatalf("incorrect host resolved, exp %s, got %s", exp, got)
		}
		return []net.IP{net.ParseIP("2001:db8::68")}, nil
	}
	client.lookupFn = lookupFn

	addrs, err := client.Lookup()
	if err != nil {
		t.Fatalf("failed to lookup SRV record: %s", err.Error())
	}
	if exp, got := 1, len(addrs); exp != got {
		t.Fatalf("wrong number of addresses returned, exp %d, got %d", exp, got)
	}
	if !reflect.DeepEqual(addrs, []string{"[2001:db8::68]:1000"}) {
		t.Fatalf("failed to get correct address: %s", addrs)
	}
}

func Test_ClientLookupDouble(t *testing.T) {
	client := New(nil)
	client.name = "rqlite-name"
	client.service = "rqlite-service"

	lookupSRVFn := func(service, proto, name string) (string, []*net.SRV, error) {
		if exp, got := "rqlite-service", service; exp != got {
			t.Fatalf("incorrect service resolved, exp %s, got %s", exp, got)
		}
		if exp, got := "rqlite-name", name; exp != got {
			t.Fatalf("incorrect name resolved, exp %s, got %s", exp, got)
		}
		return "", []*net.SRV{
			{
				Target:   "rqlite.node.1",
				Port:     1000,
				Priority: 1,
				Weight:   10,
			},
			{
				Target:   "rqlite.node.2",
				Port:     2000,
				Priority: 1,
				Weight:   10,
			},
		}, nil
	}
	client.lookupSRVFn = lookupSRVFn

	lookupFn := func(host string) ([]net.IP, error) {
		if host == "rqlite.node.1" {
			return []net.IP{net.IPv4(1, 1, 1, 1)}, nil
		} else if host == "rqlite.node.2" {
			return []net.IP{net.IPv4(2, 2, 2, 2)}, nil
		}
		t.Fatalf("incorrect host resolved, got %s", host)
		return nil, nil
	}
	client.lookupFn = lookupFn

	addrs, err := client.Lookup()
	if err != nil {
		t.Fatalf("failed to lookup SRV record: %s", err.Error())
	}
	if exp, got := 2, len(addrs); exp != got {
		t.Fatalf("wrong number of addresses returned, exp %d, got %d", exp, got)
	}
	if !reflect.DeepEqual(addrs, []string{"1.1.1.1:1000", "2.2.2.2:2000"}) {
		t.Fatalf("failed to get correct address: %s", addrs)
	}
}
