//
//
package queue

// Events that trigger scheduler queue to change.
const (
	// Unknown event
	Unknown = "Unknown"
	// PodAdd is the event when a new pod is added to API server.
	PodAdd = "PodAdd"
	// NodeAdd is the event when a new node is added to the cluster.
	NodeAdd = "NodeAdd"
	// ScheduleAttemptFailure is the event when a schedule attempt fails.
	ScheduleAttemptFailure = "ScheduleAttemptFailure"
	// BackoffComplete is the event when a pod finishes backoff.
	BackoffComplete = "BackoffComplete"
	// UnschedulableTimeout is the event when a pod stays in unschedulable for longer than timeout.
	UnschedulableTimeout = "UnschedulableTimeout"
	// AssignedPodAdd is the event when a pod is added that causes pods with matching affinity terms
	// to be more schedulable.
	AssignedPodAdd = "AssignedPodAdd"
	// AssignedPodUpdate is the event when a pod is updated that causes pods with matching affinity
	// terms to be more schedulable.
	AssignedPodUpdate = "AssignedPodUpdate"
	// AssignedPodDelete is the event when a pod is deleted that causes pods with matching affinity
	// terms to be more schedulable.
	AssignedPodDelete = "AssignedPodDelete"

	// 节点app服务开启/关闭，增加/删除
	// ServiceAdd is the event when a service is added in the cluster.
	ServiceAdd = "ServiceAdd"
	// ServiceUpdate is the event when a service is updated in the cluster.
	ServiceUpdate = "ServiceUpdate"
	// ServiceDelete is the event when a service is deleted in the cluster.
	ServiceDelete = "ServiceDelete"

	// 特殊节点增加、更新
	GPUNodeAdd = "GPUNodeAdd"

	// 节点自身调特征变更
	// NodeSpecUnschedulableChange is the event when unschedulable node spec is changed.
	NodeSpecUnschedulableChange = "NodeSpecUnschedulableChange"
	// NodeAllocatableChange is the event when node allocatable is changed.
	NodeAllocatableChange = "NodeAllocatableChange"
	// NodeLabelsChange is the event when node label is changed.
	NodeLabelChange = "NodeLabelChange"
	// NodeTaintsChange is the event when node taint is changed.
	NodeTaintChange = "NodeTaintChange"
	// NodeConditionChange is the event when node condition is changed.
	NodeConditionChange = "NodeConditionChange"
)
