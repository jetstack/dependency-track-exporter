package dependencytrack

import "os"

const (
	// DefaultAddress is the default address of the Dependency-Track server.
	DefaultAddress string = "http://localhost:8080"

	// EnvAddress sets the address the Dependency-Track server.
	EnvAddress string = "DEPENDENCY_TRACK_ADDR"

	// EnvAPIKey sets the api key for the Dependency-Track API
	EnvAPIKey string = "DEPENDENCY_TRACK_API_KEY"
)

// Option configures a Client
type Option func(*options)

type options struct {
	Address string
	APIKey  string
}

func makeOptions(opts ...Option) *options {
	o := &options{
		Address: getEnv(EnvAddress, DefaultAddress),
		APIKey:  getEnv(EnvAPIKey, ""),
	}

	for _, option := range opts {
		option(o)
	}

	return o
}

func getEnv(envVar, fallback string) string {
	v := os.Getenv(envVar)
	if v == "" {
		v = fallback
	}

	return v
}

// WithAddress sets the address
func WithAddress(addr string) Option {
	return func(o *options) {
		o.Address = addr
	}
}

// WithAPIKey sets the API key
func WithAPIKey(apiKey string) Option {
	return func(o *options) {
		o.APIKey = apiKey
	}
}
