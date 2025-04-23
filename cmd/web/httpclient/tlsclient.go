package httpclient

import (
	"context"
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

func queryWithTLS() {
	myTLS, err := config.BuildTLSConfig()
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
