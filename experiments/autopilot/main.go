package main

import (
	"fmt"
	"github.com/turtacn/cloud-prophet/recommender/logic"
	"github.com/turtacn/cloud-prophet/recommender/model"
	vpa_types "github.com/turtacn/cloud-prophet/recommender/types"
)

func main() {

	entityAggregateStateMap := make(model.ContainerNameToAggregateStateMap)
	setResourceRecommender := logic.CreatePodResourceRecommender()

	resources := setResourceRecommender.GetRecommendedPodResources(entityAggregateStateMap)

	containerResources := make([]vpa_types.RecommendedContainerResources, 0, len(resources))
	for containerName, res := range resources {
		containerResources = append(containerResources, vpa_types.RecommendedContainerResources{
			ContainerName:  containerName,
			Target:         model.ResourcesAsResourceList(res.Target),
			LowerBound:     model.ResourcesAsResourceList(res.LowerBound),
			UpperBound:     model.ResourcesAsResourceList(res.UpperBound),
			UncappedTarget: model.ResourcesAsResourceList(res.Target),
		})
	}
	recommendation := &vpa_types.RecommendedPodResources{containerResources}

	for _, recon := range recommendation.ContainerRecommendations {
		fmt.Println(recon.ContainerName)
	}

}
