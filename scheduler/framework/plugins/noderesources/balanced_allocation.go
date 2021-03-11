//
package noderesources

import (
	"context"
	"fmt"
	"math"

	framework "github.com/turtacn/cloud-prophet/scheduler/framework/base"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	"k8s.io/apimachinery/pkg/runtime"
)

type BalancedAllocation struct {
	handle framework.FrameworkHandle
	resourceAllocationScorer
}

var _ = framework.ScorePlugin(&BalancedAllocation{})

const BalancedAllocationName = "NodeResourcesBalancedAllocation"

func (ba *BalancedAllocation) Name() string {
	return BalancedAllocationName
}

func (ba *BalancedAllocation) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	nodeInfo, err := ba.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	if err != nil {
		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("getting node %q from Snapshot: %v", nodeName, err))
	}

	return ba.score(pod, nodeInfo)
}

func (ba *BalancedAllocation) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

func NewBalancedAllocation(_ runtime.Object, h framework.FrameworkHandle) (framework.Plugin, error) {
	return &BalancedAllocation{
		handle: h,
		resourceAllocationScorer: resourceAllocationScorer{
			BalancedAllocationName,
			balancedResourceScorer,
			defaultRequestedRatioResources,
			false,
		},
	}, nil
}

// 资源请求与可用的倾斜度，倾斜度越小优先
func balancedResourceScorer(requested, allocable resourceToValueMap) int64 {
	cpuFraction := fractionOfCapacity(requested[v1.ResourceCPU], allocable[v1.ResourceCPU])
	memoryFraction := fractionOfCapacity(requested[v1.ResourceMemory], allocable[v1.ResourceMemory])
	if cpuFraction >= 1 || memoryFraction >= 1 {
		return 0
	}

	diff := math.Abs(cpuFraction - memoryFraction)
	return int64((1 - diff) * float64(framework.MaxNodeScore))
}

func fractionOfCapacity(requested, capacity int64) float64 {
	if capacity == 0 {
		return 1
	}
	return float64(requested) / float64(capacity)
}
