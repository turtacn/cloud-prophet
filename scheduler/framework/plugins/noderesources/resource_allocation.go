package noderesources

import (
	framework "github.com/turtacn/cloud-prophet/scheduler/framework/v1alpha1"
	schedutil "github.com/turtacn/cloud-prophet/scheduler/util"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

// resourceToWeightMap contains resource name and weight.
type resourceToWeightMap map[v1.ResourceName]int64

// defaultRequestedRatioResources is used to set default requestToWeight map for CPU and memory
var defaultRequestedRatioResources = resourceToWeightMap{v1.ResourceMemory: 1, v1.ResourceCPU: 1}

// resourceAllocationScorer contains information to calculate resource allocation score.
type resourceAllocationScorer struct {
	Name                string
	scorer              func(requested, allocable resourceToValueMap, includeVolumes bool, requestedVolumes int, allocatableVolumes int) int64
	resourceToWeightMap resourceToWeightMap
}

// resourceToValueMap contains resource name and score.
type resourceToValueMap map[v1.ResourceName]int64

// score will use `scorer` function to calculate the score.
func (r *resourceAllocationScorer) score(
	pod *v1.Pod,
	nodeInfo *framework.NodeInfo) (int64, *framework.Status) {
	node := nodeInfo.Node()
	if node == nil {
		return 0, framework.NewStatus(framework.Error, "node not found")
	}
	if r.resourceToWeightMap == nil {
		return 0, framework.NewStatus(framework.Error, "resources not found")
	}
	requested := make(resourceToValueMap, len(r.resourceToWeightMap))
	allocatable := make(resourceToValueMap, len(r.resourceToWeightMap))
	for resource := range r.resourceToWeightMap {
		allocatable[resource], requested[resource] = calculateResourceAllocatableRequest(nodeInfo, pod, resource)
	}
	var score int64

	// Check if the pod has volumes and this could be added to scorer function for balanced resource allocation.
	if len(pod.Spec.Volumes) >= 0 && nodeInfo.TransientInfo != nil {
		score = r.scorer(requested, allocatable, true, nodeInfo.TransientInfo.TransNodeInfo.RequestedVolumes, nodeInfo.TransientInfo.TransNodeInfo.AllocatableVolumesCount)
	} else {
		score = r.scorer(requested, allocatable, false, 0, 0)
	}
	if klog.V(10).Enabled() {
		if len(pod.Spec.Volumes) >= 0 && nodeInfo.TransientInfo != nil {
			klog.Infof(
				"%v -> %v: %v, map of allocatable resources %v, map of requested resources %v , allocatable volumes %d, requested volumes %d, score %d",
				pod.Name, node.Name, r.Name,
				allocatable, requested, nodeInfo.TransientInfo.TransNodeInfo.AllocatableVolumesCount,
				nodeInfo.TransientInfo.TransNodeInfo.RequestedVolumes,
				score,
			)
		} else {
			klog.Infof(
				"%v -> %v: %v, map of allocatable resources %v, map of requested resources %v ,score %d,",
				pod.Name, node.Name, r.Name,
				allocatable, requested, score,
			)

		}
	}

	return score, nil
}

// calculateResourceAllocatableRequest returns resources Allocatable and Requested values
func calculateResourceAllocatableRequest(nodeInfo *framework.NodeInfo, pod *v1.Pod, resource v1.ResourceName) (int64, int64) {
	podRequest := calculatePodResourceRequest(pod, resource)
	switch resource {
	case v1.ResourceCPU:
		return nodeInfo.Allocatable.MilliCPU, (nodeInfo.NonZeroRequested.MilliCPU + podRequest)
	case v1.ResourceMemory:
		return nodeInfo.Allocatable.Memory, (nodeInfo.NonZeroRequested.Memory + podRequest)

	case v1.ResourceEphemeralStorage:
		return nodeInfo.Allocatable.EphemeralStorage, (nodeInfo.Requested.EphemeralStorage + podRequest)
	default:
		// 默认支持超卖
		return nodeInfo.Allocatable.ScalarResources[resource], (nodeInfo.Requested.ScalarResources[resource] + podRequest)
	}
	if klog.V(10).Enabled() {
		klog.Infof("requested resource %v not considered for node score calculation",
			resource,
		)
	}
	return 0, 0
}

// calculatePodResourceRequest returns the total non-zero requests. If Overhead is defined for the pod and the
// PodOverhead feature is enabled, the Overhead is added to the result.
// podResourceRequest = max(sum(podSpec.Containers), podSpec.InitContainers) + overHead
func calculatePodResourceRequest(pod *v1.Pod, resource v1.ResourceName) int64 {
	var podRequest int64
	for i := range pod.Spec.Containers {
		container := &pod.Spec.Containers[i]
		value := schedutil.GetNonzeroRequestForResource(resource, &container.Resources.Requests)
		podRequest += value
	}

	for i := range pod.Spec.InitContainers {
		initContainer := &pod.Spec.InitContainers[i]
		value := schedutil.GetNonzeroRequestForResource(resource, &initContainer.Resources.Requests)
		if podRequest < value {
			podRequest = value
		}
	}

	// If Overhead is being utilized, add to the total requests for the pod
	if pod.Spec.Overhead != nil {
		if quantity, found := pod.Spec.Overhead[resource]; found {
			podRequest += quantity.Value()
		}
	}

	return podRequest
}
