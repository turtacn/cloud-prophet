// 聚合状态——样本集合状态
package model

import (
	"fmt"
	"github.com/turtacn/cloud-prophet/recommender/util"
	corev1 "k8s.io/api/core/v1" // 重用kubernetes的多维资源模型
	"math"
	"time"
)

// ContainerNameToAggregateStateMap maps a container name to AggregateContainerState
// that aggregates state of containers with that name.
type ContainerNameToAggregateStateMap map[string]*AggregateContainerState

var (
	// DefaultControlledResources is a default value of Spec.ResourcePolicy.ContainerPolicies[].ControlledResources.
	DefaultControlledResources = []ResourceName{ResourceCPU, ResourceMemory}
)

// ContainerStateAggregator is an interface for objects that consume and
// aggregate container usage samples.
type ContainerStateAggregator interface {
	// AddSample aggregates a single usage sample.
	AddSample(sample *ContainerUsageSample)
	// SubtractSample removes a single usage sample. The subtracted sample
	// should be equal to some sample that was aggregated with AddSample()
	// in the past.
	SubtractSample(sample *ContainerUsageSample)
	// GetLastRecommendation returns last recommendation calculated for this
	// aggregator.
	GetLastRecommendation() corev1.ResourceList
}

// AggregateContainerState holds input signals aggregated from a set of containers.
// It can be used as an input to compute the recommendation.
// The CPU and memory distributions use decaying histograms by default
// (see NewAggregateContainerState()).
// Implements ContainerStateAggregator interface.
type AggregateContainerState struct {
	// AggregateCPUUsage is a distribution of all CPU samples.
	AggregateCPUUsage util.Histogram
	// AggregateMemoryPeaks is a distribution of memory peaks from all containers:
	// each container should add one peak per memory aggregation interval (e.g. once every 24h).
	AggregateMemoryPeaks util.Histogram
	// Note: first/last sample timestamps as well as the sample count are based only on CPU samples.
	FirstSampleStart  time.Time
	LastSampleStart   time.Time
	TotalSamplesCount int
	CreationTime      time.Time

	// Following fields are needed to correctly report quality metrics
	// for VPA. When we record a new sample in an AggregateContainerState
	// we want to know if it needs recommendation, if the recommendation
	// is present and if the automatic updates are on (are we able to
	// apply the recommendation to the pods).
	LastRecommendation  corev1.ResourceList
	ControlledResources *[]ResourceName
}

// GetLastRecommendation returns last recorded recommendation.
func (a *AggregateContainerState) GetLastRecommendation() corev1.ResourceList {
	return a.LastRecommendation
}

// GetControlledResources returns the list of resources controlled by VPA controlling this aggregator.
// Returns default if not set.
func (a *AggregateContainerState) GetControlledResources() []ResourceName {
	if a.ControlledResources != nil {
		return *a.ControlledResources
	}
	return DefaultControlledResources
}

// MergeContainerState merges two AggregateContainerStates.
func (a *AggregateContainerState) MergeContainerState(other *AggregateContainerState) {
	a.AggregateCPUUsage.Merge(other.AggregateCPUUsage)
	a.AggregateMemoryPeaks.Merge(other.AggregateMemoryPeaks)

	if a.FirstSampleStart.IsZero() ||
		(!other.FirstSampleStart.IsZero() && other.FirstSampleStart.Before(a.FirstSampleStart)) {
		a.FirstSampleStart = other.FirstSampleStart
	}
	if other.LastSampleStart.After(a.LastSampleStart) {
		a.LastSampleStart = other.LastSampleStart
	}
	a.TotalSamplesCount += other.TotalSamplesCount
}

// NewAggregateContainerState returns a new, empty AggregateContainerState.
func NewAggregateContainerState() *AggregateContainerState {
	config := GetAggregationsConfig()
	return &AggregateContainerState{
		AggregateCPUUsage:    util.NewDecayingHistogram(config.CPUHistogramOptions, config.CPUHistogramDecayHalfLife),
		AggregateMemoryPeaks: util.NewDecayingHistogram(config.MemoryHistogramOptions, config.MemoryHistogramDecayHalfLife),
		CreationTime:         time.Now(),
	}
}

// AddSample aggregates a single usage sample.
func (a *AggregateContainerState) AddSample(sample *ContainerUsageSample) {
	switch sample.Resource {
	case ResourceCPU:
		a.addCPUSample(sample)
	case ResourceMemory:
		a.AggregateMemoryPeaks.AddSample(BytesFromMemoryAmount(sample.Usage), 1.0, sample.MeasureStart)
	default:
		panic(fmt.Sprintf("AddSample doesn't support resource '%s'", sample.Resource))
	}
}

// SubtractSample removes a single usage sample from an aggregation.
// The subtracted sample should be equal to some sample that was aggregated with
// AddSample() in the past.
// Only memory samples can be subtracted at the moment. Support for CPU could be
// added if necessary.
func (a *AggregateContainerState) SubtractSample(sample *ContainerUsageSample) {
	switch sample.Resource {
	case ResourceMemory:
		a.AggregateMemoryPeaks.SubtractSample(BytesFromMemoryAmount(sample.Usage), 1.0, sample.MeasureStart)
	default:
		panic(fmt.Sprintf("SubtractSample doesn't support resource '%s'", sample.Resource))
	}
}

func (a *AggregateContainerState) addCPUSample(sample *ContainerUsageSample) {
	cpuUsageCores := CoresFromCPUAmount(sample.Usage)
	cpuRequestCores := CoresFromCPUAmount(sample.Request)
	// Samples are added with the weight equal to the current request. This means that
	// whenever the request is increased, the history accumulated so far effectively decays,
	// which helps react quickly to CPU starvation.
	a.AggregateCPUUsage.AddSample(
		cpuUsageCores, math.Max(cpuRequestCores, minSampleWeight), sample.MeasureStart)
	if sample.MeasureStart.After(a.LastSampleStart) {
		a.LastSampleStart = sample.MeasureStart
	}
	if a.FirstSampleStart.IsZero() || sample.MeasureStart.Before(a.FirstSampleStart) {
		a.FirstSampleStart = sample.MeasureStart
	}
	a.TotalSamplesCount++
}

func (a *AggregateContainerState) isExpired(now time.Time) bool {
	if a.isEmpty() {
		return now.Sub(a.CreationTime) >= GetAggregationsConfig().GetMemoryAggregationWindowLength()
	}
	return now.Sub(a.LastSampleStart) >= GetAggregationsConfig().GetMemoryAggregationWindowLength()
}

func (a *AggregateContainerState) isEmpty() bool {
	return a.TotalSamplesCount == 0
}
