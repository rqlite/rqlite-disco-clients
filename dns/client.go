package dns

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/url"
	"sync"
	"time"
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
	if err := json.Unmarshal(b, &cfg); err != nil {
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

// Lookup returns the network addresses resolved for the client's host value.
func (c *Client) Lookup() ([]string, error) {
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
		addrs[i] = fmt.Sprintf("%s:%d", ips[i].String(), c.port)
	}

	c.lastAddresses = make([]string, len(addrs))
	copy(c.lastAddresses, addrs)

	return addrs, nil
}

// LookupURLs returns a slice of URLs which can be used to attempt to join the
// cluster. Only one of the accesses is guaranteed to work, but all are
// returned so that the client can try them all.
func (c *Client) LookupURLs() ([]*url.URL, error) {
	addrs, err := c.Lookup()
	if err != nil {
		return nil, err
	}

	protocols := []string{"http", "https", "raft"}
	urls := make([]*url.URL, len(addrs)*len(protocols))
	for i, addr := range addrs {
		for j, protocol := range protocols {
			urls[i*len(protocols)+j] = &url.URL{
				Scheme: protocol,
				Host:   addr,
			}
		}
	}

	return urls, nil
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
