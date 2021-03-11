package cache

import (
	framework "github.com/turtacn/cloud-prophet/scheduler/framework/base"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
)

type Cache interface {
	PodCount() (int, error)

	AssumePod(pod *v1.Pod) error

	FinishBinding(pod *v1.Pod) error

	ForgetPod(pod *v1.Pod) error

	AddPod(pod *v1.Pod) error

	UpdatePod(oldPod, newPod *v1.Pod) error

	RemovePod(pod *v1.Pod) error

	GetPod(pod *v1.Pod) (*v1.Pod, error)

	IsAssumedPod(pod *v1.Pod) (bool, error)

	AddNode(node *v1.Node) error

	UpdateNode(oldNode, newNode *v1.Node) error

	RemoveNode(node *v1.Node) error

	UpdateSnapshot(nodeSnapshot *Snapshot) error

	Dump() *Dump
}

type Dump struct {
	AssumedPods map[string]bool
	Nodes       map[string]*framework.NodeInfo
}
