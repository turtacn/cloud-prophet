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

	NodeSpecUnschedulableChange = "NodeSpecUnschedulableChange"
	NodeAllocatableChange       = "NodeAllocatableChange"
	NodeLabelChange             = "NodeLabelChange"
	NodeTaintChange             = "NodeTaintChange"
	NodeConditionChange         = "NodeConditionChange"
)
