/*
Copyright 2020 mahuang.

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

package debugtask

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type DebugObjectInfo struct {
	// DebugPodimage is the image of debug Pod.
	DebugPodImage string `json:"debugPodImage"`
	// DebugPodName is pod's name of the tool for debugging.
	DebugPodName string `json:"debugPodName,omitempty"`
}

type TargetObjectInfo struct {
	// TargetPodNamespace is the namespace of the target pod.
	TargetPodNamespace string `json:"targetPodNamespace"`
	// TargetPodName is the target pod to be debugged.
	TargetPodName string `json:"targetPodName"`
	// TargetPodContainerName is the name of the target container in the target pod.
	TargetPodContainerName string `json:"targePodContainerName"`
}

// DebugTaskSpec defines the desired state of DebugTask
type DebugTaskSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// DebugObjectInfo is the information of Debug tools
	DebugObjectInfo *DebugObjectInfo `json:"debugObjectInfo"`
	// TargetObjectInfo is the information of target to be debugged
	TargetObjectInfo *TargetObjectInfo `json:"targetObjectInfo"`
	// StartTime means when did the debug task start.
	StartTime *metav1.Time `json:"startTime,omitempty"`
	// EndTime means when dit the debug task end.
	EndTime *metav1.Time `json:"endTime,omitempty"`
	// History record all commands in this debug task.
	History []string `json:"history,omitempty"`
}

type DebugPhase string

const (
	// DebugPending means the debug request has been accepted by the system,
	// but the debug container not been started.
	DebugPending DebugPhase = "Pending"
	// DebugRuning means the debug pod has been started, client is debuging.
	DebugRuning DebugPhase = "Debuging"
	// DebugSucceeded means the debug task has been finnished normally.
	DebugSucceeded DebugPhase = "Finished"
	// DebugFaild means the debug task failed
	DebugFaild DebugPhase = "Failed"
	// DebugUnknown means that for some reason the state of the DebugTask
	// couldn't work.
	DebugUnknown DebugPhase = "Unknown"
)

// DebugTaskStatus defines the observed state of DebugTask
type DebugTaskStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Phase DebugPhase `json:"phase"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DebugTask is the Schema for the debugtasks API
// +k8s:openapi-gen=true
// +kubebuilder:printcolumn:name="TargetPod",type=string,JSONPath=`.spec.targetObjectInfo.targetPodName`
// +kubebuilder:printcolumn:name="StartTime",type=string,JSONPath=`.spec.startTime`
// +kubebuilder:printcolumn:name="EndTime",type=string,JSONPath=`.spec.endTime`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.spec.status.phase`
type DebugTask struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DebugTaskSpec   `json:"spec,omitempty"`
	Status DebugTaskStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DebugTaskList contains a list of DebugTask
type DebugTaskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DebugTask `json:"items"`
}
