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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GogatekeeperSpec defines the desired state of Gogatekeeper
type GogatekeeperSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// OIDC connection URL
	OIDCURL string `json:"oidcurl"`
}

// GogatekeeperStatus defines the observed state of Gogatekeeper
type GogatekeeperStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Gogatekeeper is the Schema for the gogatekeepers API
type Gogatekeeper struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GogatekeeperSpec   `json:"spec,omitempty"`
	Status GogatekeeperStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GogatekeeperList contains a list of Gogatekeeper
type GogatekeeperList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Gogatekeeper `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Gogatekeeper{}, &GogatekeeperList{})
}
