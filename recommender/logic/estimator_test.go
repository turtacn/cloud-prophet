package logic

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/turtacn/cloud-prophet/recommender/model"
	"github.com/turtacn/cloud-prophet/recommender/util"
)

var (
	anyTime     = time.Unix(0, 0)
	testRequest = model.Resources{
		model.ResourceCPU:    model.CPUAmountFromCores(3.14),
		model.ResourceMemory: model.MemoryAmountFromBytes(3.14e9),
	}
)

// Verifies that the PercentileEstimator returns requested percentiles of CPU
// and memory peaks distributions.
func TestPercentileEstimator(t *testing.T) {
	config := model.GetAggregationsConfig()
	// Create a sample CPU histogram.
	cpuHistogram := util.NewHistogram(config.CPUHistogramOptions)
	cpuHistogram.AddSample(1.0, 1.0, anyTime)
	cpuHistogram.AddSample(2.0, 1.0, anyTime)
	cpuHistogram.AddSample(3.0, 1.0, anyTime)
	// Create a sample memory histogram.
	memoryPeaksHistogram := util.NewHistogram(config.MemoryHistogramOptions)
	memoryPeaksHistogram.AddSample(1e9, 1.0, anyTime)
	memoryPeaksHistogram.AddSample(2e9, 1.0, anyTime)
	memoryPeaksHistogram.AddSample(3e9, 1.0, anyTime)
	// Create an estimator.
	CPUPercentile := 0.2
	MemoryPercentile := 0.5
	estimator := NewPercentileEstimator(CPUPercentile, MemoryPercentile)

	resourceEstimation := estimator.GetResourceEstimation(
		&model.AggregateContainerState{
			AggregateCPUUsage:    cpuHistogram,
			AggregateMemoryPeaks: memoryPeaksHistogram,
		})
	maxRelativeError := 0.05 // Allow 5% relative error to account for histogram rounding.
	assert.InEpsilon(t, 1.0, model.CoresFromCPUAmount(resourceEstimation[model.ResourceCPU]), maxRelativeError)
	assert.InEpsilon(t, 2e9, model.BytesFromMemoryAmount(resourceEstimation[model.ResourceMemory]), maxRelativeError)
}

// Verifies that the confidenceMultiplier calculates the internal
// confidence based on the amount of historical samples and scales the resources
// returned by the base estimator according to the formula, using the calculated
// confidence.
func TestConfidenceMultiplier(t *testing.T) {
	baseEstimator := NewConstEstimator(model.Resources{
		model.ResourceCPU:    model.CPUAmountFromCores(3.14),
		model.ResourceMemory: model.MemoryAmountFromBytes(3.14e9),
	})
	testedEstimator := &confidenceMultiplier{0.1, 2.0, baseEstimator}

	s := model.NewAggregateContainerState()
	// Add 9 CPU samples at the frequency of 1/(2 mins).
	timestamp := anyTime
	for i := 1; i <= 9; i++ {
		s.AddSample(&model.ContainerUsageSample{
			timestamp, model.CPUAmountFromCores(1.0), testRequest[model.ResourceCPU], model.ResourceCPU})
		timestamp = timestamp.Add(time.Minute * 2)
	}

	// Expected confidence = 9/(60*24) = 0.00625.
	assert.Equal(t, 0.00625, getConfidence(s))
	// Expected CPU estimation = 3.14 * (1 + 1/confidence)^exponent =
	// 3.14 * (1 + 0.1/0.00625)^2 = 907.46.
	resourceEstimation := testedEstimator.GetResourceEstimation(s)
	assert.Equal(t, 907.46, model.CoresFromCPUAmount(resourceEstimation[model.ResourceCPU]))
}

// Verifies that the confidenceMultiplier works for the case of no
// history. This corresponds to the multiplier of +INF or 0 (depending on the
// sign of the exponent).
func TestConfidenceMultiplierNoHistory(t *testing.T) {
	baseEstimator := NewConstEstimator(model.Resources{
		model.ResourceCPU:    model.CPUAmountFromCores(3.14),
		model.ResourceMemory: model.MemoryAmountFromBytes(3.14e9),
	})
	testedEstimator1 := &confidenceMultiplier{1.0, 1.0, baseEstimator}
	testedEstimator2 := &confidenceMultiplier{1.0, -1.0, baseEstimator}
	s := model.NewAggregateContainerState()
	// Expect testedEstimator1 to return the maximum possible resource amount.
	assert.Equal(t, model.ResourceAmount(1e14),
		testedEstimator1.GetResourceEstimation(s)[model.ResourceCPU])
	// Expect testedEstimator2 to return zero.
	assert.Equal(t, model.ResourceAmount(0),
		testedEstimator2.GetResourceEstimation(s)[model.ResourceCPU])
}

// Verifies that the MarginEstimator adds margin to the originally
// estimated resources.
func TestMarginEstimator(t *testing.T) {
	// Use 10% margin on top of the recommended resources.
	marginFraction := 0.1
	baseEstimator := NewConstEstimator(model.Resources{
		model.ResourceCPU:    model.CPUAmountFromCores(3.14),
		model.ResourceMemory: model.MemoryAmountFromBytes(3.14e9),
	})
	testedEstimator := &marginEstimator{
		marginFraction: marginFraction,
		baseEstimator:  baseEstimator,
	}
	s := model.NewAggregateContainerState()
	resourceEstimation := testedEstimator.GetResourceEstimation(s)
	assert.Equal(t, 3.14*1.1, model.CoresFromCPUAmount(resourceEstimation[model.ResourceCPU]))
	assert.Equal(t, 3.14e9*1.1, model.BytesFromMemoryAmount(resourceEstimation[model.ResourceMemory]))
}

// Verifies that the MinResourcesEstimator returns at least MinResources.
func TestMinResourcesEstimator(t *testing.T) {

	minResources := model.Resources{
		model.ResourceCPU:    model.CPUAmountFromCores(0.2),
		model.ResourceMemory: model.MemoryAmountFromBytes(4e8),
	}
	baseEstimator := NewConstEstimator(model.Resources{
		model.ResourceCPU:    model.CPUAmountFromCores(3.14),
		model.ResourceMemory: model.MemoryAmountFromBytes(2e7),
	})

	testedEstimator := &minResourcesEstimator{
		minResources:  minResources,
		baseEstimator: baseEstimator,
	}
	s := model.NewAggregateContainerState()
	resourceEstimation := testedEstimator.GetResourceEstimation(s)
	// Original CPU is above min resources
	assert.Equal(t, 3.14, model.CoresFromCPUAmount(resourceEstimation[model.ResourceCPU]))
	// Original Memory is below min resources
	assert.Equal(t, 4e8, model.BytesFromMemoryAmount(resourceEstimation[model.ResourceMemory]))
}
