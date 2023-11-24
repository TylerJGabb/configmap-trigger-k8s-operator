/*
Copyright 2023.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ConfigmapTriggerSpec defines the desired state of ConfigmapTrigger
type ConfigmapTriggerSpec struct {
	ConfigmapName  string `json:"configmapName"`
	DeploymentName string `json:"deploymentName"`
}

// ConfigmapTriggerStatus defines the observed state of ConfigmapTrigger
type ConfigmapTriggerStatus struct {
	LastTriggered *metav1.Time `json:"lastTriggered"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ConfigmapTrigger is the Schema for the configmaptriggers API
type ConfigmapTrigger struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConfigmapTriggerSpec   `json:"spec,omitempty"`
	Status ConfigmapTriggerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ConfigmapTriggerList contains a list of ConfigmapTrigger
type ConfigmapTriggerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConfigmapTrigger `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ConfigmapTrigger{}, &ConfigmapTriggerList{})
}
