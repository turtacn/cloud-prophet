//
//
package model

import (
	"k8s.io/apimachinery/pkg/api/resource"
)

// Returns string version of ResourceName.
func (self ResourceName) String() string {
	return string(self)
}

// Returns the CPU limit if specified.
func (self *ResourceList) Cpu() *resource.Quantity {
	if val, ok := (*self)[ResourceCPU]; ok {
		return &val
	}
	return &resource.Quantity{Format: resource.DecimalSI}
}

func (self *ResourceList) CpuSub(q *resource.Quantity) {
	if _, ok := (*self)[ResourceCPU]; ok {
		(*self)[ResourceCPU].Sub(*q)
	}
}

func (self *ResourceList) MemSub(q *resource.Quantity) {
	if _, ok := (*self)[ResourceMemory]; ok {
		(*self)[ResourceMemory].Sub(*q)
	}
}

// Returns the Memory limit if specified.
func (self *ResourceList) Memory() *resource.Quantity {
	if val, ok := (*self)[ResourceMemory]; ok {
		return &val
	}
	return &resource.Quantity{Format: resource.BinarySI}
}

// Returns the Storage limit if specified.
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
