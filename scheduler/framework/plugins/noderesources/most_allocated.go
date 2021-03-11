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

type MostAllocated struct {
	handle framework.FrameworkHandle
	resourceAllocationScorer
}

var _ = framework.ScorePlugin(&MostAllocated{})

const MostAllocatedName = "NodeResourcesMostAllocated"

func (ma *MostAllocated) Name() string {
	return MostAllocatedName
}

func (ma *MostAllocated) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	nodeInfo, err := ma.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	if err != nil || nodeInfo.Node() == nil {
		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("getting node %q from Snapshot: %v, node is nil: %v", nodeName, err, nodeInfo.Node() == nil))
	}

	return ma.score(pod, nodeInfo)
}

func (ma *MostAllocated) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

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
			printHostFlag:       args.PrintHostTrace,
		},
	}, nil
}

func mostResourceScorer(resToWeightMap resourceToWeightMap) func(requested, allocable resourceToValueMap) int64 {
	return func(requested, allocable resourceToValueMap) int64 {
		var nodeScore, weightSum int64
		if len(resToWeightMap) == 0 {
			klog.Warningf("mostResourceScorer find resource to weight map is empty!")
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

func mostRequestedScore(requested, capacity int64) int64 {
	if capacity == 0 {
		return 0
	}
	if requested > capacity {
		return 0
	}

	return (requested * framework.MaxNodeScore) / capacity
}
