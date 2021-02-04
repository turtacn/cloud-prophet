package helper

import (
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
)

func GetPodPriority(pod *v1.Pod) int32 {
	return 0
}
