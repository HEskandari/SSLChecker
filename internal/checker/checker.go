package checker

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"time"

	"github.com/hadi/ssl-cert-monitor/internal/config"
)

// Checker performs SSL certificate checks
type Checker struct{}

// NewChecker creates a new Checker instance
func NewChecker() *Checker {
	return &Checker{}
}

// CheckDomain performs a TLS handshake and extracts certificate expiry
func (c *Checker) CheckDomain(domain config.DomainConfig) config.CheckResult {
	result := config.CheckResult{
		Domain: domain,
	}

	address := fmt.Sprintf("%s:%d", domain.Host, domain.Port)
	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: 10 * time.Second},
		"tcp",
		address,
		&tls.Config{
			InsecureSkipVerify: domain.InsecureSkipVerify,
		},
	)
	if err != nil {
		result.Success = false
		result.Error = fmt.Errorf("TLS handshake failed: %w", err)
		return result
	}
	defer conn.Close()

	// Get the certificate chain
	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		result.Success = false
		result.Error = fmt.Errorf("no certificates presented")
		return result
	}

	// Use the leaf certificate (first in chain)
	cert := certs[0]
	result.Expiry = cert.NotAfter
	result.DaysRemaining = time.Until(cert.NotAfter).Hours() / 24
	result.Success = true

	return result
}

// VerifyCertificateChain attempts to verify the certificate chain
func (c *Checker) VerifyCertificateChain(domain config.DomainConfig) error {
	address := fmt.Sprintf("%s:%d", domain.Host, domain.Port)
	conn, err := tls.DialWithDialer(
		&net.Dialer{Timeout: 10 * time.Second},
		"tcp",
		address,
		&tls.Config{
			InsecureSkipVerify: false,
		},
	)
	if err != nil {
		return fmt.Errorf("TLS handshake failed: %w", err)
	}
	defer conn.Close()

	// Verify the certificate chain
	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return fmt.Errorf("no certificates presented")
	}

	// Create a certificate pool with system certs
	pool, err := x509.SystemCertPool()
	if err != nil {
		pool = x509.NewCertPool()
	}

	// Add intermediate certificates
	for i, cert := range certs {
		if i > 0 { // Skip leaf certificate
			pool.AddCert(cert)
		}
	}

	leaf := certs[0]
	_, err = leaf.Verify(x509.VerifyOptions{
		DNSName:       domain.Host,
		Intermediates: pool,
	})
	if err != nil {
		return fmt.Errorf("certificate verification failed: %w", err)
	}

	return nil
}