package dns

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
)

// Client is a type can resolve a host for use by rqlite.
type Client struct {
	name string
	port int

	// Can be explicitly set for test purposes.
	lookupFn func(host string) ([]net.IP, error)
}

// NewConfigFromReader returns a Client configuration from the data read
// from r.
func NewConfigFromReader(r io.Reader) (*Config, error) {
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
	ips, err := c.lookupFn(c.name)
	if err != nil {
		return nil, err
	}
	addrs := make([]string, len(ips))
	for i := range ips {
		addrs[i] = fmt.Sprintf("%s:%d", ips[i].String(), c.port)
	}
	return addrs, nil
}
