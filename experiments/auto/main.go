package main

import (
	"flag"
	"time"

	"github.com/turtacn/cloud-prophet/recommender/auto"
	"github.com/turtacn/cloud-prophet/recommender/input/history"
	"github.com/turtacn/cloud-prophet/recommender/model"
	kube_flag "k8s.io/component-base/cli/flag"
	"k8s.io/klog"
)

var (
	checkpointsGCInterval = flag.Duration("checkpoints-gc-interval", 10*time.Minute, `How often orphaned checkpoints should be garbage collected`)
	storage               = flag.String("storage", "", `Specifies storage mode. Supported values: prometheus, checkpoint (default)`)
	// prometheus history provider configs
	historyLength     = flag.String("history-length", "8d", `How much time back prometheus have to be queried to get historical metrics`)
	historyResolution = flag.String("history-resolution", "1h", `Resolution at which Prometheus is queried for historical metrics`)
)

// Aggregation configuration flags
var (
	memoryAggregationInterval      = flag.Duration("memory-aggregation-interval", model.DefaultMemoryAggregationInterval, `The length of a single interval, for which the peak memory usage is computed. Memory usage peaks are aggregated in multiples of this interval. In other words there is one memory usage sample per interval (the maximum usage over that interval)`)
	memoryAggregationIntervalCount = flag.Int64("memory-aggregation-interval-count", model.DefaultMemoryAggregationIntervalCount, `The number of consecutive memory-aggregation-intervals which make up the MemoryAggregationWindowLength which in turn is the period for memory usage aggregation by VPA. In other words, MemoryAggregationWindowLength = memory-aggregation-interval * memory-aggregation-interval-count.`)
	memoryHistogramDecayHalfLife   = flag.Duration("memory-histogram-decay-half-life", model.DefaultMemoryHistogramDecayHalfLife, `The amount of time it takes a historical memory usage sample to lose half of its weight. In other words, a fresh usage sample is twice as 'important' as one with age equal to the half life period.`)
	cpuHistogramDecayHalfLife      = flag.Duration("cpu-histogram-decay-half-life", model.DefaultCPUHistogramDecayHalfLife, `The amount of time it takes a historical CPU usage sample to lose half of its weight.`)
)

func main() {
	klog.InitFlags(nil)
	kube_flag.InitFlags()

	model.InitializeAggregationsConfig(model.NewAggregationsConfig(*memoryAggregationInterval, *memoryAggregationIntervalCount, *memoryHistogramDecayHalfLife, *cpuHistogramDecayHalfLife))

	useCheckpoints := *storage != "prometheus"
	recommender := auto.NewRecommender(nil, *checkpointsGCInterval, useCheckpoints, "")

	if useCheckpoints {
		recommender.GetClusterStateFeeder().InitFromCheckpoints()
	} else {
		config := history.PrometheusHistoryProviderConfig{
			HistoryLength:     *historyLength,
			HistoryResolution: *historyResolution,
		}
		provider, err := history.NewPrometheusHistoryProvider(config)
		if err != nil {
			klog.Fatalf("Could not initialize history provider: %v", err)
		}
		recommender.GetClusterStateFeeder().InitFromHistoryProvider(provider)
	}
	recommender.RunOnce()
}
