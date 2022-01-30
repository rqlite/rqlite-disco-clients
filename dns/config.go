package dns

const (
	// exampleConfig is an example of how the DNS config file
	// should be structured.
	exampleConfig = `
{
	"name": "rqlite",
	"port": 4001
}
`
)

type Config struct {
	// Name is the hostname to resolve for node addresses.
	Name string `json:"name",omitempty"`

	// Port is the port resolved names will be listening on.
	Port int `json:"port",omitempty`
}
