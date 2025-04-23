package web

import (
	"github.com/spf13/cobra"

	"github.com/mpapenbr/otlpdemo/cmd/config"
	"github.com/mpapenbr/otlpdemo/cmd/web/httpclient"
	"github.com/mpapenbr/otlpdemo/cmd/web/webserver"
)

func NewWebCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "web",
		Short: "collection of web commands",
		Long:  ``,
	}
	cmd.PersistentFlags().BoolVar(&config.TLSSkipVerify,
		"tls-skip-verify",
		false,
		"skip verification of server certificate (used for development only)")
	cmd.PersistentFlags().StringVar(&config.TLSKey,
		"tls-key",
		"",
		"path to TLS key")
	cmd.PersistentFlags().StringVar(&config.TLSCert,
		"tls-cert",
		"",
		"path to TLS cert")
	cmd.PersistentFlags().StringVar(&config.TLSCa,
		"tls-ca",
		"",
		"path to TLS root certificate")
	cmd.PersistentFlags().BoolVar(&config.Insecure,
		"insecure",
		true,
		"don't use TLS (used for development only)")
	cmd.PersistentFlags().StringVar(&config.TLSClientAuth,
		"tls-client-auth",
		"",
		"how to handle the client cert (none, request, require-and-verify, verify-if-given)")

	cmd.AddCommand(httpclient.NewJSONPlaceholderCommand())
	cmd.AddCommand(httpclient.NewTLSClientCommand())
	cmd.AddCommand(webserver.NewSimpleWebserverCommand())
	return &cmd
}
