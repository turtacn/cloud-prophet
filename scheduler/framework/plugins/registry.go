// 注册到底有哪些插件和类型
//
package plugins

import (
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/defaultbinder"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/interpodaffinity"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/noderesources"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/podtopologyspread"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/queuesort"
	// 调度插件在这里引用扩展
	"github.com/turtacn/cloud-prophet/scheduler/framework/runtime"
)

// NewInTreeRegistry builds the registry with all the in-tree plugins.
// A scheduler that runs out of tree plugins can register additional plugins
// through the WithFrameworkOutOfTreeRegistry option.
func NewInTreeRegistry() runtime.Registry {
	return runtime.Registry{
		podtopologyspread.Name:                     podtopologyspread.New,
		noderesources.FitName:                      noderesources.NewFit,
		noderesources.BalancedAllocationName:       noderesources.NewBalancedAllocation,
		noderesources.MostAllocatedName:            noderesources.NewMostAllocated,
		noderesources.LeastAllocatedName:           noderesources.NewLeastAllocated,
		noderesources.RequestedToCapacityRatioName: noderesources.NewRequestedToCapacityRatio,
		interpodaffinity.Name:                      interpodaffinity.New,
		queuesort.Name:                             queuesort.New,
		defaultbinder.Name:                         defaultbinder.New,
		defaultbinder.NameFakeAllocater:            defaultbinder.NewFA,
	}
}
