//
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

// BalancedAllocation is a score plugin that calculates the difference between the cpu and memory fraction
// of capacity, and prioritizes the host based on how close the two metrics are to each other.
type BalancedAllocation struct {
	handle framework.FrameworkHandle
	resourceAllocationScorer
}

var _ = framework.ScorePlugin(&BalancedAllocation{})

// BalancedAllocationName is the name of the plugin used in the plugin registry and configurations.
const BalancedAllocationName = "NodeResourcesBalancedAllocation"

// Name returns name of the plugin. It is used in logs, etc.
func (ba *BalancedAllocation) Name() string {
	return BalancedAllocationName
}

// Score invoked at the score extension point.
func (ba *BalancedAllocation) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	nodeInfo, err := ba.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	if err != nil {
		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("getting node %q from Snapshot: %v", nodeName, err))
	}

	// ba.score favors nodes with balanced resource usage rate.
	// It calculates the difference between the cpu and memory fraction of capacity,
	// and prioritizes the host based on how close the two metrics are to each other.
	// Detail: score = (1 - variance(cpuFraction,memoryFraction,volumeFraction)) * MaxNodeScore. The algorithm is partly inspired by:
	// "Wei Huang et al. An Energy Efficient Virtual Machine Placement Algorithm with Balanced
	// Resource Utilization"
	return ba.score(pod, nodeInfo)
}

// ScoreExtensions of the Score plugin.
func (ba *BalancedAllocation) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

// NewBalancedAllocation initializes a new plugin and returns it.
func NewBalancedAllocation(_ runtime.Object, h framework.FrameworkHandle) (framework.Plugin, error) {
	return &BalancedAllocation{
		handle: h,
		resourceAllocationScorer: resourceAllocationScorer{
			BalancedAllocationName,
			balancedResourceScorer,
			defaultRequestedRatioResources,
			true,
		},
	}, nil
}

// 没有使用资源维度的权重
func balancedResourceScorer(requested, allocable resourceToValueMap) int64 {
	cpuFraction := fractionOfCapacity(requested[v1.ResourceCPU], allocable[v1.ResourceCPU])
	memoryFraction := fractionOfCapacity(requested[v1.ResourceMemory], allocable[v1.ResourceMemory])
	// This to find a node which has most balanced CPU, memory and volume usage.
	if cpuFraction >= 1 || memoryFraction >= 1 {
		// if requested >= capacity, the corresponding host should never be preferred.
		return 0
	}

	// Upper and lower boundary of difference between cpuFraction and memoryFraction are -1 and 1
	// respectively. Multiplying the absolute value of the difference by `MaxNodeScore` scales the value to
	// 0-MaxNodeScore with 0 representing well balanced allocation and `MaxNodeScore` poorly balanced. Subtracting it from
	// `MaxNodeScore` leads to the score which also scales from 0 to `MaxNodeScore` while `MaxNodeScore` representing well balanced.
	diff := math.Abs(cpuFraction - memoryFraction)
	return int64((1 - diff) * float64(framework.MaxNodeScore))
}

func fractionOfCapacity(requested, capacity int64) float64 {
	if capacity == 0 {
		return 1
	}
	return float64(requested) / float64(capacity)
}
