package controllers

import (
	"crypto/tls"
	"sync"
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/stretchr/testify/require"
)

func dummyGetCertificate(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
	return &tls.Certificate{}, nil
}

func tls12Profile() configv1.TLSProfileSpec {
	return configv1.TLSProfileSpec{
		MinTLSVersion: configv1.VersionTLS12,
		Ciphers: []string{
			"TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
			"TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384",
		},
	}
}

func tls13Profile() configv1.TLSProfileSpec {
	return configv1.TLSProfileSpec{
		MinTLSVersion: configv1.VersionTLS13,
	}
}

func TestNewTLSConfigProvider(t *testing.T) {
	profile := tls12Profile()
	p := NewTLSConfigProvider(dummyGetCertificate, profile)

	require.NotNil(t, p)

	cfg, unsupported := p.Get()
	require.NotNil(t, cfg)
	require.Equal(t, uint16(tls.VersionTLS12), cfg.MinVersion)
	require.Empty(t, unsupported)
}

func TestTLSConfigProvider_Get_ReturnsClonedConfig(t *testing.T) {
	profile := tls12Profile()
	p := NewTLSConfigProvider(dummyGetCertificate, profile)

	cfg1, _ := p.Get()
	cfg2, _ := p.Get()

	// Modifying the returned config should not affect the provider
	cfg1.MinVersion = tls.VersionTLS11
	cfg2After, _ := p.Get()
	require.Equal(t, uint16(tls.VersionTLS12), cfg2After.MinVersion)

	// Two successive Gets should return equivalent but distinct configs
	require.NotSame(t, cfg1, cfg2)
}

func TestTLSConfigProvider_UpdateProfile(t *testing.T) {
	initialProfile := tls12Profile()
	p := NewTLSConfigProvider(dummyGetCertificate, initialProfile)

	cfg, _ := p.Get()
	require.Equal(t, uint16(tls.VersionTLS12), cfg.MinVersion)

	// Update to TLS 1.3
	newProfile := tls13Profile()
	p.UpdateProfile(newProfile)

	cfg, _ = p.Get()
	require.Equal(t, uint16(tls.VersionTLS13), cfg.MinVersion)
}

func TestTLSConfigProvider_GetCertificatePreserved(t *testing.T) {
	called := false
	getCert := func(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
		called = true
		return &tls.Certificate{}, nil
	}

	p := NewTLSConfigProvider(getCert, tls12Profile())
	cfg, _ := p.Get()

	require.NotNil(t, cfg.GetCertificate)
	_, err := cfg.GetCertificate(nil)
	require.NoError(t, err)
	require.True(t, called, "getCertificate function should be preserved in config")
}

func TestTLSConfigProvider_ConcurrentAccess(t *testing.T) {
	p := NewTLSConfigProvider(dummyGetCertificate, tls12Profile())

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	// Half the goroutines read, half update
	for i := range goroutines {
		go func() {
			defer wg.Done()
			cfg, _ := p.Get()
			require.NotNil(t, cfg)
		}()
		go func(i int) {
			defer wg.Done()
			var profile configv1.TLSProfileSpec
			if i%2 == 0 {
				profile = tls12Profile()
			} else {
				profile = tls13Profile()
			}
			p.UpdateProfile(profile)
		}(i)
	}

	wg.Wait()

	// Provider should still be functional after concurrent access
	cfg, _ := p.Get()
	require.NotNil(t, cfg)
}
