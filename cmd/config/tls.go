package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/mpapenbr/otlpdemo/log"
)

func BuildServerTLSConfig() (*tls.Config, error) {
	if Insecure {
		log.Debug("using insecure mode. no TLS")
		return nil, nil
	} else {
		reloader, err := NewTLSReloader()
		if err != nil {
			log.Error("error creating TLS reloader", log.ErrorField(err))
			return nil, err
		}
		go reloader.watch()
		return &tls.Config{
			MinVersion:         tls.VersionTLS13,
			GetConfigForClient: reloader.GetConfigForClient,
		}, nil
	}
}

//nolint:nestif // false positive
func BuildClientTLSConfig() (*tls.Config, error) {
	if Insecure {
		log.Debug("using insecure mode. no TLS")
		return nil, nil
	} else {
		return buildTLSFromConfig()
	}
}

// used for gRPC
func BuildTransportCredentials() (credentials.TransportCredentials, error) {
	myTLS, err := BuildServerTLSConfig()
	if err != nil {
		return nil, err
	}
	if myTLS == nil {
		return insecure.NewCredentials(), nil
	} else {
		log.Debug("TLS configured")
		return credentials.NewTLS(myTLS), nil
	}
}

func ParseClientAuth(mode string) (tls.ClientAuthType, error) {
	switch strings.ToLower(mode) {
	case "none":
		return tls.NoClientCert, nil
	case "request":
		return tls.RequestClientCert, nil
	case "require":
		return tls.RequireAnyClientCert, nil
	case "verify-if-given":
		return tls.VerifyClientCertIfGiven, nil
	case "require-and-verify":
		return tls.RequireAndVerifyClientCert, nil
	default:
		return 0, fmt.Errorf("unknown client auth mode: %s", mode)
	}
}

func ParseTLSVersion(mode string) (uint16, error) {
	switch strings.ToLower(mode) {
	case "tls13":
		return tls.VersionTLS13, nil
	case "tls12":
		return tls.VersionTLS12, nil
	default:
		return 0, fmt.Errorf("unsupported TLS version: %s", mode)
	}
}

//nolint:funlen // by design
func buildTLSFromConfig() (*tls.Config, error) {
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS13,
	}
	if minVersion, err := ParseTLSVersion(TLSMinVersion); err == nil {
		tlsConfig.MinVersion = minVersion
	} else {
		return nil, err
	}
	if TLSCert != "" && TLSKey != "" {
		log.Debug("cert and key provided")
		cert, err := tls.LoadX509KeyPair(TLSCert, TLSKey)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	if len(TLSCAs) > 0 {
		log.Debug("ca provided")
		caCertPool := x509.NewCertPool()
		for _, ca := range TLSCAs {
			log.Debug("loading CA", log.String("ca", ca))
			caCert, err := os.ReadFile(ca)
			if err != nil {
				return nil, err
			}
			if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
				return nil, fmt.Errorf("failed to append server certificate")
			}
		}
		// this is used on the client side to verify the server certificate
		tlsConfig.RootCAs = caCertPool
	}
	if len(TLSClientCAs) > 0 {
		log.Debug("client ca provided")
		caCertPool := x509.NewCertPool()
		for _, ca := range TLSClientCAs {
			log.Debug("loading client CA", log.String("ca", ca))
			caCert, err := os.ReadFile(ca)
			if err != nil {
				return nil, err
			}
			if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
				return nil, fmt.Errorf("failed to append client certificate")
			}
		}
		// this is used on the server side to verify the client certificate
		tlsConfig.ClientCAs = caCertPool
	}
	if TLSClientAuth != "" {
		log.Debug("clientAuth provided")
		var clientAuth tls.ClientAuthType
		clientAuth, err := ParseClientAuth(TLSClientAuth)
		if err != nil {
			return nil, err
		}
		tlsConfig.ClientAuth = clientAuth
	}

	if TLSSkipVerify {
		log.Debug("skipVerify enabled")
		tlsConfig.InsecureSkipVerify = true
	}
	return tlsConfig, nil
}
