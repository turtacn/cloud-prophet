//
//
package metrics

// This file contains helpers for metrics that are associated to a profile.

var (
	scheduledResult     = "scheduled"
	unschedulableResult = "unschedulable"
	errorResult         = "error"
)

// PodScheduled can records a successful scheduling attempt and the duration
// since `start`.
func PodScheduled(profile string, duration float64) {
	observeScheduleAttemptAndLatency(scheduledResult, profile, duration)
}

// PodUnschedulable can records a scheduling attempt for an unschedulable pod
// and the duration since `start`.
func PodUnschedulable(profile string, duration float64) {
	observeScheduleAttemptAndLatency(unschedulableResult, profile, duration)
}

// PodScheduleError can records a scheduling attempt that had an error and the
// duration since `start`.
func PodScheduleError(profile string, duration float64) {
	observeScheduleAttemptAndLatency(errorResult, profile, duration)
}

func observeScheduleAttemptAndLatency(result, profile string, duration float64) {
	e2eSchedulingLatency.WithLabelValues(result, profile).Observe(duration)
	scheduleAttempts.WithLabelValues(result, profile).Inc()
}
