//
package model

const (
	// TaintNodeNotReady will be added when node is not ready
	// and removed when node becomes ready.
	TaintNodeNotReady = "node.free.io/not-ready"

	// TaintNodeUnreachable will be added when node becomes unreachable
	// (corresponding to NodeReady status ConditionUnknown)
	// and removed when node becomes reachable (NodeReady status ConditionTrue).
	TaintNodeUnreachable = "node.free.io/unreachable"

	// TaintNodeUnschedulable will be added when node becomes unschedulable
	// and removed when node becomes scheduable.
	TaintNodeUnschedulable = "node.free.io/unschedulable"

	// TaintNodeMemoryPressure will be added when node has memory pressure
	// and removed when node has enough memory.
	TaintNodeMemoryPressure = "node.free.io/memory-pressure"

	// TaintNodeDiskPressure will be added when node has disk pressure
	// and removed when node has enough disk.
	TaintNodeDiskPressure = "node.free.io/disk-pressure"

	// TaintNodeNetworkUnavailable will be added when node's network is unavailable
	// and removed when network becomes ready.
	TaintNodeNetworkUnavailable = "node.free.io/network-unavailable"

	// TaintNodePIDPressure will be added when node has pid pressure
	// and removed when node has enough disk.
	TaintNodePIDPressure = "node.free.io/pid-pressure"
)
