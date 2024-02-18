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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DatabaseType string

const (
	MySQL    DatabaseType = "mysql"
	Postgres DatabaseType = "postgres"
)

// DatabaseHostSpec defines the desired state of DatabaseHost
type DatabaseHostSpec struct {
	// Host is the hostname or IP address of the database host
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required
	Host string `json:"host"`
	// Type is the type of database running on the host
	// +kubebuilder:validation:Enum=postgres;mysql
	// +kubebuilder:validation:Required
	Type DatabaseType `json:"type"`
	// Superuser is the name of the superuser for the database
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required
	Superuser string `json:"superuser"`
	// Password is the password for the superuser
	// +optional
	Password string `json:"password,omitempty"`
	// PasswordSecretRef is a reference to a secret in the same namespace
	// that contains the password for the superuser
	// +optional
	PasswordSecretRef string `json:"passwordSecretRef,omitempty"`
	// Port is the port number for the database
	// +optional
	Port int32 `json:"port"`
}

// DatabaseHostStatus defines the observed state of DatabaseHost
type DatabaseHostStatus struct {
	LastConnectionTime metav1.Time `json:"lastConnectionTime,omitempty"`
	ConnectionStatus   string      `json:"connectionStatus,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DatabaseHost is the Schema for the databasehosts API
type DatabaseHost struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseHostSpec   `json:"spec,omitempty"`
	Status DatabaseHostStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DatabaseHostList contains a list of DatabaseHost
type DatabaseHostList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DatabaseHost `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DatabaseHost{}, &DatabaseHostList{})
}
