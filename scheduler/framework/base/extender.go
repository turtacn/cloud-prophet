package base

import (
	extenderv1 "github.com/turtacn/cloud-prophet/scheduler/model"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
)

type Extender interface {
	Name() string

	Filter(pod *v1.Pod, nodes []*v1.Node) (filteredNodes []*v1.Node, failedNodesMap extenderv1.FailedNodesMap, err error)

	Prioritize(pod *v1.Pod, nodes []*v1.Node) (hostPriorities *extenderv1.HostPriorityList, weight int64, err error)

	Bind(binding *v1.Binding) error

	IsBinder() bool

	IsInterested(pod *v1.Pod) bool

	ProcessPreemption(
		pod *v1.Pod,
		nodeNameToVictims map[string]*extenderv1.Victims,
		nodeInfos NodeInfoLister,
	) (map[string]*extenderv1.Victims, error)

	SupportsPreemption() bool

	IsIgnorable() bool
}
