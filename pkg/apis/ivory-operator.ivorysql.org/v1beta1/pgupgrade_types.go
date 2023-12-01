// Copyright 2021 - 2023 Crunchy Data Solutions, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1beta1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// IVYUpgradeSpec defines the desired state of IvyUpgrade
type IVYUpgradeSpec struct {

	// +optional
	Metadata *Metadata `json:"metadata,omitempty"`

	// The name of the cluster to be updated
	// +required
	// +kubebuilder:validation:MinLength=1
	IvoryClusterName string `json:"ivoryclusterName"`

	// The image name to use for major IvorySQL upgrades.
	// +optional
	Image *string `json:"image,omitempty"`

	// ImagePullPolicy is used to determine when Kubernetes will attempt to
	// pull (download) container images.
	// More info: https://kubernetes.io/docs/concepts/containers/images/#image-pull-policy
	// +kubebuilder:validation:Enum={Always,Never,IfNotPresent}
	// +optional
	ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy,omitempty"`

	// TODO(benjaminjb) Check the behavior: does updating ImagePullSecrets cause
	// all running IvyUpgrade pods to restart?

	// The image pull secrets used to pull from a private registry.
	// Changing this value causes all running IvyUpgrade pods to restart.
	// https://k8s.io/docs/tasks/configure-pod-container/pull-image-private-registry/
	// +optional
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty"`

	// TODO(benjaminjb): define webhook validation to make sure
	// `fromIvoryVersion` is below `toIvoryVersion`
	// or leverage other validation rules, such as the Common Expression Language
	// rules currently in alpha as of Kubernetes 1.23
	// - https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/#validation-rules

	// The major version of IvorySQL before the upgrade.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=10
	// +kubebuilder:validation:Maximum=16
	FromIvoryVersion int `json:"fromIvoryVersion"`

	// TODO(benjaminjb): define webhook validation to make sure
	// `fromIvoryVersion` is below `toIvoryVersion`
	// or leverage other validation rules, such as the Common Expression Language
	// rules currently in alpha as of Kubernetes 1.23

	// The major version of IvorySQL to be upgraded to.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=10
	// +kubebuilder:validation:Maximum=16
	ToIvoryVersion int `json:"toIvoryVersion"`

	// The image name to use for IvorySQL containers after upgrade.
	// When omitted, the value comes from an operator environment variable.
	// +optional
	ToIvoryImage string `json:"toIvoryImage,omitempty"`

	// Resource requirements for the IvyUpgrade container.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Scheduling constraints of the IvyUpgrade pod.
	// More info: https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// TODO(benjaminjb) Check the behavior: does updating PriorityClassName cause
	// IvyUpgrade to restart?

	// Priority class name for the IvyUpgrade pod. Changing this
	// value causes IvyUpgrade pod to restart.
	// More info: https://kubernetes.io/docs/concepts/scheduling-eviction/pod-priority-preemption/
	// +optional
	PriorityClassName *string `json:"priorityClassName,omitempty"`

	// Tolerations of the IvyUpgrade pod.
	// More info: https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// IvyUpgradeStatus defines the observed state of IvyUpgrade
type IvyUpgradeStatus struct {
	// conditions represent the observations of IvyUpgrade's current state.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// observedGeneration represents the .metadata.generation on which the status was based.
	// +optional
	// +kubebuilder:validation:Minimum=0
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// IvyUpgrade is the Schema for the ivyupgrades API
type IvyUpgrade struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IVYUpgradeSpec   `json:"spec,omitempty"`
	Status IvyUpgradeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// IvyUpgradeList contains a list of IvyUpgrade
type IvyUpgradeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IvyUpgrade `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IvyUpgrade{}, &IvyUpgradeList{})
}
