/*
Copyright 2025.

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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// UpdateStrategyType defines the type of update strategy
// +kubebuilder:validation:Enum=RollingUpdate;Recreate
type UpdateStrategyType string

const (
	// RollingUpdateStrategyType indicates that pods are replaced one by one
	RollingUpdateStrategyType UpdateStrategyType = "RollingUpdate"
	// RecreateStrategyType indicates that all pods are deleted first, then new ones are created
	RecreateStrategyType UpdateStrategyType = "Recreate"
)

// UpdateStrategy defines the update strategy configuration
type UpdateStrategy struct {
	// Type specifies the update strategy type
	// +kubebuilder:default=RollingUpdate
	Type UpdateStrategyType `json:"type,omitempty"`

	// MaxUnavailable specifies the max number of unavailable pods during update
	// +kubebuilder:default="25%"
	// +optional
	MaxUnavailable *string `json:"maxUnavailable,omitempty"`
}

// Container defines container configuration
type Container struct {
	// Image specifies the container image to use
	// +kubebuilder:validation:MinLength=1
	Image string `json:"image"`
}

// MiniCloneSetSpec defines the desired state of MiniCloneSet
type MiniCloneSetSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// Replicas specifies the number of desired replicas
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:default=1
	Replicas int `json:"replicas"`

	// Container specifies the container configuration
	Container Container `json:"container"`

	// UpdateStrategy specifies the strategy to use when updating pods
	// +optional
	UpdateStrategy UpdateStrategy `json:"updateStrategy,omitempty"`
}

// MiniCloneSetStatus defines the observed state of MiniCloneSet.
type MiniCloneSetStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// AvailableReplicas indicates the number of available replicas
	AvailableReplicas int `json:"availableReplicas"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// MiniCloneSet is the Schema for the miniclonesets API
type MiniCloneSet struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty,omitzero"`

	// spec defines the desired state of MiniCloneSet
	// +required
	Spec MiniCloneSetSpec `json:"spec"`

	// status defines the observed state of MiniCloneSet
	// +optional
	Status MiniCloneSetStatus `json:"status,omitempty,omitzero"`
}

// +kubebuilder:object:root=true

// MiniCloneSetList contains a list of MiniCloneSet
type MiniCloneSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MiniCloneSet `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MiniCloneSet{}, &MiniCloneSetList{})
}

// Hub marks this version as a conversion hub.
func (*MiniCloneSet) Hub() {}
