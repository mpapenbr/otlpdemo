package httpclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/mpapenbr/otlpdemo/cmd/config"
	"github.com/mpapenbr/otlpdemo/log"
)

var url string

func NewTLSClientCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "tlsclient",
		Short: "use a TLS client to connect to a server",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			queryWithTLS()
		},
	}
	cmd.Flags().StringVar(&url, "url", "", "url to connect to")
	return &cmd
}

//nolint:funlen // lots of stuff to do here
func queryWithTLS() {
	myTLS, err := config.BuildClientTLSConfig(
		func(c *tls.Config) {
			//nolint:whitespace // editor/linter issue
			c.VerifyPeerCertificate = func(
				rawCerts [][]byte,
				verifiedChains [][]*x509.Certificate,
			) error {
				log.Debug("custom certificate verification called")
				cert, err := x509.ParseCertificate(rawCerts[0])
				if err != nil {
					return err
				}

				log.Debug("Cert info", log.String("sn",
					hex.EncodeToString(cert.SerialNumber.Bytes())))
				return nil
			}
		},
	)
	if err != nil {
		log.Error("TLS config error", log.ErrorField(err))
		return
	}

	req, err := http.NewRequestWithContext(
		context.Background(),
		http.MethodGet,
		url,
		http.NoBody)
	if err != nil {
		log.Error("error creating request", log.ErrorField(err))
		return
	}
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: myTLS,
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("error executing request", log.ErrorField(err))
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	log.Debug("request done",
		log.Int("status", resp.StatusCode),
		log.Int("bytes", len(body)),
		log.String("body", string(body)),
	)
	resp.Body.Close()
}
