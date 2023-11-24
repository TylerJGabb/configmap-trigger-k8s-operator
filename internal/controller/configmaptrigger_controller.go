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

	triggerv1 "tylerjgabb/configmap-trigger-k8s-operator/api/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	CONFIGMAP_FIELD = ".spec.configmapName"
)

// ConfigmapTriggerReconciler reconciles a ConfigmapTrigger object
type ConfigmapTriggerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//https://book.kubebuilder.io/reference/markers/rbac

//+kubebuilder:rbac:groups=trigger.tg.io,resources=configmaptriggers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=trigger.tg.io,resources=configmaptriggers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=trigger.tg.io,resources=configmaptriggers/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ConfigmapTrigger object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *ConfigmapTriggerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	var configmapTrigger triggerv1.ConfigmapTrigger
	if err := r.Get(ctx, req.NamespacedName, &configmapTrigger); err != nil {
		log.Error(err, "unable to fetch ConfigmapTrigger")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// TODO(user): your logic here
	var configmapVersion string
	if configmapTrigger.Spec.ConfigmapName != "" {
		configmapName := configmapTrigger.Spec.ConfigmapName
		foundConfigmap := &corev1.ConfigMap{}
		if err := r.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: configmapName}, foundConfigmap); err != nil {
			log.Error(err, "unable to fetch Configmap")
			return ctrl.Result{}, err
		}
		configmapVersion = foundConfigmap.ResourceVersion
	}
	log.Info("Configmap Found", "configmapVersion", configmapVersion)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConfigmapTriggerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	/*
		The `configMap` field must be indexed by the manager, so that we will be able to lookup `ConfigDeployments` by a referenced `ConfigMap` name.
		This will allow for quickly answer the question:
		- If ConfigMap _x_ is updated, which ConfigDeployments are affected?
	*/
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &triggerv1.ConfigmapTrigger{}, CONFIGMAP_FIELD,
		func(rawObj client.Object) []string {
			configmapTrigger := rawObj.(*triggerv1.ConfigmapTrigger)
			if configmapTrigger.Spec.ConfigmapName == "" {
				return nil
			}
			return []string{configmapTrigger.Spec.ConfigmapName}
		},
	); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&triggerv1.ConfigmapTrigger{}).
		Watches(
			&corev1.ConfigMap{},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForConfigMap),
			builder.WithPredicates(predicate.ResourceVersionChangedPredicate{}),
		).
		Complete(r)
}

func (r *ConfigmapTriggerReconciler) findObjectsForConfigMap(ctx context.Context, obj client.Object) []reconcile.Request {
	// Find all ConfigmapTriggers that reference this configmap
	// For each ConfigmapTrigger, return a reconcile.Request for the Deployment it references
	log := log.FromContext(ctx)
	associatedTriggers := &triggerv1.ConfigmapTriggerList{}
	listOps := &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector(CONFIGMAP_FIELD, obj.GetName()),
		Namespace:     obj.GetNamespace(),
	}
	err := r.List(ctx, associatedTriggers, listOps)
	if err != nil {
		log.Error(err, "unable to list ConfigmapTriggers for ConfigMap", "configmap", obj.GetName())
		return []reconcile.Request{}
	}
	requests := make([]reconcile.Request, len(associatedTriggers.Items))
	for i, trigger := range associatedTriggers.Items {
		log.Info("found ConfigmapTrigger for ConfigMap", "configmap", obj.GetName(), "trigger", trigger.GetName())
		requests[i] = reconcile.Request{
			NamespacedName: client.ObjectKey{
				Name:      trigger.Spec.DeploymentName,
				Namespace: trigger.GetNamespace(),
			},
		}
	}
	return requests
}
