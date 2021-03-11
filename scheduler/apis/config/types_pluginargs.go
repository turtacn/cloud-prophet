package config

import (
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type InterPodAffinityArgs struct {
	metav1.TypeMeta

	HardPodAffinityWeight int32
}

type NodeResourcesFitArgs struct {
	metav1.TypeMeta

	IgnoredResources      []string
	IgnoredResourceGroups []string
}

type PodTopologySpreadArgs struct {
	metav1.TypeMeta

	DefaultConstraints []v1.TopologySpreadConstraint // 在label调度空间中指定分布限制
}

type RequestedToCapacityRatioArgs struct {
	metav1.TypeMeta

	Shape     []UtilizationShapePoint
	Resources []ResourceSpec
}

type NodeResourcesLeastAllocatedArgs struct {
	metav1.TypeMeta

	Resources []ResourceSpec
}

type NodeResourcesMostAllocatedArgs struct {
	metav1.TypeMeta

	Resources []ResourceSpec

	PrintHostTrace bool
}

type UtilizationShapePoint struct {
	Utilization int32
	Score       int32
}

type ResourceSpec struct {
	Name   string
	Weight int64
}
