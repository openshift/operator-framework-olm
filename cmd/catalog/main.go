package main

import (
	"context"

	configv1client "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/operator-framework/operator-lifecycle-manager/pkg/cmd/catalog"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/controller/registry/resolver"
)

var versionClient configv1client.ClusterVersionsGetter = configv1client.NewForConfigOrDie(config.GetConfigOrDie())

func main() {
	// TODO(njhale): use signals context

	// Add OpenShift-specific dependency/update resolution constraints before starting the catalog operator
	resolver.AddSystemConstraintProviders(
		prohibitIncompatible(context.Background(), versionClient),
	)

	catalog.Exec()
}
