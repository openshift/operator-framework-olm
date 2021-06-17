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
	"sync"

	configv1 "github.com/openshift/api/config/v1"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
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
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
	Lock     sync.Mutex

	Name                string
	Namespace           string
	HighlyAvailableMode bool
}

func (r *PackageServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("deployment", req.NamespacedName)

	log.Info("handling current request", "request", req.String())
	log.Info("currently topology mode", "highly available", r.getHighlyAvailableMode())
	defer log.Info("finished request reconciliation")

	deployment := &appsv1.Deployment{}
	if err := r.Client.Get(ctx, req.NamespacedName, deployment); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !ensureDeployment(deployment, r.getHighlyAvailableMode()) {
		log.Info("no updates are required for the deployment")
		return ctrl.Result{}, nil
	}

	if err := r.Client.Update(ctx, deployment); err != nil {
		log.Error(err, "failed to update the deployment")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func ensureDeployment(deployment *appsv1.Deployment, highlyAvailableMode bool) bool {
	var modified bool

	expectedReplicas := int32(defaultReplicaCount)
	if !highlyAvailableMode {
		expectedReplicas = int32(singleReplicaCount)
	}
	if *deployment.Spec.Replicas != expectedReplicas {
		deployment.Spec.Replicas = pointer.Int32Ptr(expectedReplicas)
		modified = true
	}

	intStr := intstr.FromInt(1)
	expectedRolloutConfiguration := &appsv1.RollingUpdateDeployment{
		MaxUnavailable: &intStr,
		MaxSurge:       &intStr,
	}
	if !highlyAvailableMode {
		expectedRolloutConfiguration = &appsv1.RollingUpdateDeployment{}
	}
	if deployment.Spec.Strategy.RollingUpdate != expectedRolloutConfiguration {
		deployment.Spec.Strategy.RollingUpdate = expectedRolloutConfiguration
		modified = true
	}

	return modified
}

func (r *PackageServerReconciler) infrastructureHandler(obj client.Object) []reconcile.Request {
	log := r.Log.WithValues("infrastructure", obj.GetName())

	if obj.GetName() != infrastructureName {
		return nil
	}

	var infra configv1.Infrastructure
	if err := r.Client.Get(context.Background(), types.NamespacedName{Name: obj.GetName()}, &infra); err != nil {
		return []reconcile.Request{}
	}

	r.setTopologyMode(infra.Status.ControlPlaneTopology)
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

func (r *PackageServerReconciler) setTopologyMode(topologyMode configv1.TopologyMode) {
	r.Lock.Lock()
	defer r.Lock.Unlock()

	if topologyMode == configv1.SingleReplicaTopologyMode {
		r.HighlyAvailableMode = false
		return
	}
	r.HighlyAvailableMode = true
}

func (r *PackageServerReconciler) getHighlyAvailableMode() bool {
	r.Lock.Lock()
	defer r.Lock.Unlock()

	return r.HighlyAvailableMode
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
