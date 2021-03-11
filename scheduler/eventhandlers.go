package scheduler

import (
	"fmt"
	framework "github.com/turtacn/cloud-prophet/scheduler/framework/base"
	"github.com/turtacn/cloud-prophet/scheduler/internal/queue"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	"github.com/turtacn/cloud-prophet/scheduler/profile"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog/v2"
	"reflect"
)

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
	default:
		klog.Errorf("cannot convert to *v1.Node: %v", t)
		return
	}
	klog.Infof("delete event for node %q", node.Name)
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
	switch obj.(type) {
	case *v1.Pod:
		pod = obj.(*v1.Pod)
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
	klog.Infof("scheduling pod %s/%s ", pod.Namespace, pod.Name)
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

	if oldPod.UID != newPod.UID {
		sched.deletePodFromCache(oldObj)
		sched.addPodToCache(newObj)
		return
	}

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
	default:
		klog.Errorf("cannot convert to *v1.Pod: %v", t)
		return
	}
	klog.Infof("delete event for scheduled pod %s/%s ", pod.Namespace, pod.Name)
	if err := sched.SchedulerCache.RemovePod(pod); err != nil {
		klog.Errorf("scheduler cache RemovePod failed: %v", err)
	}

	sched.SchedulingQueue.MoveAllToActiveOrBackoffQueue(queue.AssignedPodDelete)
}

func assignedPod(pod *v1.Pod) bool {
	return len(pod.Spec.NodeName) != 0
}

func responsibleForPod(pod *v1.Pod, profiles profile.Map) bool {
	return profiles.HandlesSchedulerName(pod.Spec.SchedulerName)
}

func (sched *Scheduler) skipPodUpdate(pod *v1.Pod) bool {
	isAssumed, err := sched.SchedulerCache.IsAssumedPod(pod)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("failed to check whether pod %s/%s is assumed: %v", pod.Namespace, pod.Name, err))
		return false
	}
	if !isAssumed {
		return false
	}

	assumedPod, err := sched.SchedulerCache.GetPod(pod)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("failed to get assumed pod %s/%s from cache: %v", pod.Namespace, pod.Name, err))
		return false
	}

	f := func(pod *v1.Pod) *v1.Pod {
		p := pod.DeepCopy()
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

func addAllEventHandlers(
	sched *Scheduler,
	informerFactory framework.SharedInformer, // 节点informer
	podInformer framework.SharedInformer, // 实例informer
) {
	podInformer.AddEventHandler(
		framework.FilteringResourceEventHandler{
			FilterFunc: func(obj interface{}) bool {
				switch t := obj.(type) {
				case *v1.Pod:
					return assignedPod(t)
				default:
					utilruntime.HandleError(fmt.Errorf("unable to handle object in %T: %T", sched, obj))
					return false
				}
			},
			Handler: framework.ResourceEventHandlerFuncs{
				AddFunc:    sched.addPodToCache,
				UpdateFunc: sched.updatePodInCache,
				DeleteFunc: sched.deletePodFromCache,
			},
		},
	)
	podInformer.AddEventHandler(
		framework.FilteringResourceEventHandler{
			FilterFunc: func(obj interface{}) bool {
				switch t := obj.(type) {
				case *v1.Pod:
					return !assignedPod(t) && responsibleForPod(t, sched.Profiles)
				default:
					utilruntime.HandleError(fmt.Errorf("unable to handle object in %T: %T", sched, obj))
					return false
				}
			},
			Handler: framework.ResourceEventHandlerFuncs{
				AddFunc:    sched.addPodToSchedulingQueue,
				UpdateFunc: sched.updatePodInSchedulingQueue,
				DeleteFunc: sched.deletePodFromSchedulingQueue,
			},
		},
	)

	informerFactory.AddEventHandler(
		framework.ResourceEventHandlerFuncs{
			AddFunc:    sched.addNodeToCache,
			UpdateFunc: sched.updateNodeInCache,
			DeleteFunc: sched.deleteNodeFromCache,
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
