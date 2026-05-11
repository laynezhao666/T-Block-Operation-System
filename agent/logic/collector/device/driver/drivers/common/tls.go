// Package common provides reusable utilities for HTTP-based drivers
package common

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"

	"trpc.group/trpc-go/trpc-go/log"
)

// ConfigureTLS configures TLS for HTTP client with CA certificate
// Parameters:
// - client: HTTP client to configure
// - extendKV: extended key-value map containing CA certificate (base64 encoded)
// - localCertPath: local certificate file path as fallback
// Returns:
// - error: configuration error
func ConfigureTLS(client *http.Client, extendKV map[string]string, localCertPath string) error {
	var caCert []byte
	var err error

	// Try to get CA certificate from config
	if extendKV != nil {
		if val, has := extendKV["ca.crt"]; has && len(val) != 0 {
			caCert, err = base64.StdEncoding.DecodeString(val)
			if err != nil {
				log.Warnf("Error decoding ca.crt: %v", err)
				return err
			}
		}
	}

	// Fall back to local file if not provided in config
	if len(caCert) == 0 && localCertPath != "" {
		caCert, err = os.ReadFile(localCertPath)
		if err != nil {
			return fmt.Errorf("error reading CA certificate from %s: %v", localCertPath, err)
		}
	}

	if len(caCert) == 0 {
		return fmt.Errorf("no CA certificate provided")
	}

	// Create certificate pool and add CA certificate
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Create TLS configuration
	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}

	// Create transport and set to client
	tr := &http.Transport{TLSClientConfig: tlsConfig}
	client.Transport = tr

	return nil
}

// ConfigureTLSInsecure configures TLS for HTTP client with insecure skip verify
// Use this only when certificate validation is not required
// Parameters:
// - client: HTTP client to configure
func ConfigureTLSInsecure(client *http.Client) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	tr := &http.Transport{TLSClientConfig: tlsConfig}
	client.Transport = tr
}
