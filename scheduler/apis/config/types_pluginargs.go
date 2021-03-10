//
//
package config

import (
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// InterPodAffinityArgs holds arguments used to configure the InterPodAffinity plugin.
type InterPodAffinityArgs struct {
	metav1.TypeMeta

	// HardPodAffinityWeight is the scoring weight for existing pods with a
	// matching hard affinity to the incoming pod.
	HardPodAffinityWeight int32
}

// NodeResourcesFitArgs holds arguments used to configure the NodeResourcesFit plugin.
type NodeResourcesFitArgs struct {
	metav1.TypeMeta

	// IgnoredResources is the list of resources that NodeResources fit filter
	// should ignore.
	IgnoredResources []string
	// IgnoredResourceGroups defines the list of resource groups that NodeResources fit filter should ignore.
	// e.g. if group is ["example.com"], it will ignore all resource names that begin
	// with "example.com", such as "example.com/aaa" and "example.com/bbb".
	// A resource group name can't contain '/'.
	IgnoredResourceGroups []string
}

// PodTopologySpreadArgs holds arguments used to configure the PodTopologySpread plugin.
type PodTopologySpreadArgs struct {
	metav1.TypeMeta

	// DefaultConstraints defines topology spread constraints to be applied to
	// pods that don't define any in `pod.spec.topologySpreadConstraints`.
	// `topologySpreadConstraint.labelSelectors` must be empty, as they are
	// deduced the pods' membership to Services, Replication Controllers, Replica
	// Sets or Stateful Sets.
	// Empty by default.
	DefaultConstraints []v1.TopologySpreadConstraint // 在label调度空间中指定分布限制
}

// RequestedToCapacityRatioArgs holds arguments used to configure RequestedToCapacityRatio plugin.
type RequestedToCapacityRatioArgs struct {
	metav1.TypeMeta

	// Points defining priority function shape
	Shape []UtilizationShapePoint
	// Resources to be considered when scoring.
	// The default resource set includes "cpu" and "memory" with an equal weight.
	// Allowed weights go from 1 to 100.
	Resources []ResourceSpec
}

// NodeResourcesLeastAllocatedArgs holds arguments used to configure NodeResourcesLeastAllocated plugin.
type NodeResourcesLeastAllocatedArgs struct {
	metav1.TypeMeta

	// Resources to be considered when scoring.
	// The default resource set includes "cpu" and "memory" with an equal weight.
	// Allowed weights go from 1 to 100.
	Resources []ResourceSpec
}

// NodeResourcesMostAllocatedArgs holds arguments used to configure NodeResourcesMostAllocated plugin.
type NodeResourcesMostAllocatedArgs struct {
	metav1.TypeMeta

	// Resources to be considered when scoring.
	// The default resource set includes "cpu" and "memory" with an equal weight.
	// Allowed weights go from 1 to 100.
	Resources []ResourceSpec

	PrintHostTrace bool
}

// UtilizationShapePoint represents a single point of a priority function shape.
type UtilizationShapePoint struct {
	// Utilization (x axis). Valid values are 0 to 100. Fully utilized node maps to 100.
	Utilization int32
	// Score assigned to a given utilization (y axis). Valid values are 0 to 10.
	Score int32
}

// ResourceSpec represents single resource.
type ResourceSpec struct {
	// Name of the resource.
	Name string
	// Weight of the resource.
	Weight int64
}
