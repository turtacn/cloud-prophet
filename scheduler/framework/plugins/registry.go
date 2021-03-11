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
