//
package plugins

import (
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/defaultbinder"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/interpodaffinity"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/noderesources"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/podtopologyspread"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/queuesort"
	"github.com/turtacn/cloud-prophet/scheduler/framework/runtime"
)

func NewInTreeRegistry() runtime.Registry {
	return runtime.Registry{
		podtopologyspread.Name:                     podtopologyspread.New,                     // 在更小的调度空间中pod 与 node之间的 spread 偏差
		noderesources.FitName:                      noderesources.NewFit,                      // 资源够不够
		noderesources.BalancedAllocationName:       noderesources.NewBalancedAllocation,       //
		noderesources.MostAllocatedName:            noderesources.NewMostAllocated,            //
		noderesources.LeastAllocatedName:           noderesources.NewLeastAllocated,           //
		noderesources.RequestedToCapacityRatioName: noderesources.NewRequestedToCapacityRatio, //
		interpodaffinity.Name:                      interpodaffinity.New,                      //
		queuesort.Name:                             queuesort.New,
		defaultbinder.Name:                         defaultbinder.New,
		defaultbinder.NameFakeAllocater:            defaultbinder.NewFA, // binding 的最后一步，后面立即持久化
	}
}
