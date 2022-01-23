package etcd

import (
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Config stores the configuration for the etcd client.
// The full definition is available at https://pkg.go.dev/go.etcd.io/etcd/clientv3#Config
type Config clientv3.Config
