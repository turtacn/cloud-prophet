//
package queuesort

import (
	framework "github.com/turtacn/cloud-prophet/scheduler/framework/base"
	podutil "github.com/turtacn/cloud-prophet/scheduler/helper/pod"
	"k8s.io/apimachinery/pkg/runtime"
)

const Name = "PrioritySort"

type PrioritySort struct{}

var _ framework.QueueSortPlugin = &PrioritySort{}

func (pl *PrioritySort) Name() string {
	return Name
}

func (pl *PrioritySort) Less(pInfo1, pInfo2 *framework.QueuedPodInfo) bool {
	p1 := podutil.GetPodPriority(pInfo1.Pod)
	p2 := podutil.GetPodPriority(pInfo2.Pod)
	return (p1 > p2) || (p1 == p2 && pInfo1.Timestamp.Before(pInfo2.Timestamp))
}

func New(_ runtime.Object, handle framework.FrameworkHandle) (framework.Plugin, error) {
	return &PrioritySort{}, nil
}
