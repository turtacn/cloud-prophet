package util

import (
	v1 "k8s.io/api/core/v1"
)

// For each of these resources, a pod that doesn't request the resource explicitly
// will be treated as having requested the amount indicated below, for the purpose
// of computing priority only. This ensures that when scheduling zero-request pods, such
// pods will not all be scheduled to the machine with the smallest in-use request,
// and that when scheduling regular pods, such pods will not see zero-request pods as
// consuming no resources whatsoever. We chose these values to be similar to the
// resources that we give to cluster addon pods (#10653). But they are pretty arbitrary.
// As described in #11713, we use request instead of limit to deal with resource requirements.
const (
	// DefaultMilliCPURequest defines default milli cpu request number.
	DefaultMilliCPURequest int64 = 100 // 0.1 core
	// DefaultMemoryRequest defines default memory request size.
	DefaultMemoryRequest int64 = 200 * 1024 * 1024 // 200 MB
)

// GetNonzeroRequests returns the default cpu and memory resource request if none is found or
// what is provided on the request.
func GetNonzeroRequests(requests *v1.ResourceList) (int64, int64) {
	return GetNonzeroRequestForResource(v1.ResourceCPU, requests),
		GetNonzeroRequestForResource(v1.ResourceMemory, requests)
}

// GetNonzeroRequestForResource returns the default resource request if none is found or
// what is provided on the request.
func GetNonzeroRequestForResource(resource v1.ResourceName, requests *v1.ResourceList) int64 {
	switch resource {
	case v1.ResourceCPU:
		// Override if un-set, but not if explicitly set to zero
		if _, found := (*requests)[v1.ResourceCPU]; !found {
			return DefaultMilliCPURequest
		}
		return requests.Cpu().MilliValue()
	case v1.ResourceMemory:
		// Override if un-set, but not if explicitly set to zero
		if _, found := (*requests)[v1.ResourceMemory]; !found {
			return DefaultMemoryRequest
		}
		return requests.Memory().Value()
	case v1.ResourceEphemeralStorage:
		// if the local storage capacity isolation feature gate is disabled, pods request 0 disk.
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
