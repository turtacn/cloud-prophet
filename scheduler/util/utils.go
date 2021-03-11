package util

import (
	podutil "github.com/turtacn/cloud-prophet/scheduler/helper/pod"
	extenderv1 "github.com/turtacn/cloud-prophet/scheduler/model"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	"k8s.io/klog/v2"
	"time"
)

func GetPodFullName(pod *v1.Pod) string {
	return pod.Name + "_" + pod.Namespace
}

func GetPodStartTime(pod *v1.Pod) time.Time {
	if pod.Status.StartTime != nil {
		return *pod.Status.StartTime
	}

	t := time.Now()
	return t
}

func GetEarliestPodStartTime(victims *extenderv1.Victims) *time.Time {
	if len(victims.Pods) == 0 {
		klog.Errorf("victims.Pods is empty. Should not reach here.")
		return nil
	}

	earliestPodStartTime := GetPodStartTime(victims.Pods[0])
	maxPriority := podutil.GetPodPriority(victims.Pods[0])

	for _, pod := range victims.Pods {
		if podutil.GetPodPriority(pod) == maxPriority {
			if GetPodStartTime(pod).Before(earliestPodStartTime) {
				earliestPodStartTime = GetPodStartTime(pod)
			}
		} else if podutil.GetPodPriority(pod) > maxPriority {
			maxPriority = podutil.GetPodPriority(pod)
			earliestPodStartTime = GetPodStartTime(pod)
		}
	}

	return &earliestPodStartTime
}

func MoreImportantPod(pod1, pod2 *v1.Pod) bool {
	p1 := podutil.GetPodPriority(pod1)
	p2 := podutil.GetPodPriority(pod2)
	if p1 != p2 {
		return p1 > p2
	}
	return GetPodStartTime(pod1).Before(GetPodStartTime(pod2))
}

func GetPodAffinityTerms(affinity *v1.Affinity) (terms []v1.PodAffinityTerm) {
	if affinity != nil && affinity.PodAffinity != nil {
		if len(affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != 0 {
			terms = affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution
		}
	}
	return terms
}

func GetPodAntiAffinityTerms(affinity *v1.Affinity) (terms []v1.PodAffinityTerm) {
	if affinity != nil && affinity.PodAntiAffinity != nil {
		if len(affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != 0 {
			terms = affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution
		}
	}
	return terms
}
