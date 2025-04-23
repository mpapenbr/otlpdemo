package webserver

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/mpapenbr/otlpdemo/cmd/config"
	"github.com/mpapenbr/otlpdemo/log"
)

func NewSimpleWebserverCommand() *cobra.Command {
	cmd := cobra.Command{
		Use:   "webserver",
		Short: "create a simple webserver",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			simpleWebserver()
		},
	}
	cmd.Flags().StringVar(&config.Address, "addr", "localhost:8080", "listen address")

	return &cmd
}

//nolint:funlen,gosec // by design,
func simpleWebserver() {
	fmt.Printf("Starting server on %s\n", config.Address)
	myTLS, err := config.BuildTLSConfig()
	if err != nil {
		log.Error("TLS config error", log.ErrorField(err))
		return
	}
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		if r.TLS == nil {
			log.Debug("No TLS connection")
			fmt.Fprint(w, "Hello, world! (no TLS)\n")
			return
		}
		showPeers := func(msg string) {
			log.Debug(msg,
				log.Int("peerCerts", len(r.TLS.PeerCertificates)))
			for i := 0; i < len(r.TLS.PeerCertificates); i++ {
				log.Debug("Client cert",
					log.String("subject", r.TLS.PeerCertificates[i].Subject.String()),
					log.String("issuer", r.TLS.PeerCertificates[i].Issuer.String()),
				)
			}
			fmt.Fprintf(w, "Hello, world! (%s, peer certs:%d)\n",
				msg,
				len(r.TLS.PeerCertificates))
		}
		switch myTLS.ClientAuth {
		case tls.NoClientCert:
			fmt.Fprint(w, "Hello, world! (no client cert required)\n")

		case tls.RequestClientCert:
			showPeers("request client cert")

		case tls.RequireAnyClientCert:
			showPeers("require any client cert")

		case tls.RequireAndVerifyClientCert:
			showPeers("require and verify client cert")

		case tls.VerifyClientCertIfGiven:
			showPeers("verify client cert if given")

		default:
			log.Debug("Unknown client auth mode",
				log.String("mode", myTLS.ClientAuth.String()))
			http.Error(w, "Hello, world! (unknown client auth mode)",
				http.StatusUnauthorized)
		}
	})
	if config.Insecure {
		log.Info("Using insecure mode with http")
		if err = http.ListenAndServe(config.Address, nil); err != nil {
			log.Error("Error starting http server", log.ErrorField(err))
			return
		}
	} else {
		log.Info("TLS config present. Server accepts TLS connections only")
		server := &http.Server{
			Addr:      config.Address,
			Handler:   nil, // Use the default ServeMux
			TLSConfig: myTLS,
		}
		if err = server.ListenAndServeTLS(config.TLSCert, config.TLSKey); err != nil {
			log.Error("Error starting TLS server", log.ErrorField(err))
			return
		}
	}
}
