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
