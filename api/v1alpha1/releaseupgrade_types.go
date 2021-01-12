/*
Copyright 2021 Luis Rascao.

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

// ReleaseUpgradeSpec defines the desired state of ReleaseUpgrade
type ReleaseUpgradeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ImageSpec      ReleaseUpgradeImageSpec      `json:"relup,omitempty"`
	VolumeSpec     ReleaseUpgradeVolumeSpec     `json:"volume,omitempty"`
	DeploymentSpec ReleaseUpgradeDeploymentSpec `json:"deployment,omitempty"`
}

type ReleaseUpgradeImageSpec struct {
	Name          string `json:"name,omitempty"`
	Image         string `json:"image,omitempty"`
	Tarball       string `json:"tarball,omitempty"`
	SourceVersion string `json:"sourceVersion,omitempty"`
	TargetVersion string `json:"targetVersion,omitempty"`
}

type ReleaseUpgradeDeploymentSpec struct {
	Name string `json:"name,omitempty"`
}

type ReleaseUpgradeVolumeSpec struct {
	HostPath string `json:"hostPath,omitempty"`
}

// ReleaseUpgradeStatus defines the observed state of ReleaseUpgrade
type ReleaseUpgradeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ReleaseUpgrade is the Schema for the releaseupgrades API
type ReleaseUpgrade struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ReleaseUpgradeSpec   `json:"spec,omitempty"`
	Status ReleaseUpgradeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ReleaseUpgradeList contains a list of ReleaseUpgrade
type ReleaseUpgradeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ReleaseUpgrade `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ReleaseUpgrade{}, &ReleaseUpgradeList{})
}
