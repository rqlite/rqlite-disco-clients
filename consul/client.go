package consul

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/consul/api"
	"github.com/rqlite/rqlite-disco-clients/expand"
)

// Client represents a Consul client.
type Client struct {
	client    *api.KV
	key       string
	leaderKey string
}

// NewConfigFromFile parses the file at path and returns a Config.
func NewConfigFromFile(path string) (*Config, error) {
	cfgFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer cfgFile.Close()

	b, err := ioutil.ReadAll(cfgFile)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(expand.ExpandEnvBytes(b), &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// NewConfigFromReader parses the data returned by the reader and
// returns a Config. A nil reader results in nil config.
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
	if strings.HasPrefix(cfg.Address, "http") {
		return nil, fmt.Errorf("address should not contain HTTP or HTTPS")
	}
	return &cfg, nil
}

// New returns an instantiated Consul client. If the cfg is nil, the default
// config is used.
func New(key string, cfg *Config) (*Client, error) {
	c, err := api.NewClient(consulConfigFromClientConfig(cfg))
	if err != nil {
		return nil, err
	}
	return &Client{
		client:    c.KV(),
		key:       key,
		leaderKey: fmt.Sprintf("%s/leader", key),
	}, nil
}

// GetLeader returns the leader as recorded in Consul. If a leader exists, ok will
// be set to true, false otherwise.
func (c *Client) GetLeader() (id string, apiAddr string, addr string, ok bool, e error) {
	pair, _, err := c.client.Get(c.leaderKey, nil)
	if err != nil {
		e = err
		return
	}
	if pair == nil {
		ok = false
		return
	}

	n := node{}
	if err := json.Unmarshal(pair.Value, &n); err != nil {
		e = err
		return
	}
	return n.ID, n.APIAddr, n.Addr, true, nil
}

// InitializeLeader sets the leader to the given details, but only if no leader
// has already been set. This operation is a check-and-set type operation. If
// initialization succeeds, ok is set to true.
func (c *Client) InitializeLeader(id, apiAddr, addr string) (bool, error) {
	b, err := json.Marshal(node{
		ID:      id,
		APIAddr: apiAddr,
		Addr:    addr,
	})
	if err != nil {
		return false, err
	}
	p := &api.KVPair{Key: c.leaderKey, Value: b}
	ok, _, err := c.client.CAS(p, nil)
	if err != nil {
		return false, err
	}
	return ok, nil
}

// SetLeader unconditionally sets the leader to the given details.
func (c *Client) SetLeader(id, apiAddr, addr string) error {
	b, err := json.Marshal(node{
		ID:      id,
		APIAddr: apiAddr,
		Addr:    addr,
	})
	if err != nil {
		return err
	}
	p := &api.KVPair{Key: c.leaderKey, Value: b}
	_, err = c.client.Put(p, nil)
	if err != nil {
		return err
	}
	return nil
}

// String implements the Stringer interface.
func (c *Client) String() string {
	return "consul-kv"
}

// Close closes the client.
func (c *Client) Close() error {
	return nil
}

type node struct {
	ID      string `json:"id,omitempty"`
	APIAddr string `json:"api_addr,omitempty"`
	Addr    string `json:"addr,omitempty"` // Needs TLS settings, etc I think so anyway. Maybe join handles?
}

func consulConfigFromClientConfig(cfg *Config) *api.Config {
	if cfg == nil {
		return api.DefaultConfig()
	}

	var basicAuth *api.HttpBasicAuth
	if cfg.BasicAuth != nil {
		basicAuth = &api.HttpBasicAuth{
			Username: cfg.BasicAuth.Username,
			Password: cfg.BasicAuth.Password,
		}
	}

	apiConfig := &api.Config{
		Address:    cfg.Address,
		Scheme:     cfg.Scheme,
		Datacenter: cfg.Datacenter,
		HttpAuth:   basicAuth,
		Token:      cfg.Token,
		TokenFile:  cfg.TokenFile,
		Namespace:  cfg.Namespace,
		Partition:  cfg.Partition,
	}

	if cfg.TLSConfig != nil {
		apiConfig.TLSConfig.Address = cfg.TLSConfig.Address
		apiConfig.TLSConfig.CAFile = cfg.TLSConfig.CAFile
		apiConfig.TLSConfig.CAPath = cfg.TLSConfig.CAPath
		apiConfig.TLSConfig.CAPem = cfg.TLSConfig.CAPem
		apiConfig.TLSConfig.CertFile = cfg.TLSConfig.CertFile
		apiConfig.TLSConfig.CertPEM = cfg.TLSConfig.CertPEM
		apiConfig.TLSConfig.KeyFile = cfg.TLSConfig.KeyFile
		apiConfig.TLSConfig.KeyPEM = cfg.TLSConfig.KeyPEM
		apiConfig.TLSConfig.InsecureSkipVerify = cfg.TLSConfig.InsecureSkipVerify
	}

	return apiConfig
}
