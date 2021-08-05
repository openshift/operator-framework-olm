package main

import (
	"time"

	configv1 "github.com/openshift/api/config/v1"
	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

const (
	// Note: In order for SNO to GA, controllers need to handle ~60s of API server
	// disruptions when attempting to get and sustain leader election:
	// - https://github.com/openshift/library-go/pull/1104#discussion_r649313822
	// - https://bugzilla.redhat.com/show_bug.cgi?id=1985697
	defaultRetryPeriod   = 30 * time.Second
	defaultRenewDeadline = 60 * time.Second
	defaultLeaseDuration = 90 * time.Second
)

func timeDurationPtr(t time.Duration) *time.Duration {
	return &t
}

func setupScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(configv1.Install(scheme))
	utilruntime.Must(olmv1alpha1.AddToScheme(scheme))

	return scheme
}
