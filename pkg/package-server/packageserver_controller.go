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
	_ "embed"
	"fmt"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/operator-framework-olm/pkg/manifests"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/utils/pointer"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	name                = "packageserver"
	infrastructureName  = "cluster"
	defaultReplicaCount = int32(2)
	singleReplicaCount  = int32(1)
)

// PackageServerReconciler reconciles the PackageServer deployment object
type PackageServerReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	Name      string
	Namespace string
	// TODO(tflannag): Should this be a custom struct instead?
	// TODO(tflannag): Add a mutex here?
	// TODO(tflannag): Add an event recorder?
	HighlyAvailableMode bool
}

func (r *PackageServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("deployment", req.NamespacedName)

	log.Info("handling current request", "request", req.String())
	log.Info("currently topology mode", "highly available", r.HighlyAvailableMode)
	defer log.Info("finished request reconciliation")

	deployment := &appsv1.Deployment{}
	err := r.Client.Get(ctx, req.NamespacedName, deployment)
	if err != nil && !apierrors.IsNotFound(err) {
		return ctrl.Result{}, err
	}
	if apierrors.IsNotFound(err) {
		deployment, err := r.syncDeployment()
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, r.Client.Create(ctx, deployment)
	}

	// TODO(tflannag): Need to handle updating any non-replica high availability fields
	// e.g. hard anti affinity, maxUnavailable, etc.
	expectedReplicas := int32(defaultReplicaCount)
	if !r.HighlyAvailableMode {
		expectedReplicas = int32(singleReplicaCount)
	}
	if *deployment.Spec.Replicas == expectedReplicas {
		log.Info("deployment does not require any updates")
		return ctrl.Result{}, nil
	}

	// TODO(tflannag): May want to avoid using Update and instead migrate towards
	// using SSA patching. Can also just nuke any spec fields that don't match
	// what's being defined in the default deployment.yaml.
	log.Info("updating replica count", "old replicas", *deployment.Spec.Replicas, "new replicas", expectedReplicas)
	deployment.Spec.Replicas = pointer.Int32Ptr(expectedReplicas)
	if err := r.Client.Update(ctx, deployment); err != nil {
		log.Error(err, "failed to update the packageserver deployment")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *PackageServerReconciler) syncDeployment() (*appsv1.Deployment, error) {
	deployment, err := manifests.NewPackageServerDeployment(r.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize deployment manifest into Go structure: %v", err)
	}

	if !r.HighlyAvailableMode {
		r.Log.Info("packageserver will be deployed in non-HA mode")
		deployment.Spec.Replicas = pointer.Int32Ptr(1)
	}

	return deployment, nil
}

func (r *PackageServerReconciler) infrastructureHandler(obj client.Object) []reconcile.Request {
	// TODO(tflannag): Is this thread safe? Do I need to instantiate a mutex in the reconciler struct?
	// TODO(tflannag): Is this an abuse of the mapping handler function? Should we be altering state here?
	log := r.Log.WithValues("infrastructure", obj.GetName())

	if obj.GetName() != infrastructureName {
		return nil
	}

	var infra configv1.Infrastructure
	if err := r.Client.Get(context.Background(), types.NamespacedName{Name: obj.GetName()}, &infra); err != nil {
		return []reconcile.Request{}
	}

	// TODO(tflannag): Build up a simple setter if we need a more concrete structure that
	// holds this information, vs. a single field?
	r.HighlyAvailableMode = true
	topologyMode := infra.Status.ControlPlaneTopology
	if topologyMode == configv1.SingleReplicaTopologyMode {
		r.HighlyAvailableMode = false
	}
	log.Info("requeueing the packageserver deployment after handling infrastructure event")

	return []reconcile.Request{
		{
			NamespacedName: types.NamespacedName{
				Name:      name,
				Namespace: r.Namespace,
			},
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *PackageServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	predicateNameFilter := predicate.NewPredicateFuncs(func(object client.Object) bool {
		return object.GetName() == name && object.GetNamespace() == r.Namespace
	})
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}, builder.WithPredicates(predicateNameFilter)).
		Watches(&source.Kind{Type: &configv1.Infrastructure{}}, handler.EnqueueRequestsFromMapFunc(r.infrastructureHandler)).
		Complete(r)
}
