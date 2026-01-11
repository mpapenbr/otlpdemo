package config

type (
	DBConfig struct {
		Enabled       bool
		Host          string
		Port          int
		Database      string
		StaticSecrets DBSecrets // static secrets defined in config
		SecretsFile   string    // path to file with secrets (e.g., from Vault)
		SSLMode       string
		TLSCert       string // path to TLS certificate
		TLSKey        string // path to TLS key
		TLSCA         string // path to TLS CA (to validate server certificate)
	}
	// this holds secrets for the DB that are managed outside of the main config

	DBSecrets struct {
		User     string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		// TODO: mTLS settings
	}
)

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
	DBConf            DBConfig
)
