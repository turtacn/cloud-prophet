package model

import (
	corev1 "k8s.io/api/core/v1"
)

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
