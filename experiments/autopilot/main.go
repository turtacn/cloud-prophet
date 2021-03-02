package main

import (
	"flag"
	"github.com/turtacn/cloud-prophet/recommender/auto"
	"github.com/turtacn/cloud-prophet/recommender/model"
)

// Aggregation configuration flags
var (
	elementId                 = flag.String("element-id", "unknow", `资源利用率预测的实例ID`)
	csvFile                   = flag.String("csv-file", "unkowfile", `指标数据来自监控系统，csv文件名称`)
	cpuHistogramDecayHalfLife = flag.Duration("cpu-histogram-decay-half-life", model.DefaultCPUHistogramDecayHalfLife, `CPU利用率权重减半的周期，半衰期.`)
)

func main() {
	flag.Parse()
	model.InitializeAggregationsConfig(model.NewAggregationsConfig(model.DefaultMemoryAggregationInterval, model.DefaultMemoryAggregationIntervalCount,
		model.DefaultMemoryHistogramDecayHalfLife, *cpuHistogramDecayHalfLife))
	recommender := auto.NewRecommender()
	recommender.RunOnce(*elementId, *csvFile)
}
