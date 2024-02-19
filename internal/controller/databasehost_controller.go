/*
Copyright 2024.

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
	"database/sql"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	k8sv1 "github.com/tuunit/external-database-operator/api/v1"
)

// DatabaseHostReconciler reconciles a DatabaseHost object
type DatabaseHostReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=k8s.tuunit.com,resources=databasehosts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=k8s.tuunit.com,resources=databasehosts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=k8s.tuunit.com,resources=databasehosts/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the DatabaseHost object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *DatabaseHostReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	databaseHost := &k8sv1.DatabaseHost{}
	if err := r.Get(ctx, req.NamespacedName, databaseHost); err != nil {
		log.Error(err, "unable to fetch DatabaseHost")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	spec := databaseHost.Spec

	switch databaseHost.Spec.Type {
	case k8sv1.MySQL:
		log.Info("MySQL database host")
		// Todo: Implement MySQL connection
	case k8sv1.Postgres:
		log.Info("Postgres database host")
		connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s database=postgres sslmode=disable", spec.Host, spec.Port, spec.Superuser, spec.Password)

		db, err := sql.Open("postgres", connectionString)
		if err != nil {
			log.Error(err, "unable to connect to host")
			databaseHost.Status.ConnectionStatus = fmt.Sprintf("Failed to connect to '%s@%s': %s", spec.Superuser, spec.Host, err.Error())
			if err := r.Status().Update(ctx, databaseHost); err != nil {
				log.Error(err, "unable to update DatabaseHost status")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, err
		}

		if err := db.Ping(); err != nil {
			log.Error(err, "unable to ping host")
			databaseHost.Status.ConnectionStatus = fmt.Sprintf("Failed to ping '%s@%s': %s", spec.Superuser, spec.Host, err.Error())
			if err := r.Status().Update(ctx, databaseHost); err != nil {
				log.Error(err, "unable to update DatabaseHost status")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, err
		}
	default:
		databaseHost.Status.ConnectionStatus = fmt.Sprintf("Database type '%s' not supported", spec.Type)
		if err := r.Status().Update(ctx, databaseHost); err != nil {
			log.Error(err, "unable to update DatabaseHost status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, fmt.Errorf("database type '%s' not supported", spec.Type)
	}

	databaseHost.Status.ConnectionStatus = fmt.Sprintf("Connection with host '%s' was successful", spec.Host)
	databaseHost.Status.LastConnectionTime = metav1.Now()

	if err := r.Status().Update(ctx, databaseHost); err != nil {
		log.Error(err, "unable to update DatabaseHost status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DatabaseHostReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&k8sv1.DatabaseHost{}).
		Complete(r)
}
