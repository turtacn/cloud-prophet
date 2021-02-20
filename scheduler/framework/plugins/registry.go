//
//
package plugins

import (
	//"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/defaultbinder"
	//"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/defaultpreemption"
	//"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/imagelocality"
	//"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/interpodaffinity"
	//"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/nodeaffinity"
	//"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/nodelabel"
	//"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/nodename"
	//"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/noderesources"
	//"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/nodeunschedulable"
	//"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/podtopologyspread"
	//"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/queuesort"
	//"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/selectorspread"
	//"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/tainttoleration"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/podtopologyspread"
	// 调度插件在这里引用扩展
	"github.com/turtacn/cloud-prophet/scheduler/framework/runtime"
)

// NewInTreeRegistry builds the registry with all the in-tree plugins.
// A scheduler that runs out of tree plugins can register additional plugins
// through the WithFrameworkOutOfTreeRegistry option.
func NewInTreeRegistry() runtime.Registry {
	return runtime.Registry{
		//selectorspread.Name:                        selectorspread.New,
		//imagelocality.Name:                         imagelocality.New,
		//tainttoleration.Name:                       tainttoleration.New,
		//nodename.Name:                              nodename.New,
		//nodeaffinity.Name:                          nodeaffinity.New,
		podtopologyspread.Name: podtopologyspread.New,
		//nodeunschedulable.Name:                     nodeunschedulable.New,
		//noderesources.FitName:                      noderesources.NewFit,
		//noderesources.BalancedAllocationName:       noderesources.NewBalancedAllocation,
		//noderesources.MostAllocatedName:            noderesources.NewMostAllocated,
		//noderesources.LeastAllocatedName:           noderesources.NewLeastAllocated,
		//noderesources.RequestedToCapacityRatioName: noderesources.NewRequestedToCapacityRatio,
		//interpodaffinity.Name:                      interpodaffinity.New,
		//nodelabel.Name:                             nodelabel.New,
		//queuesort.Name:                             queuesort.New,
		//defaultbinder.Name:                         defaultbinder.New,
		//defaultpreemption.Name:                     defaultpreemption.New,
	}
}
