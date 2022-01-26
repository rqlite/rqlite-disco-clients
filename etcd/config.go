package etcd

import (
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	// exampleConfig is an example of how the etcd config file
	// should be structured. The time-related values are in
	// units of nanoseconds.
	exampleConfig = `
{
	"endpoints": ["http://1.2.3.4:8080", "https://5.6.7.8"],
	"auto-sync-interval": 10000,
	"dial-timeout": 30000,
	"dial-keep-alive-timeout": 900000,
	"username": "me",
	"password": "my password",
	"reject-old-cluster": true
}
`
)

// Config stores the configuration for the etcd client.
// The full definition is available at https://pkg.go.dev/go.etcd.io/etcd/clientv3#Config
type Config clientv3.Config
