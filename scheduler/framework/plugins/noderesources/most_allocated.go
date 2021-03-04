//
//
package noderesources

import (
	"context"
	"fmt"

	"github.com/turtacn/cloud-prophet/scheduler/apis/config"
	"github.com/turtacn/cloud-prophet/scheduler/apis/config/validation"
	framework "github.com/turtacn/cloud-prophet/scheduler/framework/base"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
)

// MostAllocated is a score plugin that favors nodes with high allocation based on requested resources.
type MostAllocated struct {
	handle framework.FrameworkHandle
	resourceAllocationScorer
}

var _ = framework.ScorePlugin(&MostAllocated{})

// MostAllocatedName is the name of the plugin used in the plugin registry and configurations.
const MostAllocatedName = "NodeResourcesMostAllocated"

// Name returns name of the plugin. It is used in logs, etc.
func (ma *MostAllocated) Name() string {
	return MostAllocatedName
}

// Score invoked at the Score extension point.
func (ma *MostAllocated) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	nodeInfo, err := ma.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	if err != nil || nodeInfo.Node() == nil {
		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("getting node %q from Snapshot: %v, node is nil: %v", nodeName, err, nodeInfo.Node() == nil))
	}

	// ma.score favors nodes with most requested resources.
	// It calculates the percentage of memory and CPU requested by pods scheduled on the node, and prioritizes
	// based on the maximum of the average of the fraction of requested to capacity.
	// Details: (cpu(MaxNodeScore * sum(requested) / capacity) + memory(MaxNodeScore * sum(requested) / capacity)) / weightSum
	return ma.score(pod, nodeInfo)
}

// ScoreExtensions of the Score plugin.
func (ma *MostAllocated) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

// NewMostAllocated initializes a new plugin and returns it.
func NewMostAllocated(maArgs runtime.Object, h framework.FrameworkHandle) (framework.Plugin, error) {
	args, ok := maArgs.(*config.NodeResourcesMostAllocatedArgs)
	if !ok {
		return nil, fmt.Errorf("want args to be of type NodeResourcesMostAllocatedArgs, got %T", args)
	}

	if err := validation.ValidateNodeResourcesMostAllocatedArgs(args); err != nil {
		return nil, err
	}

	resToWeightMap := make(resourceToWeightMap)
	for _, resource := range (*args).Resources {
		resToWeightMap[v1.ResourceName(resource.Name)] = resource.Weight
	}

	return &MostAllocated{
		handle: h,
		resourceAllocationScorer: resourceAllocationScorer{
			Name:                MostAllocatedName,
			scorer:              mostResourceScorer(resToWeightMap),
			resourceToWeightMap: resToWeightMap,
		},
	}, nil
}

func mostResourceScorer(resToWeightMap resourceToWeightMap) func(requested, allocable resourceToValueMap, includeVolumes bool, requestedVolumes int, allocatableVolumes int) int64 {
	return func(requested, allocable resourceToValueMap, includeVolumes bool, requestedVolumes int, allocatableVolumes int) int64 {
		var nodeScore, weightSum int64
		klog.Infof("mostResourceScorer find resource to weight map size is %d", len(resToWeightMap))
		if len(resToWeightMap) == 0 {
			return 0
		}
		for resource, weight := range resToWeightMap {
			resourceScore := mostRequestedScore(requested[resource], allocable[resource])
			nodeScore += resourceScore * weight
			weightSum += weight
		}
		return (nodeScore / weightSum)
	}
}

// The used capacity is calculated on a scale of 0-MaxNodeScore (MaxNodeScore is
// constant with value set to 100).
// 0 being the lowest priority and 100 being the highest.
// The more resources are used the higher the score is. This function
// is almost a reversed version of noderesources.leastRequestedScore.
func mostRequestedScore(requested, capacity int64) int64 {
	if capacity == 0 {
		return 0
	}
	if requested > capacity {
		return 0
	}

	return (requested * framework.MaxNodeScore) / capacity
}
