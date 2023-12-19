package dnssrv

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/rqlite/rqlite-disco-clients/expand"
)

// Client is a type can retrieve SRV records for rqlite
type Client struct {
	name    string
	service string

	mu            sync.Mutex
	lastContact   time.Time
	lastAddresses []string
	lastError     error

	// Can be explicitly set for test purposes.
	lookupSRVFn func(service, proto, name string) (string, []*net.SRV, error)
	lookupFn    func(host string) ([]net.IP, error)
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

// New returns an instantiated DNS SRV client. If the cfg is nil, the default
// config is used.
func New(cfg *Config) *Client {
	client := &Client{
		name:        "rqlite",
		service:     "rqlite",
		lookupSRVFn: net.LookupSRV,
		lookupFn:    net.LookupIP,
	}

	if cfg != nil {
		if cfg.Name != "" {
			client.name = cfg.Name
		}
		if cfg.Service != "" {
			client.service = cfg.Service
		}
	}
	return client
}

// Lookup returns the network addresses from the DNS SRV records
func (c *Client) Lookup() ([]string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var records []*net.SRV
	_, records, c.lastError = c.lookupSRVFn(c.service, "tcp", c.name)
	if c.lastError != nil {
		return nil, c.lastError
	}
	c.lastContact = time.Now()

	addrs := make([]string, 0)
	for i := range records {
		// Now look up the IP address for the target. If there are more than
		// one, add them all.
		var ips []net.IP
		ips, c.lastError = c.lookupFn(records[i].Target)
		if c.lastError != nil {
			return nil, c.lastError
		}

		for j := range ips {
			addrs = append(addrs, net.JoinHostPort(ips[j].String(), fmt.Sprintf("%d", records[i].Port)))
		}
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
		"mode":      "dns-srv",
		"name":      c.name,
		"service":   c.service,
		"dns_name:": fmt.Sprintf("_%s._tcp.%s.", c.service, c.name),
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
