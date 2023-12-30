package dns

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rqlite/rqlite-disco-clients/expand"
)

const (
	DNSOverrideEnv = "RQLITE_DISCO_DNS_HOSTS"
)

// Client is a type can resolve a host for use by rqlite.
type Client struct {
	name string
	port int

	mu            sync.Mutex
	lastContact   time.Time
	lastAddresses []string
	lastError     error

	// Can be explicitly set for test purposes.
	lookupFn func(host string) ([]net.IP, error)
}

// NewConfigFromReader returns a Client configuration from the data read
// from r. If r is nil, a nil Configuration is returned.
func NewConfigFromReader(r io.Reader) (*Config, error) {
	if r == nil {
		return nil, nil
	}
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(expand.ExpandEnvBytes(b), &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// New returns an instantiated DNS client. If the cfg is nil, the default
// config is used.
func New(cfg *Config) *Client {
	client := &Client{
		name:     "rqlite",
		port:     4001,
		lookupFn: net.LookupIP,
	}

	if cfg != nil {
		if cfg.Name != "" {
			client.name = cfg.Name
		}
		if cfg.Port != 0 {
			client.port = cfg.Port
		}
	}
	return client
}

// NewWithPort returns an instantiated DNS client, with an explicit default port.
// If the cfg is nil, the default config is used but the port is overridden when
// using the default config.
func NewWithPort(cfg *Config, port int) *Client {
	client := &Client{
		name:     "rqlite",
		port:     port,
		lookupFn: net.LookupIP,
	}

	if cfg != nil {
		if cfg.Name != "" {
			client.name = cfg.Name
		}
		if cfg.Port != 0 {
			client.port = cfg.Port
		}
	}
	return client
}

// Lookup returns the network addresses resolved for the client's host value.
//
// If the environment variable RQLITE_DISCO_DNS_HOSTS is set, its value is used
// instead of that used by DNS resolution. That value is a comma-separated list
// of addresses, each of which is a host:port pair. This is useful for testing,
// and is not suitable for production use.
func (c *Client) Lookup() ([]string, error) {
	val, ok := os.LookupEnv(DNSOverrideEnv)
	if ok {
		addrs := make([]string, 0)
		for _, addr := range strings.Split(val, ",") {
			if _, _, err := net.SplitHostPort(addr); err != nil {
				return nil, fmt.Errorf("%s: invalid address %s", DNSOverrideEnv, addr)
			}
			addrs = append(addrs, addr)
		}
		return addrs, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	var ips []net.IP
	ips, c.lastError = c.lookupFn(c.name)
	if c.lastError != nil {
		return nil, c.lastError
	}
	c.lastContact = time.Now()

	addrs := make([]string, len(ips))
	for i := range ips {
		addrs[i] = net.JoinHostPort(ips[i].String(), strconv.Itoa(c.port))
	}

	c.lastAddresses = make([]string, len(addrs))
	copy(c.lastAddresses, addrs)

	return addrs, nil
}

// Stats returns some basic diagnostics information about the client.
func (c *Client) Stats() (map[string]interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	stats := map[string]interface{}{
		"mode": "dns",
		"name": c.name,
		"port": c.port,
	}

	if c.lastError != nil {
		stats["last_error"] = c.lastError.Error()
	}
	if !c.lastContact.IsZero() {
		stats["last_contact"] = c.lastContact
		stats["last_addresses"] = c.lastAddresses
	}

	return stats, nil
}
