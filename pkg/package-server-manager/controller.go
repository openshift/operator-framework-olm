/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"sync"

	"github.com/go-logr/logr"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/operator-framework-olm/pkg/manifests"
	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	infrastructureName  = "cluster"
	defaultReplicaCount = int32(2)
	defaultRolloutCount = 1
	singleReplicaCount  = int32(1)
)

// PackageServerCSVReconciler reconciles the PackageServer CSV object
type PackageServerCSVReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Lock   sync.Mutex

	Name      string
	Namespace string
	Image     string
}

// Reconcile is part of the main kubernetes reconciliation loop which is responsible
// for ensuring that the PackageServer CSV is deployed in a HA or non-HA mode depending
// on the topology mode exposed by the current state of the Infrastructure cluster
// singleton resource.
func (r *PackageServerCSVReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("csv", req.NamespacedName)

	log.Info("handling current request", "request", req.String())
	defer log.Info("finished request reconciliation")

	var infra configv1.Infrastructure
	if err := r.Client.Get(ctx, types.NamespacedName{Name: infrastructureName}, &infra); err != nil {
		return ctrl.Result{}, err
	}
	highAvailabilityMode := getTopologyModeFromInfra(&infra)
	log.Info("currently topology mode", "highly available", highAvailabilityMode)

	required, err := manifests.NewPackageServerCSV(
		manifests.WithName(r.Name),
		manifests.WithNamespace(r.Namespace),
		manifests.WithImage(r.Image),
	)
	if err != nil {
		log.Error(err, "failed to serialize a new packageserver csv from the base YAML manifest")
		return ctrl.Result{}, err
	}
	res, err := controllerutil.CreateOrUpdate(ctx, r.Client, required, func() error {
		return reconcileCSV(r.Log, r.Image, required, highAvailabilityMode)
	})

	log.Info("reconciliation result", "res", res)
	if err != nil {
		log.Error(err, "failed to create or update the packageserver csv")
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

func reconcileCSV(log logr.Logger, image string, csv *olmv1alpha1.ClusterServiceVersion, highAvailabilityMode bool) error {
	if csv.ObjectMeta.CreationTimestamp.IsZero() {
		log.Info("attempting to create the packageserver csv")
	}

	modified, err := ensureCSV(log, image, csv, highAvailabilityMode)
	if err != nil {
		return fmt.Errorf("error ensuring CSV: %v", err)
	}

	if !modified {
		log.V(3).Info("no further updates are necessary to the packageserver csv")
	}

	return nil
}

func (r *PackageServerCSVReconciler) infrastructureHandler(_ context.Context, obj client.Object) []reconcile.Request {
	log := r.Log.WithValues("infrastructure", obj.GetName())
	if obj.GetName() != infrastructureName {
		log.Info("not processing events for the non-cluster infrastructure resource")
		return nil
	}

	log.Info("requeueing the packageserver deployment after encountering infrastructure event")
	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      r.Name,
				Namespace: r.Namespace,
			},
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *PackageServerCSVReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&olmv1alpha1.ClusterServiceVersion{}).
		Watches(&configv1.Infrastructure{}, handler.EnqueueRequestsFromMapFunc(r.infrastructureHandler)).
		Complete(r)
}
