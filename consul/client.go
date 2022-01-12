package consul // Maybe put in own repo -- rqlite-disco-clients

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/consul/api"
)

// Client represents a Consul client.
type Client struct {
	client    *api.KV
	key       string
	leaderKey string
}

type node struct {
	ID      string `json:"id,omitempty"`
	APIAddr string `json:"api_addr,omitempty"`
	Addr    string `json:"addr,omitempty"` // Needs TLS settings, etc I think so anyway. Maybe join handles?
}

// New returns an instantiated Consul client.
func New(key string) (*Client, error) {
	c, err := api.NewClient(api.DefaultConfig())
	if err != nil {
		return nil, err
	}
	return &Client{
		client:    c.KV(),
		key:       key,
		leaderKey: fmt.Sprintf("rqlite/%s/leader", key),
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

// SetLeader unconditionally sets the leader to the gien details.
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
