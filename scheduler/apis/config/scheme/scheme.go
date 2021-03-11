//
package scheme

import (
	kubeschedulerconfig "github.com/turtacn/cloud-prophet/scheduler/apis/config"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewFromSchemeByName(name string) runtime.Object {
	switch {
	case name == "PodTopologySpread":
		return &kubeschedulerconfig.PodTopologySpreadArgs{}
	case name == "NodeResourcesLeastAllocated":
		return &kubeschedulerconfig.NodeResourcesLeastAllocatedArgs{
			Resources: []kubeschedulerconfig.ResourceSpec{
				{Name: "cpu", Weight: 1}, {Name: "memory", Weight: 1},
			},
		}
	case name == "InterPodAffinity":
		return &kubeschedulerconfig.InterPodAffinityArgs{}
	case name == "NodeResourcesFit":
		return &kubeschedulerconfig.NodeResourcesFitArgs{}
	case name == "NodeResourcesMostAllocated":
		return &kubeschedulerconfig.NodeResourcesMostAllocatedArgs{
			Resources: []kubeschedulerconfig.ResourceSpec{
				{Name: "cpu", Weight: 1}, {Name: "memory", Weight: 1},
			},
		}
	case name == "RequestedToCapacityRatio":
		return &kubeschedulerconfig.RequestedToCapacityRatioArgs{
			Shape: []kubeschedulerconfig.UtilizationShapePoint{
				{Utilization: 50, Score: 5}, {Utilization: 60, Score: 6},
			},
			Resources: []kubeschedulerconfig.ResourceSpec{
				{Name: "cpu", Weight: 1}, {Name: "memory", Weight: 1},
			},
		}
	default:
		return nil
	}
	return nil
}
