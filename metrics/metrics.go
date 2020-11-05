// THIS IS FOR TESTING MODULAR
package metrics

type metricsController interface {
	getThreshold() uint64
	setThreshold(metrics uint64) bool
	checkMeetThreshold(current uint64) bool
}

type latencyMet struct {
	latencyThreshold uint64
}

type memMet struct {
	// TODO : type should be uint64
	memThreshold uint64
}

type cpuMet struct {
	// TODO : cpu type should be float64
	cpuThreshold uint64
}

// implement PID Controller
// Latency implemented
func (l latencyMet) setThreshold(metrics uint64) bool {
	l.latencyThreshold = metrics

	return true
}
func (l latencyMet) getThreshold() uint64 {
	return l.latencyThreshold
}
func (l latencyMet) checkMeetThreshold(current uint64) bool {
	return l.latencyThreshold == current
}

// Memory Resource implemented
func (rm memMet) setThreshold(metrics uint64) bool {
	rm.memThreshold = metrics

	return true
}
func (rm memMet) getThreshold() uint64 {
	return rm.memThreshold
}
func (rm memMet) checkMeetThreshold(current uint64) bool {
	return rm.memThreshold == current
}

// CPU Resource implemented
func (rc cpuMet) getThreshold() uint64 {
	return rc.cpuThreshold
}
func (rc cpuMet) setThreshold(metrics uint64) bool {
	rc.cpuThreshold = metrics

	return true
}
func (rc cpuMet) checkMeetThreshold(current uint64) bool {
	return rc.cpuThreshold == current
}
