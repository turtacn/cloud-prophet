//
//
package scheduler

import (
	"fmt"
	"reflect"

	"k8s.io/klog/v2"

	"github.com/turtacn/cloud-prophet/scheduler/internal/queue"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	"github.com/turtacn/cloud-prophet/scheduler/profile"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	coreinformers "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"
)

func (sched *Scheduler) onServiceAdd(obj interface{}) {
	sched.SchedulingQueue.MoveAllToActiveOrBackoffQueue(queue.ServiceAdd)
}

func (sched *Scheduler) onServiceUpdate(oldObj interface{}, newObj interface{}) {
	sched.SchedulingQueue.MoveAllToActiveOrBackoffQueue(queue.ServiceUpdate)
}

func (sched *Scheduler) onServiceDelete(obj interface{}) {
	sched.SchedulingQueue.MoveAllToActiveOrBackoffQueue(queue.ServiceDelete)
}

func (sched *Scheduler) addNodeToScheduling(obj interface{}) {
	node, ok := obj.(*v1.Node)
	if !ok {
		klog.Errorf("cannot convert to *v1.Node: %v", obj)
		return
	}

	if err := sched.SchedulerCache.AddNode(node); err != nil {
		klog.Errorf("scheduler cache AddNode failed: %v", err)
	}
}

func (sched *Scheduler) addNodeToCache(obj interface{}) {
	node, ok := obj.(*v1.Node)
	if !ok {
		klog.Errorf("cannot convert to *v1.Node: %v", obj)
		return
	}

	if err := sched.SchedulerCache.AddNode(node); err != nil {
		klog.Errorf("scheduler cache AddNode failed: %v", err)
	}

	klog.Infof("add event for node %q", node.Name)
	sched.SchedulingQueue.MoveAllToActiveOrBackoffQueue(queue.NodeAdd)
}

func (sched *Scheduler) updateNodeInCache(oldObj, newObj interface{}) {
	oldNode, ok := oldObj.(*v1.Node)
	if !ok {
		klog.Errorf("cannot convert oldObj to *v1.Node: %v", oldObj)
		return
	}
	newNode, ok := newObj.(*v1.Node)
	if !ok {
		klog.Errorf("cannot convert newObj to *v1.Node: %v", newObj)
		return
	}

	if err := sched.SchedulerCache.UpdateNode(oldNode, newNode); err != nil {
		klog.Errorf("scheduler cache UpdateNode failed: %v", err)
	}

	// Only activate unschedulable pods if the node became more schedulable.
	// We skip the node property comparison when there is no unschedulable pods in the queue
	// to save processing cycles. We still trigger a move to active queue to cover the case
	// that a pod being processed by the scheduler is determined unschedulable. We want this
	// pod to be reevaluated when a change in the cluster happens.
	if sched.SchedulingQueue.NumUnschedulablePods() == 0 {
		sched.SchedulingQueue.MoveAllToActiveOrBackoffQueue(queue.Unknown)
	} else if event := nodeSchedulingPropertiesChange(newNode, oldNode); event != "" {
		sched.SchedulingQueue.MoveAllToActiveOrBackoffQueue(event)
	}
}

func (sched *Scheduler) deleteNodeFromCache(obj interface{}) {
	var node *v1.Node
	switch t := obj.(type) {
	case *v1.Node:
		node = t
	case cache.DeletedFinalStateUnknown:
		var ok bool
		node, ok = t.Obj.(*v1.Node)
		if !ok {
			klog.Errorf("cannot convert to *v1.Node: %v", t.Obj)
			return
		}
	default:
		klog.Errorf("cannot convert to *v1.Node: %v", t)
		return
	}
	klog.Infof("delete event for node %q", node.Name)
	// NOTE: Updates must be written to scheduler cache before invalidating
	// equivalence cache, because we could snapshot equivalence cache after the
	// invalidation and then snapshot the cache itself. If the cache is
	// snapshotted before updates are written, we would update equivalence
	// cache with stale information which is based on snapshot of old cache.
	if err := sched.SchedulerCache.RemoveNode(node); err != nil {
		klog.Errorf("scheduler cache RemoveNode failed: %v", err)
	}
}

func (sched *Scheduler) addPodToSchedulingQueue(obj interface{}) {
	pod := obj.(*v1.Pod)
	klog.Infof("add event for unscheduled pod %s/%s", pod.Namespace, pod.Name)
	if err := sched.SchedulingQueue.Add(pod); err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to queue %T: %v", obj, err))
	}
}

func (sched *Scheduler) updatePodInSchedulingQueue(oldObj, newObj interface{}) {
	pod := newObj.(*v1.Pod)
	if sched.skipPodUpdate(pod) {
		return
	}
	if err := sched.SchedulingQueue.Update(oldObj.(*v1.Pod), pod); err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to update %T: %v", newObj, err))
	}
}

func (sched *Scheduler) deletePodFromSchedulingQueue(obj interface{}) {
	var pod *v1.Pod
	switch t := obj.(type) {
	case *v1.Pod:
		pod = obj.(*v1.Pod)
	case cache.DeletedFinalStateUnknown:
		var ok bool
		pod, ok = t.Obj.(*v1.Pod)
		if !ok {
			utilruntime.HandleError(fmt.Errorf("unable to convert object %T to *v1.Pod in %T", obj, sched))
			return
		}
	default:
		utilruntime.HandleError(fmt.Errorf("unable to handle object in %T: %T", sched, obj))
		return
	}
	klog.Infof("delete event for unscheduled pod %s/%s", pod.Namespace, pod.Name)
	if err := sched.SchedulingQueue.Delete(pod); err != nil {
		utilruntime.HandleError(fmt.Errorf("unable to dequeue %T: %v", obj, err))
	}
	prof, err := sched.profileForPod(pod)
	if err != nil {
		// This shouldn't happen, because we only accept for scheduling the pods
		// which specify a scheduler name that matches one of the profiles.
		klog.Error(err)
		return
	}
	prof.Framework.RejectWaitingPod(pod.UID)
}

func (sched *Scheduler) addPodToCache(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		klog.Errorf("cannot convert to *v1.Pod: %v", obj)
		return
	}
	klog.Infof("add event for scheduled pod %s/%s ", pod.Namespace, pod.Name)

	if err := sched.SchedulerCache.AddPod(pod); err != nil {
		klog.Errorf("scheduler cache AddPod failed: %v", err)
	}

	sched.SchedulingQueue.AssignedPodAdded(pod)
}

func (sched *Scheduler) addPodToScheduling(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		klog.Errorf("cannot convert to *v1.Pod: %v", obj)
		return
	}
	klog.Infof("add event for scheduled pod %s/%s ", pod.Namespace, pod.Name)

	if err := sched.SchedulerCache.AddPod(pod); err != nil {
		klog.Errorf("scheduler cache AddPod failed: %v", err)
	}

	sched.SchedulingQueue.Add(pod)
}

func (sched *Scheduler) updatePodInCache(oldObj, newObj interface{}) {
	oldPod, ok := oldObj.(*v1.Pod)
	if !ok {
		klog.Errorf("cannot convert oldObj to *v1.Pod: %v", oldObj)
		return
	}
	newPod, ok := newObj.(*v1.Pod)
	if !ok {
		klog.Errorf("cannot convert newObj to *v1.Pod: %v", newObj)
		return
	}

	// A Pod delete event followed by an immediate Pod add event may be merged
	// into a Pod update event. In this case, we should invalidate the old Pod, and
	// then add the new Pod.
	if oldPod.UID != newPod.UID {
		sched.deletePodFromCache(oldObj)
		sched.addPodToCache(newObj)
		return
	}

	// NOTE: Updates must be written to scheduler cache before invalidating
	// equivalence cache, because we could snapshot equivalence cache after the
	// invalidation and then snapshot the cache itself. If the cache is
	// snapshotted before updates are written, we would update equivalence
	// cache with stale information which is based on snapshot of old cache.
	if err := sched.SchedulerCache.UpdatePod(oldPod, newPod); err != nil {
		klog.Errorf("scheduler cache UpdatePod failed: %v", err)
	}

	sched.SchedulingQueue.AssignedPodUpdated(newPod)
}

func (sched *Scheduler) deletePodFromCache(obj interface{}) {
	var pod *v1.Pod
	switch t := obj.(type) {
	case *v1.Pod:
		pod = t
	case cache.DeletedFinalStateUnknown:
		var ok bool
		pod, ok = t.Obj.(*v1.Pod)
		if !ok {
			klog.Errorf("cannot convert to *v1.Pod: %v", t.Obj)
			return
		}
	default:
		klog.Errorf("cannot convert to *v1.Pod: %v", t)
		return
	}
	klog.Infof("delete event for scheduled pod %s/%s ", pod.Namespace, pod.Name)
	// NOTE: Updates must be written to scheduler cache before invalidating
	// equivalence cache, because we could snapshot equivalence cache after the
	// invalidation and then snapshot the cache itself. If the cache is
	// snapshotted before updates are written, we would update equivalence
	// cache with stale information which is based on snapshot of old cache.
	if err := sched.SchedulerCache.RemovePod(pod); err != nil {
		klog.Errorf("scheduler cache RemovePod failed: %v", err)
	}

	sched.SchedulingQueue.MoveAllToActiveOrBackoffQueue(queue.AssignedPodDelete)
}

// assignedPod selects pods that are assigned (scheduled and running).
func assignedPod(pod *v1.Pod) bool {
	return len(pod.Spec.NodeName) != 0
}

// responsibleForPod returns true if the pod has asked to be scheduled by the given scheduler.
func responsibleForPod(pod *v1.Pod, profiles profile.Map) bool {
	return profiles.HandlesSchedulerName(pod.Spec.SchedulerName)
}

// skipPodUpdate checks whether the specified pod update should be ignored.
// This function will return true if
//   - The pod has already been assumed, AND
//   - The pod has only its ResourceVersion, Spec.NodeName, Annotations,
//     ManagedFields, Finalizers and/or Conditions updated.
func (sched *Scheduler) skipPodUpdate(pod *v1.Pod) bool {
	// Non-assumed pods should never be skipped.
	isAssumed, err := sched.SchedulerCache.IsAssumedPod(pod)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("failed to check whether pod %s/%s is assumed: %v", pod.Namespace, pod.Name, err))
		return false
	}
	if !isAssumed {
		return false
	}

	// Gets the assumed pod from the cache.
	assumedPod, err := sched.SchedulerCache.GetPod(pod)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("failed to get assumed pod %s/%s from cache: %v", pod.Namespace, pod.Name, err))
		return false
	}

	// Compares the assumed pod in the cache with the pod update. If they are
	// equal (with certain fields excluded), this pod update will be skipped.
	f := func(pod *v1.Pod) *v1.Pod {
		p := pod.DeepCopy()
		// Spec.NodeName must be excluded because the pod assumed in the cache
		// is expected to have a node assigned while the pod update may nor may
		// not have this field set.
		p.Spec.NodeName = ""
		return p
	}
	assumedPodCopy, podCopy := f(assumedPod), f(pod)
	if !reflect.DeepEqual(assumedPodCopy, podCopy) {
		return false
	}
	klog.Infof("Skipping pod %s/%s update", pod.Namespace, pod.Name)
	return true
}

// addAllEventHandlers is a helper function used in tests and in Scheduler
// to add event handlers for various informers.
func addAllEventHandlers(
	sched *Scheduler,
	informerFactory informers.SharedInformerFactory,
	podInformer coreinformers.PodInformer,
) {
	// scheduled pod cache
	podInformer.Informer().AddEventHandler(
		cache.FilteringResourceEventHandler{
			FilterFunc: func(obj interface{}) bool {
				switch t := obj.(type) {
				case *v1.Pod:
					return assignedPod(t)
				case cache.DeletedFinalStateUnknown:
					if pod, ok := t.Obj.(*v1.Pod); ok {
						return assignedPod(pod)
					}
					utilruntime.HandleError(fmt.Errorf("unable to convert object %T to *v1.Pod in %T", obj, sched))
					return false
				default:
					utilruntime.HandleError(fmt.Errorf("unable to handle object in %T: %T", sched, obj))
					return false
				}
			},
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc:    sched.addPodToCache,
				UpdateFunc: sched.updatePodInCache,
				DeleteFunc: sched.deletePodFromCache,
			},
		},
	)
	// unscheduled pod queue
	podInformer.Informer().AddEventHandler(
		cache.FilteringResourceEventHandler{
			FilterFunc: func(obj interface{}) bool {
				switch t := obj.(type) {
				case *v1.Pod:
					return !assignedPod(t) && responsibleForPod(t, sched.Profiles)
				case cache.DeletedFinalStateUnknown:
					if pod, ok := t.Obj.(*v1.Pod); ok {
						return !assignedPod(pod) && responsibleForPod(pod, sched.Profiles)
					}
					utilruntime.HandleError(fmt.Errorf("unable to convert object %T to *v1.Pod in %T", obj, sched))
					return false
				default:
					utilruntime.HandleError(fmt.Errorf("unable to handle object in %T: %T", sched, obj))
					return false
				}
			},
			Handler: cache.ResourceEventHandlerFuncs{
				AddFunc:    sched.addPodToSchedulingQueue,
				UpdateFunc: sched.updatePodInSchedulingQueue,
				DeleteFunc: sched.deletePodFromSchedulingQueue,
			},
		},
	)

	informerFactory.Core().V1().Nodes().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    sched.addNodeToCache,
			UpdateFunc: sched.updateNodeInCache,
			DeleteFunc: sched.deleteNodeFromCache,
		},
	)

	// This is for ServiceAffinity: affected by the selector of the service is updated.
	// Also, if new service is added, equivalence cache will also become invalid since
	// existing pods may be "captured" by this service and change this predicate result.
	informerFactory.Core().V1().Services().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    sched.onServiceAdd,
			UpdateFunc: sched.onServiceUpdate,
			DeleteFunc: sched.onServiceDelete,
		},
	)

}

func nodeSchedulingPropertiesChange(newNode *v1.Node, oldNode *v1.Node) string {
	if nodeSpecUnschedulableChanged(newNode, oldNode) {
		return queue.NodeSpecUnschedulableChange
	}
	if nodeAllocatableChanged(newNode, oldNode) {
		return queue.NodeAllocatableChange
	}
	if nodeLabelsChanged(newNode, oldNode) {
		return queue.NodeLabelChange
	}
	if nodeTaintsChanged(newNode, oldNode) {
		return queue.NodeTaintChange
	}
	if nodeConditionsChanged(newNode, oldNode) {
		return queue.NodeConditionChange
	}

	return ""
}

func nodeAllocatableChanged(newNode *v1.Node, oldNode *v1.Node) bool {
	return !reflect.DeepEqual(oldNode.Status.Allocatable, newNode.Status.Allocatable)
}

func nodeLabelsChanged(newNode *v1.Node, oldNode *v1.Node) bool {
	return !reflect.DeepEqual(oldNode.GetLabels(), newNode.GetLabels())
}

func nodeTaintsChanged(newNode *v1.Node, oldNode *v1.Node) bool {
	return !reflect.DeepEqual(newNode.Spec.Taints, oldNode.Spec.Taints)
}

func nodeConditionsChanged(newNode *v1.Node, oldNode *v1.Node) bool {
	strip := func(conditions []v1.NodeCondition) map[v1.NodeConditionType]v1.ConditionStatus {
		conditionStatuses := make(map[v1.NodeConditionType]v1.ConditionStatus, len(conditions))
		for i := range conditions {
			conditionStatuses[conditions[i].Type] = conditions[i].Status
		}
		return conditionStatuses
	}
	return !reflect.DeepEqual(strip(oldNode.Status.Conditions), strip(newNode.Status.Conditions))
}

func nodeSpecUnschedulableChanged(newNode *v1.Node, oldNode *v1.Node) bool {
	return newNode.Spec.Unschedulable != oldNode.Spec.Unschedulable && newNode.Spec.Unschedulable == false
}
