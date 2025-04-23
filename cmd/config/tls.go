package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
)

//nolint:nestif // false positive
func BuildTLSConfig() (*tls.Config, error) {
	if Insecure {
		return nil, nil
	} else {
		tlsConfig := &tls.Config{
			MinVersion: tls.VersionTLS13, // Set the minimum TLS version to TLS 1.3
		}
		if TLSCert != "" && TLSKey != "" {
			cert, err := tls.LoadX509KeyPair(TLSCert, TLSKey)
			if err != nil {
				return nil, err
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
		if TLSCa != "" {
			caCert, err := os.ReadFile(TLSCa)
			if err != nil {
				return nil, err
			}
			caCertPool := x509.NewCertPool()
			if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
				return nil, fmt.Errorf("failed to append server certificate")
			}
			// this is used on the server side to verify the client certificate
			tlsConfig.ClientCAs = caCertPool
			// this is used on the client side to verify the server certificate
			tlsConfig.RootCAs = caCertPool
		}
		if TLSClientAuth != "" {
			clientAuth, err := ParseClientAuth(TLSClientAuth)
			if err != nil {
				return nil, err
			}
			tlsConfig.ClientAuth = clientAuth
		}
		if TLSSkipVerify {
			tlsConfig.InsecureSkipVerify = true
		}
		return tlsConfig, nil
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
