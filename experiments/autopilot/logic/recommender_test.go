package logic

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/turtacn/cloud-prophet/recommender/model"
)

func TestMinResourcesApplied(t *testing.T) {
	constEstimator := NewConstEstimator(model.Resources{
		model.ResourceCPU:    model.CPUAmountFromCores(0.001),
		model.ResourceMemory: model.MemoryAmountFromBytes(1e6),
	})
	recommender := podResourceRecommender{
		constEstimator,
		constEstimator,
		constEstimator}

	containerNameToAggregateStateMap := model.ContainerNameToAggregateStateMap{
		"container-1": &model.AggregateContainerState{},
	}

	recommendedResources := recommender.GetRecommendedPodResources(containerNameToAggregateStateMap)
	assert.Equal(t, model.CPUAmountFromCores(*podMinCPUMillicores/1000), recommendedResources["container-1"].Target[model.ResourceCPU])
	assert.Equal(t, model.MemoryAmountFromBytes(*podMinMemoryMb*1024*1024), recommendedResources["container-1"].Target[model.ResourceMemory])
}

func TestMinResourcesSplitAcrossContainers(t *testing.T) {
	constEstimator := NewConstEstimator(model.Resources{
		model.ResourceCPU:    model.CPUAmountFromCores(0.001),
		model.ResourceMemory: model.MemoryAmountFromBytes(1e6),
	})
	recommender := podResourceRecommender{
		constEstimator,
		constEstimator,
		constEstimator}

	containerNameToAggregateStateMap := model.ContainerNameToAggregateStateMap{
		"container-1": &model.AggregateContainerState{},
		"container-2": &model.AggregateContainerState{},
	}

	recommendedResources := recommender.GetRecommendedPodResources(containerNameToAggregateStateMap)
	assert.Equal(t, model.CPUAmountFromCores((*podMinCPUMillicores/1000)/2), recommendedResources["container-1"].Target[model.ResourceCPU])
	assert.Equal(t, model.CPUAmountFromCores((*podMinCPUMillicores/1000)/2), recommendedResources["container-2"].Target[model.ResourceCPU])
	assert.Equal(t, model.MemoryAmountFromBytes((*podMinMemoryMb*1024*1024)/2), recommendedResources["container-1"].Target[model.ResourceMemory])
	assert.Equal(t, model.MemoryAmountFromBytes((*podMinMemoryMb*1024*1024)/2), recommendedResources["container-2"].Target[model.ResourceMemory])
}

func TestControlledResourcesFiltered(t *testing.T) {
	constEstimator := NewConstEstimator(model.Resources{
		model.ResourceCPU:    model.CPUAmountFromCores(0.001),
		model.ResourceMemory: model.MemoryAmountFromBytes(1e6),
	})
	recommender := podResourceRecommender{
		constEstimator,
		constEstimator,
		constEstimator}

	containerName := "container-1"
	containerNameToAggregateStateMap := model.ContainerNameToAggregateStateMap{
		containerName: &model.AggregateContainerState{
			ControlledResources: &[]model.ResourceName{model.ResourceMemory},
		},
	}

	recommendedResources := recommender.GetRecommendedPodResources(containerNameToAggregateStateMap)
	assert.Contains(t, recommendedResources[containerName].Target, model.ResourceMemory)
	assert.Contains(t, recommendedResources[containerName].LowerBound, model.ResourceMemory)
	assert.Contains(t, recommendedResources[containerName].UpperBound, model.ResourceMemory)
	assert.NotContains(t, recommendedResources[containerName].Target, model.ResourceCPU)
	assert.NotContains(t, recommendedResources[containerName].LowerBound, model.ResourceCPU)
	assert.NotContains(t, recommendedResources[containerName].UpperBound, model.ResourceCPU)
}

func TestControlledResourcesFilteredDefault(t *testing.T) {
	constEstimator := NewConstEstimator(model.Resources{
		model.ResourceCPU:    model.CPUAmountFromCores(0.001),
		model.ResourceMemory: model.MemoryAmountFromBytes(1e6),
	})
	recommender := podResourceRecommender{
		constEstimator,
		constEstimator,
		constEstimator}

	containerName := "container-1"
	containerNameToAggregateStateMap := model.ContainerNameToAggregateStateMap{
		containerName: &model.AggregateContainerState{
			ControlledResources: &[]model.ResourceName{model.ResourceMemory, model.ResourceCPU},
		},
	}

	recommendedResources := recommender.GetRecommendedPodResources(containerNameToAggregateStateMap)
	assert.Contains(t, recommendedResources[containerName].Target, model.ResourceMemory)
	assert.Contains(t, recommendedResources[containerName].LowerBound, model.ResourceMemory)
	assert.Contains(t, recommendedResources[containerName].UpperBound, model.ResourceMemory)
	assert.Contains(t, recommendedResources[containerName].Target, model.ResourceCPU)
	assert.Contains(t, recommendedResources[containerName].LowerBound, model.ResourceCPU)
	assert.Contains(t, recommendedResources[containerName].UpperBound, model.ResourceCPU)
}
