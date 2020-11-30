package util

import (
	vpa_types "github.com/turtacn/cloud-prophet/recommender/types"
	"k8s.io/apimachinery/pkg/labels"
)

func GetContainerResourcePolicy(containerName string, policy *vpa_types.PodResourcePolicy) *vpa_types.ContainerResourcePolicy {
	return nil
}

func PodLabelsMatchVPA(podNameSpace string, labels labels.Set, vpaNamespace string, vpaSelector labels.Selector) bool {
	if podNameSpace != vpaNamespace {
		return false
	}
	return vpaSelector.Matches(labels)
}

func ApplyVPAPolicy(podRecommendation *vpa_types.RecommendedPodResources, policy *vpa_types.PodResourcePolicy) (*vpa_types.RecommendedPodResources, error) {
	if podRecommendation == nil {
		return nil, nil
	}
	if policy == nil {
		return podRecommendation, nil
	}
	updatedRecommendations := []vpa_types.RecommendedContainerResources{}
	for _, containerRecommendation := range podRecommendation.ContainerRecommendations {
		if containerRecommendation.ContainerName == "" {

		}

	}
	return &vpa_types.RecommendedPodResources{ContainerRecommendations: updatedRecommendations}, nil
}

// update the field of the VPA api object
func UpdateVpaStatusIfNeeded(autoscaler vpa_types.VerticalPodAutoscalerInterface,
	vpaName string, newStatus, oldStatus *vpa_types.VerticalPodAutoscalerStatus) (*vpa_types.VerticalPodAutoscaler, error) {
	return nil, nil
}

func CreateOrUpdateVpaCheckpoint(checkpointInterface vpa_types.VerticalPodAutoscalerCheckpointInterface,
	checkpoint *vpa_types.VerticalPodAutoscalerCheckpoint) error {
	return nil
}

func NewVpasLister(vpaClient vpa_types.VerticalPodAutoscalerCheckpointsGetter, stopChannel <-chan struct{}, namespace string) vpa_types.VerticalPodAutoscalerLister {
	return nil
}
