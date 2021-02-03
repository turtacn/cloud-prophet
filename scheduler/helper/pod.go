package helper

import (
	v1 "k8s.io/api/core/v1"
)

func GetPodPriority(pod *v1.Pod) int32 {
	return 0
}
