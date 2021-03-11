package noderesources

import (
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	"k8s.io/apimachinery/pkg/api/resource"
)

func makeNode(node string, milliCPU, memory int64) *v1.Node {
	return &v1.Node{
		ObjectMeta: v1.ObjectMeta{Name: node},
		Status: v1.NodeStatus{
			Capacity: v1.ResourceList{
				v1.ResourceCPU:    *resource.NewMilliQuantity(milliCPU, resource.DecimalSI),
				v1.ResourceMemory: *resource.NewQuantity(memory, resource.BinarySI),
			},
			Allocatable: v1.ResourceList{
				v1.ResourceCPU:    *resource.NewMilliQuantity(milliCPU, resource.DecimalSI),
				v1.ResourceMemory: *resource.NewQuantity(memory, resource.BinarySI),
			},
		},
	}
}

func makeNodeWithExtendedResource(node string, milliCPU, memory int64, extendedResource map[string]int64) *v1.Node {
	resourceList := make(map[v1.ResourceName]resource.Quantity)
	for res, quantity := range extendedResource {
		resourceList[v1.ResourceName(res)] = *resource.NewQuantity(quantity, resource.DecimalSI)
	}
	resourceList[v1.ResourceCPU] = *resource.NewMilliQuantity(milliCPU, resource.DecimalSI)
	resourceList[v1.ResourceMemory] = *resource.NewQuantity(memory, resource.BinarySI)
	return &v1.Node{
		ObjectMeta: v1.ObjectMeta{Name: node},
		Status: v1.NodeStatus{
			Capacity:    resourceList,
			Allocatable: resourceList,
		},
	}
}
