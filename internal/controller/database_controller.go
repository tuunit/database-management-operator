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
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	k8sv1 "github.com/tuunit/external-database-operator/api/v1"
	k8sv1alpha1 "github.com/tuunit/external-database-operator/api/v1alpha1"
	"github.com/tuunit/external-database-operator/internal/provider"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

// DatabaseReconciler reconciles a Database object
type DatabaseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=k8s.tuunit.com,resources=databases,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=k8s.tuunit.com,resources=databases/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=k8s.tuunit.com,resources=databases/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Database object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.0/pkg/reconcile
func (r *DatabaseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	finalizer := "k8s.tuunit.com/finalizer"

	database := &k8sv1alpha1.Database{}
	if err := r.Get(ctx, req.NamespacedName, database); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if database.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(database, finalizer) {
			controllerutil.AddFinalizer(database, finalizer)
			if err := r.Update(ctx, database); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		if controllerutil.ContainsFinalizer(database, finalizer) {
			// do stuff before deletion of database

			controllerutil.RemoveFinalizer(database, finalizer)
			if err := r.Update(ctx, database); err != nil {
				return ctrl.Result{}, err
			}
		}

		return ctrl.Result{}, nil
	}

	spec := database.Spec

	if spec.DatabaseHostRef == "" {
		log.Info("DatabaseHostRef is not set")
		database.Status.CreationStatus = "DatabaseHostRef is not set"
		if err := r.Status().Update(ctx, database); err != nil {
			log.Error(err, "unable to update Database status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	databaseHost := &k8sv1.DatabaseHost{}
	if err := r.Get(ctx, client.ObjectKey{Namespace: database.Namespace, Name: spec.DatabaseHostRef}, databaseHost); err != nil {
		log.Error(err, "unable to fetch DatabaseHost")

		database.Status.CreationStatus = fmt.Sprintf("DatabaseHost '%s' not found", spec.DatabaseHostRef)
		if err := r.Status().Update(ctx, database); err != nil {
			log.Error(err, "unable to update Database status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var err error

	switch databaseHost.Spec.Type {
	case k8sv1.MySQL:
		log.Info("MySQL database host")
	case k8sv1.Postgres:
		log.Info("Postgres database host")

		client := provider.NewPostgresClient(databaseHost.Spec)
		err = client.CreateDB(&spec)
	}

	if err != nil {
		database.Status.CreationStatus = err.Error()

		if err := r.Status().Update(ctx, database); err != nil {
			log.Error(err, "unable to update database status")
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	database.Status.CreationStatus = fmt.Sprintf("Database '%s' successfully created.", spec.Name)
	database.Status.CreationTime = metav1.Now()

	if err := r.Status().Update(ctx, database); err != nil {
		log.Error(err, "unable to update database status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *DatabaseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&k8sv1alpha1.Database{}).
		Complete(r)
}
