package controllers

import (
	"context"
	"crypto/tls"
	"reflect"
	"sync"

	"github.com/go-logr/logr"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/library-go/pkg/crypto"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/lib/apiserver"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	// Name of the cluster-scoped APIServer resource
	clusterAPIServerName = "cluster"
)

// TLSConfig holds the TLS configuration extracted from the APIServer resource
type TLSConfig struct {
	minVersion   uint16
	cipherSuites []uint16
	// String representations for passing to lifecycle-server
	minVersionString   string
	cipherSuiteStrings []string
}

// TLSConfigProvider provides thread-safe access to dynamically updated TLS configuration.
// It implements controllers.TLSConfigProvider interface.
type TLSConfigProvider struct {
	mu     sync.RWMutex
	config *tls.Config
}

// NewTLSConfigProvider creates a new TLSConfigProvider with the given initial config.
func NewTLSConfigProvider(initial *tls.Config) *TLSConfigProvider {
	return &TLSConfigProvider{config: initial}
}

// Get returns the current TLS configuration.
func (p *TLSConfigProvider) Get() *tls.Config {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.config
}

// Update sets a new TLS configuration.
func (p *TLSConfigProvider) Update(cfg *tls.Config) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.config = cfg
}

// GetClusterTLSConfig reads the APIServer "cluster" resource and extracts TLS settings.
// Falls back to defaults if an error occurs looking up the apiserver config.
func GetClusterTLSConfig(ctx context.Context, cl client.Client, log logr.Logger) *tls.Config {
	var (
		apiServer    configv1.APIServer
		minVersion   uint16
		cipherSuites []uint16
	)
	if err := cl.Get(ctx, types.NamespacedName{Name: clusterAPIServerName}, &apiServer); err != nil {
		log.Error(err, "failed to lookup APIServer; using default TLS security profile")
		minVersion, cipherSuites = apiserver.GetSecurityProfileConfig(nil)
	} else {
		minVersion, cipherSuites = apiserver.GetSecurityProfileConfig(apiServer.Spec.TLSSecurityProfile)
	}

	log.Info("loaded TLS configuration from APIServer",
		"minVersion", crypto.TLSVersionToNameOrDie(minVersion),
		"cipherSuites", crypto.CipherSuitesToNamesOrDie(cipherSuites),
	)

	return &tls.Config{
		MinVersion:   minVersion,
		CipherSuites: cipherSuites,
	}
}

// ClusterTLSProfileReconciler watches the APIServer "cluster" resource and updates TLS config dynamically
type ClusterTLSProfileReconciler struct {
	Client      client.Client
	Log         logr.Logger
	TLSProvider *TLSConfigProvider
	OnChange    func(prev, cur *tls.Config)
}

func (r *ClusterTLSProfileReconciler) Reconcile(ctx context.Context, _ reconcile.Request) (reconcile.Result, error) {
	// Check if config changed
	oldConfig := r.TLSProvider.Get()
	newConfig := GetClusterTLSConfig(ctx, r.Client, r.Log)
	if reflect.DeepEqual(oldConfig, newConfig) {
		// No change
		return reconcile.Result{}, nil
	}

	r.Log.Info("TLS security profile changed, updating configuration and triggering reconciliation",
		"oldMinVersion", crypto.TLSVersionToNameOrDie(oldConfig.MinVersion),
		"newMinVersion", crypto.TLSVersionToNameOrDie(newConfig.MinVersion),
		"oldCipherSuites", crypto.CipherSuitesToNamesOrDie(oldConfig.CipherSuites),
		"newCipherSuites", crypto.CipherSuitesToNamesOrDie(newConfig.CipherSuites),
	)

	// Update the provider and call the OnChange callback
	r.TLSProvider.Update(newConfig)
	r.OnChange(oldConfig, newConfig)

	return reconcile.Result{}, nil
}

func (r *ClusterTLSProfileReconciler) SetupWithManager(mgr manager.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named("tlsprofile-reconciler").
		WatchesRawSource(source.Kind(mgr.GetCache(), &configv1.APIServer{},
			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, obj *configv1.APIServer) []reconcile.Request {
				if obj.Name == clusterAPIServerName {
					return []reconcile.Request{{NamespacedName: types.NamespacedName{Name: clusterAPIServerName}}}
				}
				return nil
			}),
		)).
		Complete(r)
}
