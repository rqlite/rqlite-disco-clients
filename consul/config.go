package consul

const (
	// exampleConfig is an example of how the Consul config file
	// should be structured.
	exampleConfig = `
{
	"address": "1.2.3.4:8500",
	"scheme": "https",
	"datacenter": "my_dc",
	"basic_auth": {
		"username": "me",
		"password": "my password"
	},
	"token": "my_token",
	"token_file": "my_token_file",
	"namespace": "my_namespace",
	"partition": "my_partition",
	"tls_config": {
		"insecure_skip_verify": true
	}
}
`
)

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
	Scheme string `json:"scheme,omitempty"`

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
