package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/rqlite/rqlite-disco-clients/expand"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Client represents an etcd client.
type Client struct {
	client    *clientv3.Client
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
// returns a Config. A nil reader results in a nil config.
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

// New returns an instantiated etcd client. If cfg is nil, use
// the default config.
func New(key string, cfg *Config) (*Client, error) {
	c, err := clientv3.New(*etcdConfigFromClientConfig(cfg))
	if err != nil {
		return nil, err
	}
	return &Client{
		client:    c,
		key:       key,
		leaderKey: fmt.Sprintf("/%s/leader", key),
	}, nil
}

// GetLeader returns the leader as recorded in Consul. If a leader exists, ok will
// be set to true, false otherwise.
func (c *Client) GetLeader() (id string, apiAddr string, addr string, ok bool, e error) {
	kv := clientv3.NewKV(c.client)
	resp, err := kv.Get(context.Background(), c.leaderKey)
	if err != nil {
		e = err
		return
	}
	if len(resp.Kvs) == 0 || resp.Kvs[0].Value == nil {
		ok = false
		return
	}

	n := node{}
	if err := json.Unmarshal(resp.Kvs[0].Value, &n); err != nil {
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

	kv := clientv3.NewKV(c.client)
	resp, err := kv.Txn(context.TODO()).
		If(clientv3.Compare(clientv3.Version(c.leaderKey), "=", 0)).
		Then(
			clientv3.OpPut(c.leaderKey, string(b))).Commit()
	if err != nil {
		return false, err
	}

	return resp.Succeeded, nil
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

	kv := clientv3.NewKV(c.client)
	_, err = kv.Put(context.Background(), c.leaderKey, string(b))
	if err != nil {
		return err
	}
	return nil
}

// String implements the Stringer interface.
func (c *Client) String() string {
	return "etcd-kv"
}

// Close closes the client.
func (c *Client) Close() error {
	return c.client.Close()
}

func etcdConfigFromClientConfig(cfg *Config) *clientv3.Config {
	if cfg == nil {
		return &clientv3.Config{
			Endpoints: []string{"localhost:2379"},
		}
	}
	return (*clientv3.Config)(cfg)
}

type node struct {
	ID      string `json:"id,omitempty"`
	APIAddr string `json:"api_addr,omitempty"`
	Addr    string `json:"addr,omitempty"`
}
