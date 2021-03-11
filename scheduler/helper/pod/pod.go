//
package pod

import (
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
)

func GetPodPriority(pod *v1.Pod) int32 {
	return 0
}

func UpdatePodCondition(status *v1.PodStatus, condition *v1.PodCondition) bool {
	return false
}
