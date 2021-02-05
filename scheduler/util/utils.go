package util

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	podutil "github.com/turtacn/cloud-prophet/scheduler/helper"
	extenderv1 "github.com/turtacn/cloud-prophet/scheduler/model"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// GetPodFullName returns a name that uniquely identifies a pod.
func GetPodFullName(pod *v1.Pod) string {
	// Use underscore as the delimiter because it is not allowed in pod name
	// (DNS subdomain format).
	return pod.Name + "_" + pod.Namespace
}

// GetPodStartTime returns start time of the given pod or current timestamp
// if it hasn't started yet.
func GetPodStartTime(pod *v1.Pod) time.Time {
	if pod.Status.StartTime != nil {
		return *pod.Status.StartTime
	}

	t := time.Now()
	// Assumed pods and bound pods that haven't started don't have a StartTime yet.
	return t
}

// GetEarliestPodStartTime returns the earliest start time of all pods that
// have the highest priority among all victims.
func GetEarliestPodStartTime(victims *extenderv1.Victims) *time.Time {
	if len(victims.Pods) == 0 {
		// should not reach here.
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

// MoreImportantPod return true when priority of the first pod is higher than
// the second one. If two pods' priorities are equal, compare their StartTime.
// It takes arguments of the type "interface{}" to be used with SortableList,
// but expects those arguments to be *v1.Pod.
func MoreImportantPod(pod1, pod2 *v1.Pod) bool {
	p1 := podutil.GetPodPriority(pod1)
	p2 := podutil.GetPodPriority(pod2)
	if p1 != p2 {
		return p1 > p2
	}
	return GetPodStartTime(pod1).Before(GetPodStartTime(pod2))
}

// GetPodAffinityTerms gets pod affinity terms by a pod affinity object.
func GetPodAffinityTerms(affinity *v1.Affinity) (terms []v1.PodAffinityTerm) {
	if affinity != nil && affinity.PodAffinity != nil {
		if len(affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != 0 {
			terms = affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution
		}
	}
	return terms
}

// GetPodAntiAffinityTerms gets pod affinity terms by a pod anti-affinity.
func GetPodAntiAffinityTerms(affinity *v1.Affinity) (terms []v1.PodAffinityTerm) {
	if affinity != nil && affinity.PodAntiAffinity != nil {
		if len(affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution) != 0 {
			terms = affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution
		}
		// TODO: Uncomment this block when implement RequiredDuringSchedulingRequiredDuringExecution.
		//if len(affinity.PodAntiAffinity.RequiredDuringSchedulingRequiredDuringExecution) != 0 {
		//	terms = append(terms, affinity.PodAntiAffinity.RequiredDuringSchedulingRequiredDuringExecution...)
		//}
	}
	return terms
}

// PatchPod calculates the delta bytes change from <old> to <new>,
// and then submit a request to API server to patch the pod changes.
func PatchPod(cs kubernetes.Interface, old *v1.Pod, new *v1.Pod) error {
	oldData, err := json.Marshal(old)
	if err != nil {
		return err
	}

	newData, err := json.Marshal(new)
	if err != nil {
		return err
	}
	patchBytes, err := strategicpatch.CreateTwoWayMergePatch(oldData, newData, &v1.Pod{})
	if err != nil {
		return fmt.Errorf("failed to create merge patch for pod %q/%q: %v", old.Namespace, old.Name, err)
	}
	_, err = cs.CoreV1().Pods(old.Namespace).Patch(context.TODO(), old.Name, types.StrategicMergePatchType, patchBytes, metav1.PatchOptions{}, "status")
	return err
}

// GetUpdatedPod returns the latest version of <pod> from API server.
func GetUpdatedPod(cs kubernetes.Interface, pod *v1.Pod) (*v1.Pod, error) {
	return nil, nil
}

// DeletePod deletes the given <pod> from API server
func DeletePod(cs kubernetes.Interface, pod *v1.Pod) error {
	return cs.CoreV1().Pods(pod.Namespace).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
}

// ClearNominatedNodeName internally submit a patch request to API server
// to set each pods[*].Status.NominatedNodeName> to "".
func ClearNominatedNodeName(cs kubernetes.Interface, pods ...*v1.Pod) utilerrors.Aggregate {
	var errs []error
	for _, p := range pods {
		if len(p.Status.NominatedNodeName) == 0 {
			continue
		}
		podCopy := p.DeepCopy()
		podCopy.Status.NominatedNodeName = ""
		if err := PatchPod(cs, p, podCopy); err != nil {
			errs = append(errs, err)
		}
	}
	return utilerrors.NewAggregate(errs)
}
