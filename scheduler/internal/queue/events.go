package queue

const (
	Unknown                = "Unknown"
	PodAdd                 = "PodAdd"
	NodeAdd                = "NodeAdd"
	ScheduleAttemptFailure = "ScheduleAttemptFailure"
	BackoffComplete        = "BackoffComplete"
	UnschedulableTimeout   = "UnschedulableTimeout"
	AssignedPodAdd         = "AssignedPodAdd"
	AssignedPodUpdate      = "AssignedPodUpdate"
	AssignedPodDelete      = "AssignedPodDelete"

	ServiceAdd    = "ServiceAdd"
	ServiceUpdate = "ServiceUpdate"
	ServiceDelete = "ServiceDelete"

	GPUNodeAdd = "GPUNodeAdd"

	NodeSpecUnschedulableChange = "NodeSpecUnschedulableChange"
	NodeAllocatableChange       = "NodeAllocatableChange"
	NodeLabelChange             = "NodeLabelChange"
	NodeTaintChange             = "NodeTaintChange"
	NodeConditionChange         = "NodeConditionChange"
)
