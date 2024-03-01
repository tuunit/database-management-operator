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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	k8sv1 "github.com/tuunit/external-database-operator/api/v1"
	k8sv1alpha1 "github.com/tuunit/external-database-operator/api/v1alpha1"

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

	database := &k8sv1alpha1.Database{}
	if err := r.Get(ctx, req.NamespacedName, database); err != nil {
		log.Error(err, "unable to fetch Database")
		return ctrl.Result{}, client.IgnoreNotFound(err)
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

	switch databaseHost.Spec.Type {
	case k8sv1.MySQL:
		log.Info("MySQL database host")
	case k8sv1.Postgres:
		log.Info("Postgres database host")

		conn := databaseHost.Spec
		connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s database=postgres sslmode=disable", conn.Host, conn.Port, conn.Superuser, conn.Password)

		db, err := sql.Open("postgres", connectionString)
		if err != nil {
			log.Error(err, "unable to connect to host")
			// Todo: Proper error handling for connection failure
			// Introduce a new status condition for connection in the databasehost
			if err := r.Status().Update(ctx, database); err != nil {
				log.Error(err, "unable to update DatabaseHost status")
				return ctrl.Result{}, err
			}
			// Todo: Configure RequeueAfter to retry the connection
			return ctrl.Result{}, err
		}
		defer db.Close()

		owner := conn.Superuser
		charset := "UTF8"
		collation := "en_US.UTF-8"

		if spec.Owner != "" {
			owner = spec.Owner
		}

		if spec.Charset != "" {
			charset = spec.Charset
		}

		if spec.Collation != "" {
			collation = spec.Collation
		}

		var datname string
		row := db.QueryRow(`SELECT datname FROM pg_database WHERE datname = $1`, spec.Name)
		err = row.Scan(&datname)

		if err == sql.ErrNoRows {
			_, err = db.Exec(`CREATE DATABASE "` + spec.Name + `"
												  WITH OWNER "` + owner + `" 
													ENCODING '` + charset + `'
													LC_COLLATE '` + collation + `'
													LC_CTYPE '` + collation + `'`)
		}

		if err != nil && err != sql.ErrNoRows {
			log.Error(err, "unable to create database")
			database.Status.CreationStatus = fmt.Sprintf("Failed to create database '%s': %s", spec.Name, err.Error())
			if err := r.Status().Update(ctx, database); err != nil {
				log.Error(err, "unable to update database status")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, err
		}
	}

	database.Status.CreationStatus = fmt.Sprintf("Database '%s' successfully created.", spec.Name)
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
