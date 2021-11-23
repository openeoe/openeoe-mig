/*
Copyright 2018 The Multicluster-Controller Authors.

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
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MigrationSpec defines the desired state of Migration
type MigrationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	MigrationServiceSources []MigrationServiceSource `json:"migrationServiceSource"`
}

type MigrationServiceSource struct {
	// container service spec
	MigrationSources []MigrationSource `json:"migrationSource"`
	ServiceName      string            `json:"serviceName"`
	TargetCluster    string            `json:"targetCluster"`
	SourceCluster    string            `json:"sourceCluster"`
	NameSpace        string            `json:"nameSpace"`
}
type MigrationSource struct {
	// Migration source
	ResourceType string `json:"resourceType"`
	ResourceName string `json:"resourceName"`
}

// MigrationStatus defines the observed state of Migration
type MigrationStatus struct {
	ElapsedTime string `json:"elapsedTime,omitempty"`

	// Status of the condition, one of True, False, Unknown.
	IsZeroDownTime v1.ConditionStatus `json:"isZeroDownTime,omitempty" protobuf:"bytes,2,opt,name=isZeroDownTime,casttype=k8s.io/api/core/v1.ConditionStatus"`
	// 현재 진행도
	CurrentCount int32 `json:"currentCount,omitempty"`
	// 최대 진행도
	MaxCount int32 `json:"maxCount,omitempty"`
	// 최대 컨디션 카운트. currentCount/maxCount
	ConditionProgress string `json:"progress,omitempty"`

	// Condition 정보 리스트
	//Conditions []MigrationCondition `json:"conditions,omitempty"`
	// Status of the condition, one of True, False, Unknown.
	Status      v1.ConditionStatus `json:"status,omitempty" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/api/core/v1.ConditionStatus"`
	Description string             `json:"description"`
}

/*
type MigrationConditionType string

// These are valid conditions of a deployment.
const (
	//
	MigrationSourceInit MigrationConditionType = "SourceInit"
	//
	MigrationProgressing MigrationConditionType = "Progressing"
	//
	MigrationFailure MigrationConditionType = "Failure"
)

// MigrationCondition describes the state of a migration at a certain point.
type MigrationCondition struct {
	// Type of Migration condition.
	Type MigrationConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=MigrationConditionType"`
	// Status of the condition, one of True, False, Unknown.
	Status v1.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=k8s.io/api/core/v1.ConditionStatus"`
	// 최근 컨디션의 업데이트 타임
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty" protobuf:"bytes,6,opt,name=lastUpdateTime"`
	// 해당 컨디션으로 업데이트 된 타임
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,7,opt,name=lastTransitionTime"`
	// The reason for the condition's last transition.
	Reason string `json:"reason,omitempty" protobuf:"bytes,4,opt,name=reason"`
	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty" protobuf:"bytes,5,opt,name=message"`
}
*/
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Migration is the Schema for the Migrations API
// +kubebuilder:subresource:status
// +k8s:openapi-gen=true
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="Created time stamp"
// +kubebuilder:printcolumn:name="SourceCluster",type="string",JSONPath=".spec.migrationServiceSource[*].sourceCluster",description="-"
// +kubebuilder:printcolumn:name="TargetCluster",type="string",JSONPath=".spec.migrationServiceSource[*].targetCluster",description="-"
// +kubebuilder:printcolumn:name="ServiceName",type="string",JSONPath=".spec.migrationServiceSource[*].serviceName",description="-"
// +kubebuilder:printcolumn:name="NameSpace",type="string",JSONPath=".spec.migrationServiceSource[*].nameSpace",description="-"
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status",description="-"
// +kubebuilder:printcolumn:name="Description",type="string",JSONPath=".status.description",description="-"
// +kubebuilder:printcolumn:name="Progress",type="string",JSONPath=".status.progress",description="-"
// +kubebuilder:printcolumn:name="ElapsedTime",type="string",JSONPath=".status.elapsedTime",description="ElapsedTime"
// +kubebuilder:printcolumn:name="IsZeroDownTime",type="string",JSONPath=".status.isZeroDownTime",description="-"
type Migration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MigrationSpec   `json:"spec,omitempty"`
	Status MigrationStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MigrationList contains a list of Migration
type MigrationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Migration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Migration{}, &MigrationList{})
}
