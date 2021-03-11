//
package util

import (
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
)

const (
	DefaultMilliCPURequest int64 = 100               // 0.1 core
	DefaultMemoryRequest   int64 = 200 * 1024 * 1024 // 200 MB
)

func GetNonzeroRequests(requests *v1.ResourceList) (int64, int64) {
	return GetNonzeroRequestForResource(v1.ResourceCPU, requests),
		GetNonzeroRequestForResource(v1.ResourceMemory, requests)
}

func GetNonzeroRequestForResource(resource v1.ResourceName, requests *v1.ResourceList) int64 {
	switch resource {
	case v1.ResourceCPU:
		if _, found := (*requests)[v1.ResourceCPU]; !found {
			return DefaultMilliCPURequest
		}
		return requests.Cpu().MilliValue()
	case v1.ResourceMemory:
		if _, found := (*requests)[v1.ResourceMemory]; !found {
			return DefaultMemoryRequest
		}
		return requests.Memory().Value()
	case v1.ResourceEphemeralStorage:
		quantity, found := (*requests)[v1.ResourceEphemeralStorage]
		if !found {
			return 0
		}
		return quantity.Value()
	default:
		quantity, found := (*requests)[resource]
		if !found {
			return 0
		}
		return quantity.Value()
	}
	return 0
}
