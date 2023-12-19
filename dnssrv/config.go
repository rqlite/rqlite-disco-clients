package dnssrv

const (
	// exampleConfig is an example of how the DNS SRV config file
	// should be structured. 'name' is the host to resolve for the
	// DNS records, and 'service' is the service to request.
	//
	// In the example below, rqlite would look up:
	//
	//  _rqlite-raft._tcp.rqlite.com
	//
	// and resolve the returned names for the actual
	// node IP addresses and ports. All other portions of the DNS
	// SRV record (priority, weight, TTL, etc.) are ignored.
	// Note that the 'proto' part of the DNS SRV hostname
	// is always 'tcp' and cannot be changed.
	exampleConfig = `
{
	"name": "rqlite.com",
	"service": "rqlite-raft"
}
`
)

// Config is the configuration for a DNS disco client.
type Config struct {
	// Name is the hostname to contact for DNS SRV records.
	Name string `json:"name,omitempty"`

	// Service is the service to request when making the
	// DNS SRV request.
	Service string `json:"service,omitempty"`
}
