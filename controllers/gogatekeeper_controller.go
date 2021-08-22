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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	gatekeeperv1alpha1 "github.com/theEndBeta/gogatekeeper-operator/api/v1alpha1"
)

// GogatekeeperReconciler reconciles a Gogatekeeper object
type GogatekeeperReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=gatekeeper.adstein,resources=gogatekeepers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gatekeeper.adstein,resources=gogatekeepers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gatekeeper.adstein,resources=gogatekeepers/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;
//+kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete

var gatekeeperDefaultConfig = map[string]string{
	"upstream-url":          "http://127.0.0.1:80",
	"listen":                ":3000",
	"listen-admin":          ":4000",
	"enable-refresh-tokens": "true",
	"secure-cookie":         "false",
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Gogatekeeper object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *GogatekeeperReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	gatekeeper := &gatekeeperv1alpha1.Gogatekeeper{}
	err := r.Get(ctx, req.NamespacedName, gatekeeper)

	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Gogatekeeper resource not found - ignoring")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get Gogatekeeper")
		return ctrl.Result{}, err
	}

	// Check for existing config
	foundConf := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: gatekeeper.Name, Namespace: gatekeeper.Namespace}, foundConf)
	if err != nil {
		// Create new configMap
		if errors.IsNotFound(err) {
			config := r.newGatekeeperConfigMap(gatekeeper)
			log.Info("Creating new gogatekeeper config map", "ConfigMap.Name", config.Name, "ConfigMap.Namespace", config.Namespace)
			err = r.Create(ctx, config)

			if err != nil {
				log.Error(err, "Failed to create new gogatekeeper config map", "ConfigMap.Name", config.Name, "ConfigMap.Namespace", config.Namespace)
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}

		log.Error(err, "Failed to get ConfigMap")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *GogatekeeperReconciler) newGatekeeperConfigMap(gk *gatekeeperv1alpha1.Gogatekeeper) *corev1.ConfigMap {

	tempConfig := map[string]string{
		"discovery-url": gk.Spec.OIDCURL,
	}

	var config string
	for k, v := range gatekeeperDefaultConfig {
		config += fmt.Sprintf("%s: %s\n", k, v)
	}
	for k, v := range tempConfig {
		config += fmt.Sprintf("%s: %s\n", k, v)
	}

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      gk.Name,
			Namespace: gk.Namespace,
		},
		Data: map[string]string{
			"gatekeeper.yaml": config,
		},
	}

	ctrl.SetControllerReference(gk, configMap, r.Scheme)
	return configMap
}

// SetupWithManager sets up the controller with the Manager.
func (r *GogatekeeperReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gatekeeperv1alpha1.Gogatekeeper{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
