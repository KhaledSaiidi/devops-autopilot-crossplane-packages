// Package v1beta1 contains the input type for this Function
// +kubebuilder:object:generate=true
// +groupName=functions.devops-autopilot-crossplane-packages.io
// +versionName=v1beta1
package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// This isn't a custom resource, in the sense that we never install its CRD.
// It is a KRM-like object, so we generate a CRD to describe its schema.

// Input can be used to provide input to this Function.
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:resource:categories=crossplane
type NameNormalizerInput struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Provider        string `json:"provider"`
	SourceFieldPath string `json:"sourceFieldPath"`
	TargetFieldPath string `json:"targetFieldPath"`
	// MaxLength is the maximum allowed length for the normalized name.
	// +kubebuilder:validation:Minimum=1
	MaxLength int `json:"maxLength"`
}
