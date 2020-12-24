package main

import (
	"flag"
	"github.com/turtacn/cloud-prophet/recommender/auto"
	"github.com/turtacn/cloud-prophet/recommender/model"
	"k8s.io/klog"
)

// Aggregation configuration flags
var (
	elementId                      = flag.String("element-id", "unknow", `element id for resource usage prediction.`)
	csvFile                        = flag.String("csv-file", "unkowfile", `metrics data file from monitor system, csv file`)
	memoryAggregationInterval      = flag.Duration("memory-aggregation-interval", model.DefaultMemoryAggregationInterval, `The length of a single interval, for which the peak memory usage is computed. Memory usage peaks are aggregated in multiples of this interval. In other words there is one memory usage sample per interval (the maximum usage over that interval)`)
	memoryAggregationIntervalCount = flag.Int64("memory-aggregation-interval-count", model.DefaultMemoryAggregationIntervalCount, `The number of consecutive memory-aggregation-intervals which make up the MemoryAggregationWindowLength which in turn is the period for memory usage aggregation by VPA. In other words, MemoryAggregationWindowLength = memory-aggregation-interval * memory-aggregation-interval-count.`)
	memoryHistogramDecayHalfLife   = flag.Duration("memory-histogram-decay-half-life", model.DefaultMemoryHistogramDecayHalfLife, `The amount of time it takes a historical memory usage sample to lose half of its weight. In other words, a fresh usage sample is twice as 'important' as one with age equal to the half life period.`)
	cpuHistogramDecayHalfLife      = flag.Duration("cpu-histogram-decay-half-life", model.DefaultCPUHistogramDecayHalfLife, `The amount of time it takes a historical CPU usage sample to lose half of its weight.`)
)

func main() {
	klog.InitFlags(nil)

	model.InitializeAggregationsConfig(model.NewAggregationsConfig(*memoryAggregationInterval, *memoryAggregationIntervalCount, *memoryHistogramDecayHalfLife, *cpuHistogramDecayHalfLife))
	recommender := auto.NewRecommender()
	recommender.RunOnce(*elementId, *csvFile)
}
