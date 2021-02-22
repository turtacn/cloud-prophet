//
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
		return &kubeschedulerconfig.NodeResourcesLeastAllocatedArgs{}
	case name == "InterPodAffinity":
		return &kubeschedulerconfig.InterPodAffinityArgs{}
	case name == "NodeResourcesFit":
		return &kubeschedulerconfig.NodeResourcesFitArgs{}
	default:
		return nil
	}
	return nil
}
