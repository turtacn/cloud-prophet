package types

import (
	apiv1 "k8s.io/api/core/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type VerticalPodAutoscalerConditionType string

var (
	RecommendationProvided VerticalPodAutoscalerConditionType = "RecommendationProvided"
	NoPodsMatched          VerticalPodAutoscalerConditionType = "NoPodsMatched"
	FetchingHistory        VerticalPodAutoscalerConditionType = "FetchingHistory"
)

type VerticalPodAutoscalerCondition struct {
	Type               VerticalPodAutoscalerConditionType `json:"type"`
	Status             apiv1.ConditionStatus              `json:"status"`
	LastTransitionTime metav1.Time                        `json:"last_transition_time"`
	Reason             string                             `json:"reason"`
	Message            string                             `json:"message"`
}

type VerticalPodAutoscalerStatus struct {
	Recommendation *RecommendedPodResources         `json:"recommendation"`
	Conditions     []VerticalPodAutoscalerCondition `json:"conditions"`
}

type VerticalPodAutoscalarCheckpoint struct {
}

type VerticalPodAutoscalarCheckpointList struct {
	Items []VerticalPodAutoscalarCheckpoint `json:"items"`
}

type VerticalPodAutoscalerCheckpointsGetter interface {
	VerticalPodAutoscalarCheckpoints() VerticalPodAutoscalarCheckpointInterface
}

// 创建、更新、删除、列表删除、查询、列表
type VerticalPodAutoscalarCheckpointInterface interface {
	Create()
	Update()
	Delete()
	DeleteCollection()
	Get()
	List()
}

type RecommendedPodResources struct {
	ContainerRecommendations []RecommendedContainerResources `json:"container_recommendations"`
}

type RecommendedContainerResources struct {
	ContainerName  string              `json:"container_name"`
	Target         corev1.ResourceList `json:"target"`
	LowerBound     corev1.ResourceList `json:"lower_bound"`
	UpperBound     corev1.ResourceList `json:"upper_bound"`
	UncappedTarget corev1.ResourceList `json:"uncapped_target"`
}

type PodResourcePolicy struct {
	ContainerPolicies []ContainerResourcePolicy `json:"container_polices"`
}
type ContainerResourcePolicy struct {
	ContainerName string                `json:"continaer_name"`
	Mode          *ContainerScalingMode `json:"mode"`
	MinAllowed    corev1.ResourceList   `json:"min_allowed"`
	MaxAllowed    corev1.ResourceList   `json:"max_allowed"`
}

type PodUpdatePolicy struct {
	UpdateMode *UpdateMode `json:"update_mode"`
}
type UpdateMode string

const (
	UpdateModeOff  UpdateMode = "Off"
	UpdateModeAuto UpdateMode = "Auto"
)

type ContainerScalingMode string

const (
	ContainerScalingModeAuto ContainerScalingMode = "Auto"
	ContainerScalingModeOff  ContainerScalingMode = "Off"
)
