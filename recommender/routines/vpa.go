package routines

import (
	"github.com/turtacn/cloud-prophet/recommender/model"
	vpa_types "github.com/turtacn/cloud-prophet/recommender/types"
	//vpa_types "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	//api_utils "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/utils/vpa"
	api_utils "github.com/turtacn/cloud-prophet/recommender/util"
)

// GetContainerNameToAggregateStateMap returns ContainerNameToAggregateStateMap for pods.
func GetContainerNameToAggregateStateMap(vpa *model.Vpa) model.ContainerNameToAggregateStateMap {
	containerNameToAggregateStateMap := vpa.AggregateStateByContainerName()
	filteredContainerNameToAggregateStateMap := make(model.ContainerNameToAggregateStateMap)

	for containerName, aggregatedContainerState := range containerNameToAggregateStateMap {
		containerResourcePolicy := api_utils.GetContainerResourcePolicy(containerName, vpa.ResourcePolicy)
		autoscalingDisabled := containerResourcePolicy != nil && containerResourcePolicy.Mode != nil &&
			*containerResourcePolicy.Mode == vpa_types.ContainerScalingModeOff
		if !autoscalingDisabled && aggregatedContainerState.TotalSamplesCount > 0 {
			aggregatedContainerState.UpdateFromPolicy(containerResourcePolicy)
			filteredContainerNameToAggregateStateMap[containerName] = aggregatedContainerState
		}
	}
	return filteredContainerNameToAggregateStateMap
}
