package noderesources

import (
	"context"
	"fmt"
	"math"

	"github.com/turtacn/cloud-prophet/scheduler/apis/config"
	"github.com/turtacn/cloud-prophet/scheduler/apis/config/validation"
	framework "github.com/turtacn/cloud-prophet/scheduler/framework/base"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	RequestedToCapacityRatioName = "RequestedToCapacityRatio"
	maxUtilization               = 100
)

type functionShape []functionShapePoint

type functionShapePoint struct {
	utilization int64
	score       int64
}

func NewRequestedToCapacityRatio(plArgs runtime.Object, handle framework.FrameworkHandle) (framework.Plugin, error) {
	args, err := getRequestedToCapacityRatioArgs(plArgs)
	if err != nil {
		return nil, err
	}

	if err := validation.ValidateRequestedToCapacityRatioArgs(args); err != nil {
		return nil, err
	}

	shape := make([]functionShapePoint, 0, len(args.Shape))
	for _, point := range args.Shape {
		shape = append(shape, functionShapePoint{
			utilization: int64(point.Utilization),
			score:       int64(point.Score) * (framework.MaxNodeScore / config.MaxCustomPriorityScore),
		})
	}

	resourceToWeightMap := make(resourceToWeightMap)
	for _, resource := range args.Resources {
		resourceToWeightMap[v1.ResourceName(resource.Name)] = resource.Weight
		if resource.Weight == 0 {
			resourceToWeightMap[v1.ResourceName(resource.Name)] = 1
		}
	}

	return &RequestedToCapacityRatio{
		handle: handle,
		resourceAllocationScorer: resourceAllocationScorer{
			RequestedToCapacityRatioName,
			buildRequestedToCapacityRatioScorerFunction(shape, resourceToWeightMap),
			resourceToWeightMap,
			false,
		},
	}, nil
}

func getRequestedToCapacityRatioArgs(obj runtime.Object) (config.RequestedToCapacityRatioArgs, error) {
	ptr, ok := obj.(*config.RequestedToCapacityRatioArgs)
	if !ok {
		return config.RequestedToCapacityRatioArgs{}, fmt.Errorf("want args to be of type RequestedToCapacityRatioArgs, got %T", obj)
	}
	return *ptr, nil
}

type RequestedToCapacityRatio struct {
	handle framework.FrameworkHandle
	resourceAllocationScorer
}

var _ framework.ScorePlugin = &RequestedToCapacityRatio{}

func (pl *RequestedToCapacityRatio) Name() string {
	return RequestedToCapacityRatioName
}

func (pl *RequestedToCapacityRatio) Score(ctx context.Context, _ *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	nodeInfo, err := pl.handle.SnapshotSharedLister().NodeInfos().Get(nodeName)
	if err != nil {
		return 0, framework.NewStatus(framework.Error, fmt.Sprintf("getting node %q from Snapshot: %v", nodeName, err))
	}
	return pl.score(pod, nodeInfo)
}

func (pl *RequestedToCapacityRatio) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

func buildRequestedToCapacityRatioScorerFunction(scoringFunctionShape functionShape, resourceToWeightMap resourceToWeightMap) func(resourceToValueMap, resourceToValueMap) int64 {
	rawScoringFunction := buildBrokenLinearFunction(scoringFunctionShape)
	resourceScoringFunction := func(requested, capacity int64) int64 {
		if capacity == 0 || requested > capacity {
			return rawScoringFunction(maxUtilization)
		}

		return rawScoringFunction(maxUtilization - (capacity-requested)*maxUtilization/capacity)
	}
	return func(requested, allocable resourceToValueMap) int64 {
		var nodeScore, weightSum int64
		for resource, weight := range resourceToWeightMap {
			resourceScore := resourceScoringFunction(requested[resource], allocable[resource])
			if resourceScore > 0 {
				nodeScore += resourceScore * weight
				weightSum += weight
			}
		}
		if weightSum == 0 {
			return 0
		}
		return int64(math.Round(float64(nodeScore) / float64(weightSum)))
	}
}

func buildBrokenLinearFunction(shape functionShape) func(int64) int64 {
	return func(p int64) int64 {
		for i := 0; i < len(shape); i++ {
			if p <= int64(shape[i].utilization) {
				if i == 0 {
					return shape[0].score
				}
				return shape[i-1].score + (shape[i].score-shape[i-1].score)*(p-shape[i-1].utilization)/(shape[i].utilization-shape[i-1].utilization)
			}
		}
		return shape[len(shape)-1].score
	}
}
