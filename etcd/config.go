package etcd

import (
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Config stores the configuration for the etcd client.
type Config clientv3.Config
