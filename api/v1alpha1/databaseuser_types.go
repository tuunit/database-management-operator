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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ACL for PostgreSQL
// https://www.postgresql.org/docs/15/ddl-priv.html
// ACL for MySQL
// https://dev.mysql.com/doc/refman/8.3/en/grant.html
type Privilege struct {
	// The type of object for which to grant privileges
	// +kubebuilder:validation:Required
	ObjectType string `json:"objectType"`
	// The list of privileges to grant
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:Required
	Privileges []string `json:"privileges"`
}

type SecretKeySelector struct {
	// The name of the secret in the object's namespace to select from.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required
	Name string `json:"name"`
	// The key of the secret to select from.  Must be a valid secret key.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required
	Key string `json:"key"`
}

// DatabaseUserSpec defines the desired state of DatabaseUser
type DatabaseUserSpec struct {
	// DatabaseRef is a reference to a Database object in the same namespace
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required
	DatabaseRef string `json:"databaseRef"`
	// Username is the name of the user to create
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required
	Username string `json:"username"`
	// Password is the password for the user
	// +kubebuilder:validation:MinLength=1
	// +optional
	Password string `json:"password,omitempty"`
	// PasswordSecretRef is a reference to an existing secret in the same namespace
	// +optional
	PasswordSecretRef *SecretKeySelector `json:"passwordSecretRef,omitempty"`
	// Privileges is a list of privileges to grant to the user
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:Required
	Privileges []Privilege `json:"privileges"`
}

// DatabaseUserStatus defines the observed state of DatabaseUser
type DatabaseUserStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DatabaseUser is the Schema for the databaseusers API
type DatabaseUser struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatabaseUserSpec   `json:"spec,omitempty"`
	Status DatabaseUserStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DatabaseUserList contains a list of DatabaseUser
type DatabaseUserList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DatabaseUser `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DatabaseUser{}, &DatabaseUserList{})
}
