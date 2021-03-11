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

type LeastAllocated struct {
	handle framework.FrameworkHandle
	resourceAllocationScorer
}

var _ = framework.ScorePlugin(&LeastAllocated{})

const LeastAllocatedName = "NodeResourcesLeastAllocated"

func (la *LeastAllocated) Name() string {
	return LeastAllocatedName
}

func (la *LeastAllocated) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	nodeInfo, err := la.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	if err != nil {
		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("getting node %q from Snapshot: %v", nodeName, err))
	}

	return la.score(pod, nodeInfo)
}

func (la *LeastAllocated) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

func NewLeastAllocated(laArgs runtime.Object, h framework.FrameworkHandle) (framework.Plugin, error) {
	args, ok := laArgs.(*config.NodeResourcesLeastAllocatedArgs)
	if !ok {
		return nil, fmt.Errorf("want args to be of type NodeResourcesLeastAllocatedArgs, got %T", laArgs)
	}

	if err := validation.ValidateNodeResourcesLeastAllocatedArgs(args); err != nil {
		return nil, err
	}

	resToWeightMap := make(resourceToWeightMap)
	for _, resource := range (*args).Resources {
		resToWeightMap[v1.ResourceName(resource.Name)] = resource.Weight
	}

	return &LeastAllocated{
		handle: h,
		resourceAllocationScorer: resourceAllocationScorer{
			Name:                LeastAllocatedName,
			scorer:              leastResourceScorer(resToWeightMap),
			resourceToWeightMap: resToWeightMap,
		},
	}, nil
}

// 剩余资源越多则越优先 加入各个维度的权重
func leastResourceScorer(resToWeightMap resourceToWeightMap) func(resourceToValueMap, resourceToValueMap) int64 {
	return func(requested, allocable resourceToValueMap) int64 {
		var nodeScore, weightSum int64
		if len(resToWeightMap) == 0 {
			klog.Warningf("leastResourceScorer find resource to weight map is empty!")
			return 0
		}
		for resource, weight := range resToWeightMap {
			resourceScore := leastRequestedScore(requested[resource], allocable[resource])
			nodeScore += resourceScore * weight
			weightSum += weight
		}
		return nodeScore / weightSum
	}
}

func leastRequestedScore(requested, capacity int64) int64 {
	if capacity == 0 {
		return 0
	}
	if requested > capacity {
		return 0
	}

	return ((capacity - requested) * int64(framework.MaxNodeScore)) / capacity
}
