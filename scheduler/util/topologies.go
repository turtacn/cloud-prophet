//
package util

import (
	labels "github.com/turtacn/cloud-prophet/scheduler/helper/label"
	"github.com/turtacn/cloud-prophet/scheduler/helper/sets"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
)

func GetNamespacesFromPodAffinityTerm(pod *v1.Pod, podAffinityTerm *v1.PodAffinityTerm) sets.String {
	names := sets.String{}
	if len(podAffinityTerm.Namespaces) == 0 {
		names.Insert(pod.Namespace)
	} else {
		names.Insert(podAffinityTerm.Namespaces...)
	}
	return names
}

func PodMatchesTermsNamespaceAndSelector(pod *v1.Pod, namespaces sets.String, selector labels.Selector) bool {
	if !namespaces.Has(pod.Namespace) {
		return false
	}

	if !selector.Matches(labels.Set(pod.Labels)) {
		return false
	}
	return true
}

func NodesHaveSameTopologyKey(nodeA, nodeB *v1.Node, topologyKey string) bool {
	if len(topologyKey) == 0 {
		return false
	}

	if nodeA.Labels == nil || nodeB.Labels == nil {
		return false
	}

	nodeALabel, okA := nodeA.Labels[topologyKey]
	nodeBLabel, okB := nodeB.Labels[topologyKey]

	if okB && okA {
		return nodeALabel == nodeBLabel
	}

	return false
}

type Topologies struct {
	DefaultKeys []string
}

func (tps *Topologies) NodesHaveSameTopologyKey(nodeA, nodeB *v1.Node, topologyKey string) bool {
	return NodesHaveSameTopologyKey(nodeA, nodeB, topologyKey)
}
