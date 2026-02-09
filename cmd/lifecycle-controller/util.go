package main

import (
	"fmt"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/selection"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	configv1 "github.com/openshift/api/config/v1"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
)

func setupScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(operatorsv1alpha1.AddToScheme(scheme))
	utilruntime.Must(configv1.AddToScheme(scheme))

	return scheme
}

// catalogPodLabelSelector returns a label selector matching pods with olm.catalogSource label
func catalogPodLabelSelector() labels.Selector {
	// This call cannot fail: the label key is valid and selection.Exists requires no values.
	req, err := labels.NewRequirement("olm.catalogSource", selection.Exists, nil)
	if err != nil {
		// Panic on impossible error to satisfy static analysis and catch programming errors
		panic(fmt.Sprintf("BUG: failed to create label requirement: %v", err))
	}
	return labels.NewSelector().Add(*req)
}
