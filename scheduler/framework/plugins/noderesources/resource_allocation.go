package noderesources

import (
	framework "github.com/turtacn/cloud-prophet/scheduler/framework/base"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	schedutil "github.com/turtacn/cloud-prophet/scheduler/util"
	"k8s.io/klog/v2"
)

type resourceToWeightMap map[v1.ResourceName]int64

var defaultRequestedRatioResources = resourceToWeightMap{v1.ResourceMemory: 1, v1.ResourceCPU: 1}

type resourceAllocationScorer struct {
	Name                string
	scorer              func(requested, allocable resourceToValueMap) int64
	resourceToWeightMap resourceToWeightMap
	printHostFlag       bool
}

type resourceToValueMap map[v1.ResourceName]int64

func (r *resourceAllocationScorer) score(
	pod *v1.Pod,
	nodeInfo *framework.NodeInfo) (int64, *framework.Status) {
	node := nodeInfo.Node()
	if node == nil {
		klog.Warningf("not found score node %v", node)
		return 0, framework.NewStatus(framework.Error, "node not found")
	}
	if r.resourceToWeightMap == nil {
		klog.Warningf("not found resource to weightmap, resource allocation score %v", r)
		return 0, framework.NewStatus(framework.Error, "resources not found")
	}
	requested := make(resourceToValueMap, len(r.resourceToWeightMap))
	allocatable := make(resourceToValueMap, len(r.resourceToWeightMap))
	for resource := range r.resourceToWeightMap {
		allocatable[resource], requested[resource] = calculateResourceAllocatableRequest(nodeInfo, pod, resource)
	}
	var score int64

	score = r.scorer(requested, allocatable)

	if r.printHostFlag {
		klog.Infof(
			"%v -> %v: %v, map of allocatable resources %v, map of requested resources %v ,score %d,",
			pod.Name, node.Name, r.Name,
			allocatable, requested, score,
		)
	}

	return score, nil
}

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
		return nodeInfo.Allocatable.ScalarResources[resource], (nodeInfo.Requested.ScalarResources[resource] + podRequest)
	}
	if true {
		klog.Warningf("requested resource %v not considered for node score calculation",
			resource,
		)
	}
	return 0, 0
}

func calculatePodResourceRequest(pod *v1.Pod, resource v1.ResourceName) int64 {
	var podRequest int64
	for i := range pod.Spec.Containers {
		container := &pod.Spec.Containers[i]
		value := schedutil.GetNonzeroRequestForResource(resource, &container.Resources.Requests)
		podRequest += value
	}
	if pod.Spec.Overhead != nil {
		if quantity, found := pod.Spec.Overhead[resource]; found {
			podRequest += quantity.Value()
		}
	}

	return podRequest
}
