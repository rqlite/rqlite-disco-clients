package dns

const (
	// exampleConfig is an example of how the DNS config file
	// should be structured. In this example 'rqlite' is the
	// hostname the node will resolve for IP addresses of other
	// nodes it should attempt to form a cluster with. Port
	// is the HTTP[S] port those nodes will be listening on.
	// Because Port can only be set once, every node in the
	// cluster must listen to the same HTTP port. If you need
	// different ports for different rqlite nodes, then you
	// should use DNS SRV disco mode.
	exampleConfig = `
{
	"name": "rqlite",
	"port": 4002
}
`
)

// Config is the configuration for a DNS disco client.
type Config struct {
	// Name is the hostname to resolve for node addresses.
	Name string `json:"name,omitempty"`

	// Port is the port resolved names will be listening on.
	Port int `json:"port,omitempty"`
}
