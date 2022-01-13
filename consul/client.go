package consul // Maybe put in own repo -- rqlite-disco-clients

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hashicorp/consul/api"
)

// Client represents a Consul client.
type Client struct {
	client    *api.KV
	key       string
	leaderKey string
}

// BasicAuthConfig stores HTTP Basic Auth credentials.
type BasicAuthConfig struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// TLSConfig sets the configuration for TLS communication with Consul.
type TLSConfig struct {
	// Address is the optional address of the Consul server. The port, if any
	// will be removed from here and this will be set to the ServerName of the
	// resulting config.
	Address string `json:"address,omitempty"`

	// CAFile is the optional path to the CA certificate used for Consul
	// communication, defaults to the system bundle if not specified.
	CAFile string `json:"ca_file,omitempty"`

	// CAPath is the optional path to a directory of CA certificates to use for
	// Consul communication, defaults to the system bundle if not specified.
	CAPath string `json:"ca_path,omitempty"`

	// CAPem is the optional PEM-encoded CA certificate used for Consul
	// communication, defaults to the system bundle if not specified.
	CAPem []byte `json:"ca_pem,omitempty"`

	// CertFile is the optional path to the certificate for Consul
	// communication. If this is set then you need to also set KeyFile.
	CertFile string `json:"cert_file,omitempty"`

	// CertPEM is the optional PEM-encoded certificate for Consul
	// communication. If this is set then you need to also set KeyPEM.
	CertPEM []byte `json:"cert_pem,omitempty"`

	// KeyFile is the optional path to the private key for Consul communication.
	// If this is set then you need to also set CertFile.
	KeyFile string `json:"key_file,omitempty"`

	// KeyPEM is the optional PEM-encoded private key for Consul communication.
	// If this is set then you need to also set CertPEM.
	KeyPEM []byte `json:"key_pem,omitempty"`

	// InsecureSkipVerify if set to true will disable TLS host verification.
	InsecureSkipVerify bool `json:"insecure_skip_verify,omitempty"`
}

// Config for Consul client.
type Config struct {
	// Address is the address of the Consul server
	Address string `json:"address,omitempty"`

	// Scheme is the URI scheme for the Consul server
	Scheme string `json:"schema,omitempty"`

	// Datacenter to use. If not provided, the default agent datacenter is used.
	Datacenter string `json:"datacenter,omitempty"`

	// BasicAuth sets the HTTP BasicAuth credentials for talking to Consul
	BasicAuth *BasicAuthConfig `json:"basic_auth,omitempty"`

	// Token is used to provide a per-request ACL token
	// which overrides the agent's default token.
	Token string `json:"token,omitempty"`

	// TokenFile is a file containing the current token to use for this client.
	// If provided it is read once at startup and never again.
	TokenFile string `json:"token_file,omitempty"`

	// Namespace is the name of the namespace to send along for the request
	// when no other Namespace is present in the QueryOptions
	Namespace string `json:"namespace,omitempty"`

	// Partition is the name of the partition to send along for the request
	// when no other Partition is present in the QueryOptions
	Partition string `json:"partition,omitempty"`

	// TLSConfig is the TLS config for talking to Consul
	TLSConfig *TLSConfig `json:"tls_config,omitempty"`
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
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, err
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
		leaderKey: fmt.Sprintf("/%s/leader", key),
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
		apiConfig.TLSConfig.CAFile = cfg.TLSConfig.CertFile
		apiConfig.TLSConfig.CertPEM = cfg.TLSConfig.CertPEM
		apiConfig.TLSConfig.KeyFile = cfg.TLSConfig.KeyFile
		apiConfig.TLSConfig.KeyPEM = cfg.TLSConfig.KeyPEM
		apiConfig.TLSConfig.InsecureSkipVerify = cfg.TLSConfig.InsecureSkipVerify
	}

	return apiConfig
}
