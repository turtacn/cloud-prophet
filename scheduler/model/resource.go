//
package model

import (
	"k8s.io/apimachinery/pkg/api/resource"
)

func (self ResourceName) String() string {
	return string(self)
}

func (self *ResourceList) Cpu() *resource.Quantity {
	if val, ok := (*self)[ResourceCPU]; ok {
		return &val
	}
	return &resource.Quantity{Format: resource.DecimalSI}
}

func (self *ResourceList) CpuSub(q *resource.Quantity) {
	if _, ok := (*self)[ResourceCPU]; ok {
		r := (*self)[ResourceCPU]
		r.Sub(*q)
		(*self)[ResourceCPU] = r
	}
}

func (self *ResourceList) MemSub(q *resource.Quantity) {
	if _, ok := (*self)[ResourceMemory]; ok {
		r := (*self)[ResourceMemory]
		r.Sub(*q)
		(*self)[ResourceMemory] = r
	}
}

func (self *ResourceList) Memory() *resource.Quantity {
	if val, ok := (*self)[ResourceMemory]; ok {
		return &val
	}
	return &resource.Quantity{Format: resource.BinarySI}
}

func (self *ResourceList) Storage() *resource.Quantity {
	if val, ok := (*self)[ResourceStorage]; ok {
		return &val
	}
	return &resource.Quantity{Format: resource.BinarySI}
}

func (self *ResourceList) Pods() *resource.Quantity {
	if val, ok := (*self)[ResourcePods]; ok {
		return &val
	}
	return &resource.Quantity{}
}

func (self *ResourceList) StorageEphemeral() *resource.Quantity {
	if val, ok := (*self)[ResourceEphemeralStorage]; ok {
		return &val
	}
	return &resource.Quantity{}
}
