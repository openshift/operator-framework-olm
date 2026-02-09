package controllers

import (
	"crypto/tls"
	"slices"
	"sync"

	configv1 "github.com/openshift/api/config/v1"
	tlsutil "github.com/openshift/controller-runtime-common/pkg/tls"
)

// TLSConfigProvider provides thread-safe access to dynamically updated TLS configuration.
// It implements controllers.TLSConfigProvider interface.
type TLSConfigProvider struct {
	mu                 sync.RWMutex
	getCertificateFunc func(info *tls.ClientHelloInfo) (*tls.Certificate, error)

	profileSpec configv1.TLSProfileSpec

	tlsConfig          *tls.Config
	unsupportedCiphers []string
}

// NewTLSConfigProvider creates a new TLSConfigProvider with the given initial profileSpec.
func NewTLSConfigProvider(getCertificateFunc func(*tls.ClientHelloInfo) (*tls.Certificate, error), initial configv1.TLSProfileSpec) *TLSConfigProvider {
	p := &TLSConfigProvider{getCertificateFunc: getCertificateFunc}
	p.UpdateProfile(initial)
	return p
}

// Get returns the current TLS configuration.
func (p *TLSConfigProvider) Get() (*tls.Config, []string) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.tlsConfig.Clone(), slices.Clone(p.unsupportedCiphers)
}

// UpdateProfile sets a new TLS profile spec.
func (p *TLSConfigProvider) UpdateProfile(profileSpec configv1.TLSProfileSpec) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.profileSpec = profileSpec

	p.tlsConfig, p.unsupportedCiphers = p.generateTLSConfig()
}

func (p *TLSConfigProvider) generateTLSConfig() (*tls.Config, []string) {
	tlsConfigFunc, unsupportedCiphers := tlsutil.NewTLSConfigFromProfile(p.profileSpec)
	tlsConfig := &tls.Config{
		GetCertificate: p.getCertificateFunc,
	}
	tlsConfigFunc(tlsConfig)
	return tlsConfig, unsupportedCiphers
}
