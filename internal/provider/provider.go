package provider

import (
	"github.com/tuunit/external-database-operator/api/v1alpha1"
)

type DatabaseProvider interface {
	CheckConnection() error
	CreateDB(spec *v1alpha1.DatabaseSpec) error
	CreateUser(spec *v1alpha1.DatabaseUserSpec) error
	CreateRole() error
}
