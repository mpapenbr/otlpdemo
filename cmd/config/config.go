package config

var (
	EnableTelemetry   bool
	TelemetryEndpoint string
	Worker            int
	LogConfig         string
	LogFile           string
	LogLevel          string
	Insecure          bool   // connect to server without TLS
	TLSSkipVerify     bool   // skip TLS verification
	TLSCert           string // path to TLS certificate
	TLSKey            string // path to TLS key
	TLSCa             string // path to TLS CA
	TLSClientAuth     string // TLS client authentication mode
	Address           string // address to listen on/connect to
	EnableOtelLogger  bool   // if true, otel-logger is setup in rootCmd
	OtelOuput         string // output for otel-logger (stdout, grpc)
)
