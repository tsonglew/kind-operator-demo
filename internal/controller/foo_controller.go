/*
Copyright 2023.

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

package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	tutorialv1 "my.domain/tutorial/api/v1"
)

// FooReconciler reconciles a Foo object
type FooReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=tutorial.my.domain,resources=foos,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tutorial.my.domain,resources=foos/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tutorial.my.domain,resources=foos/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Foo object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.15.0/pkg/reconcile
func (r *FooReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("reconsiling foo custom resource")

	var foo tutorialv1.Foo
	if err := r.Get(ctx, req.NamespacedName, &foo); err != nil {
		log.Error(err, "unable to fetch foo")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var podList corev1.PodList
	var friendFound bool
	if err := r.List(ctx, &podList); err != nil {
		log.Error(err, "unable to list pods")
	} else {
		for _, pod := range podList.Items {
			if pod.GetName() == foo.Spec.Name {
				log.Info("found friend pod", "pod", pod.GetName())
				friendFound = true
				break
			}
		}
	}

	foo.Status.Happy = friendFound
	if err := r.Status().Update(ctx, &foo); err != nil {
		log.Error(err, "unable to update foo status")
		return ctrl.Result{}, err
	}
	log.Info("updated foo status", "foo", foo.GetName(), "happy", foo.Status.Happy)
	log.Info("reconcile complete")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *FooReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tutorialv1.Foo{}).
		Watches(
			&source.Kind{Type: &corev1.Pod{}},
			handler.EnqueueRequestsFromMapFunc(r.mapPodsReqToFooReqs),
		).
		Complete(r)
}

func (r *FooReconciler) mapPodsReqToFooReqs(ctx context.Context, obj client.Object) []reconcile.Request {
	log := log.FromContext(ctx)

	req := []reconcile.Request{}
	var list tutorialv1.FooList
	if err := r.Client.List(context.TODO(), &list); err != nil {
		log.Error(err, "unable to list foos custom resource")
	} else {
		for _, item := range list.Items {
			req = append(req, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: item.Name, Namespace: item.Namespace}})
			log.Info("pod linked to a foo custom resource issued an event", "name", obj.GetName())

		}
	}
	return req
}
