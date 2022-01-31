package dnssrv

const (
	// exampleConfig is an example of how the DNS SRV config file
	// should be structured.
	exampleConfig = `
{
	"name": "rqlite",
	"service": "rqlite"
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
