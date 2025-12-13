//nolint:gosec // ignore G404 (rand) mainly
package webserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"sync"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

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

type itemConfig struct {
	Name      string
	URL       string
	ErrorRate int
	MaxID     int
}
type itemType int

const (
	NoItem itemType = iota
	CommentItem
	PostItem
	PhotoItem
)

var conf map[itemType]itemConfig = map[itemType]itemConfig{
	CommentItem: {
		Name:      "comment",
		URL:       "https://jsonplaceholder.typicode.com/comments",
		ErrorRate: 10,
		MaxID:     500,
	},
	PostItem: {
		Name:      "post",
		URL:       "https://jsonplaceholder.typicode.com/posts",
		ErrorRate: 20,
		MaxID:     100,
	},
	PhotoItem: {
		Name:      "photo",
		URL:       "https://jsonplaceholder.typicode.com/photos",
		ErrorRate: 15,
		MaxID:     5000,
	},
}

var tracer = otel.Tracer("webserver")

//nolint:lll // readability
func simpleWebserver() {
	fmt.Printf("Starting server on %s\n", config.Address)
	myTLS, err := config.BuildServerTLSConfig()
	if err != nil {
		log.Error("TLS config error", log.ErrorField(err))
		return
	}

	mux := http.NewServeMux()
	addToMux(mux, "/hello", hello(myTLS))
	addToMux(mux, "/relay/comment", relayComment())
	addToMux(mux, "/relay/post", relayPost())
	addToMux(mux, "/relay/photo", relayPhoto())
	addToMux(mux, "/relay/random", relayRandom())
	addToMux(mux, "/relay/sequence", relaySequence())
	addToMux(mux, "/relay/concurrent", relayConcurrent())
	mainHander := otelhttp.NewHandler(mux, "oteldemo-webserver",
		otelhttp.WithMessageEvents(
			otelhttp.ReadEvents,
			otelhttp.WriteEvents))
	if config.Insecure {
		log.Info("Using insecure mode with http")
		if err = http.ListenAndServe(config.Address, mainHander); err != nil {
			log.Error("Error starting http server", log.ErrorField(err))
			return
		}
	} else {
		log.Info("TLS config present. Server accepts TLS connections only")
		server := &http.Server{
			Addr:      config.Address,
			Handler:   mainHander,
			TLSConfig: myTLS,
		}
		if err = server.ListenAndServeTLS(config.TLSCert, config.TLSKey); err != nil {
			log.Error("Error starting TLS server", log.ErrorField(err))
			return
		}
	}
}

func addToMux(mux *http.ServeMux, pattern string, handler http.Handler) {
	mux.Handle(pattern,
		TraceIDMiddleware(LoggingMiddleware(handler)))
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fields := []log.Field{
			log.String("method", r.Method),
			log.String("url", r.URL.String()),
			log.String("remoteAddr", r.RemoteAddr),
		}
		span := trace.SpanFromContext(r.Context())
		if span.SpanContext().HasTraceID() {
			fields = append(fields,
				log.Any("ctx", r.Context()),
				log.String("trace_id", span.SpanContext().TraceID().String()))
		}
		log.Debug("Request received", fields...)

		next.ServeHTTP(w, r)
	})
}

func TraceIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := trace.SpanFromContext(r.Context())
		if span.SpanContext().HasTraceID() {
			w.Header().Set("X-Trace-ID", span.SpanContext().TraceID().String())
		}
		next.ServeHTTP(w, r)
	})
}

func hello(myTLS *tls.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		clientAuth, _ := config.ParseClientAuth(config.TLSClientAuth)
		switch clientAuth {
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
	}
}

func relayComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fetchWithErrorRate(r.Context(), CommentItem)
	}
}

func relayPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fetchWithErrorRate(r.Context(), PostItem)
	}
}

func relayPhoto() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fetchWithErrorRate(r.Context(), PhotoItem)
	}
}

func relaySequence() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "relay sequence")
		defer span.End()
		fetchWithErrorRate(ctx, CommentItem)
		fetchWithErrorRate(ctx, PostItem)
		fetchWithErrorRate(ctx, PhotoItem)
	}
}

func relayConcurrent() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "relay concurrent")
		defer span.End()
		items := []itemType{CommentItem, PostItem, PhotoItem}
		wg := &sync.WaitGroup{}
		wg.Add(len(items))
		span.AddEvent("spawning go routines", trace.WithAttributes(
			attribute.Int("count", len(items))))
		for _, item := range items {
			go func(item itemType) {
				defer wg.Done()
				fetchWithErrorRate(ctx, item)
			}(item)
		}
		span.AddEvent("waiting for go routines")
		wg.Wait()
		span.AddEvent("go routines finished")
	}
}

func relayRandom() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idx := rand.IntN(3)

		_, span := tracer.Start(r.Context(), "pick random")
		defer span.End()
		switch idx {
		case 0:
			span.AddEvent("picked comment")
			relayComment().ServeHTTP(w, r)
		case 1:
			span.AddEvent("picked post")
			relayPost().ServeHTTP(w, r)
		case 2:
			span.AddEvent("picked photo")
			relayPhoto().ServeHTTP(w, r)
		}
	}
}

func fetchWithErrorRate(ctx context.Context, item itemType) {
	c := conf[item]
	idx := rand.IntN(c.MaxID) + 1

	spanCtx, span := tracer.Start(ctx, fmt.Sprintf("fetch %s", c.Name))
	defer span.End()

	if rand.IntN(c.ErrorRate) == 0 {
		idx = 9999
		span.AddEvent("picked invalid index")
	} else {
		span.AddEvent("picked", trace.WithAttributes(
			attribute.String("item", c.Name),
			attribute.Int("id", idx)))
	}
	//nolint:errcheck // dummy
	doCall(spanCtx,
		fmt.Sprintf("%s/%d", c.URL, idx),
		fmt.Sprintf("fetching %s", c.Name))
}

func doCall(ctx context.Context, url, operation string) (ret []byte, err error) {
	ctx, span := tracer.Start(ctx, operation)
	defer span.End()
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		http.NoBody)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	client := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	span.AddEvent("got result")
	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	attrs := []attribute.KeyValue{
		attribute.Int("status", resp.StatusCode),
		attribute.Int("bytes", len(body)),
	}
	span.SetAttributes(attrs...)

	if err != nil {
		span.RecordError(err)
		return nil, err
	}
	log.Debug("request done",
		log.Any("spanCtx", ctx),
		log.Int("status", resp.StatusCode), log.Int("bytes", len(body)))
	return body, nil
}
