package model

import (
	"k8s.io/apimachinery/pkg/api/resource"
)

// 借用K8S的概念，资源分配的单位，可扩展支持1:vm ; 2: docker ; 3: nc ; 4: pod
type Pod struct {
}

// 计算节点
type Node struct {
}

// 绑定
type Binding struct {
}

// 节点列表
type NodeList []Node

// 亲和, 支持节点亲和、Pod亲和、Pod反亲和
type Affinity struct {
}

type PodAffinity struct {
}

type ResourceList map[string]resource.Quantity
type ResourceName string
