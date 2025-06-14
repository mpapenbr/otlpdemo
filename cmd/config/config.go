package config

var (
	EnableTelemetry   bool
	TelemetryEndpoint string
	LogConfig         string
	LogLevel          string
	Insecure          bool     // connect to server without TLS
	TLSMinVersion     string   // minimum TLS version (e.g., "TLS13")
	TLSSkipVerify     bool     // skip TLS verification
	TLSCert           string   // path to TLS certificate
	TLSKey            string   // path to TLS key
	TLSCAs            []string // path to TLS CA (to validate server certificate)
	TLSClientCAs      []string // path to TLS CA (to validate client certificate)
	TLSClientAuth     string   // TLS client authentication mode
	Address           string   // address to listen on/connect to
	OtelOutput        string   // output for otel-logger (stdout, grpc)
)
