//
//
package scheme

import (
	kubeschedulerconfig "github.com/turtacn/cloud-prophet/scheduler/apis/config"
	"k8s.io/apimachinery/pkg/runtime"
)

// 插件增加，对应的产检参数需要这里集中注册，后续改成通过反射来实现
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
	case name == "NodeResourcesMostAllocated":
		return &kubeschedulerconfig.NodeResourcesMostAllocatedArgs{}
	default:
		return nil
	}
	return nil
}
