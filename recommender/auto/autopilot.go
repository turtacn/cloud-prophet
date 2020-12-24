package auto

import (
	"context"
	"flag"
	"fmt"
	"github.com/turtacn/cloud-prophet/recommender/logic"
	"github.com/turtacn/cloud-prophet/recommender/model"
	vpa_types "github.com/turtacn/cloud-prophet/recommender/types"
	vpa_utils "github.com/turtacn/cloud-prophet/recommender/util"
	"k8s.io/klog"
	"time"
)

var (
	runOnceTimeout       = flag.Duration("runonce-timeout", time.Hour, `一次运行的最大超时时间`)
	sampleSecondInterval = flag.Int("sample-second-interval", 60, `样本的采样间隔，单位(秒)，整型`)
)

// Recommender recommend resources for certain containers, based on utilization periodically got from metrics api.
type Recommender interface {
	// RunOnce performs one iteration of recommender
	// elementid , csv data
	RunOnce(string, string)
}

type recommender struct {
	podResourceRecommender logic.PodResourceRecommender
}

// getCappedRecommendation creates a recommendation based on recommended pod
// resources, setting the UncappedTarget to the calculated recommended target
// and if necessary, capping the Target, LowerBound and UpperBound according
// to the ResourcePolicy.
func getCappedRecommendation(vpaID model.VpaID, resources logic.RecommendedPodResources,
	policy *vpa_types.PodResourcePolicy) *vpa_types.RecommendedPodResources {
	containerResources := make([]vpa_types.RecommendedContainerResources, 0, len(resources))
	for containerName, res := range resources {
		containerResources = append(containerResources, vpa_types.RecommendedContainerResources{
			ContainerName:  containerName,
			Target:         model.ResourcesAsResourceList(res.Target),
			LowerBound:     model.ResourcesAsResourceList(res.LowerBound),
			UpperBound:     model.ResourcesAsResourceList(res.UpperBound),
			UncappedTarget: model.ResourcesAsResourceList(res.Target),
		})
	}
	recommendation := &vpa_types.RecommendedPodResources{containerResources}
	cappedRecommendation, err := vpa_utils.ApplyVPAPolicy(recommendation, policy)
	if err != nil {
		klog.Errorf("Failed to apply policy for VPA %v/%v: %v", vpaID.Namespace, vpaID.VpaName, err)
		return recommendation
	}
	return cappedRecommendation
}

func laodCSVData(csv string) []float64 {
	return nil
}

func (r *recommender) RunOnce(element, csv string) {

	ctx := context.Background()
	ctx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(*runOnceTimeout))
	defer cancelFunc()

	klog.V(0).Infof("Recommender Run")
	// load

	var anyTime = time.Unix(0, 0)
	entityAggregateStateMap := make(model.ContainerNameToAggregateStateMap)
	entityAggregateStateMap[element] = model.NewAggregateContainerState()

	data := laodCSVData(csv)
	for _, s := range entityAggregateStateMap {
		timestamp := anyTime
		for _, d := range data {
			s.AddSample(&model.ContainerUsageSample{
				timestamp, model.CPUAmountFromCores(d / 100), model.CPUAmountFromCores(1), model.ResourceCPU})

			timestamp = timestamp.Add(time.Duration(*sampleSecondInterval) * time.Second)

			resources := r.podResourceRecommender.GetRecommendedPodResources(entityAggregateStateMap)
			containerResources := make([]vpa_types.RecommendedContainerResources, 0, len(resources))
			for containerName, res := range resources {
				containerResources = append(containerResources, vpa_types.RecommendedContainerResources{
					ContainerName:  containerName,
					Target:         model.ResourcesAsResourceList(res.Target),
					LowerBound:     model.ResourcesAsResourceList(res.LowerBound),
					UpperBound:     model.ResourcesAsResourceList(res.UpperBound),
					UncappedTarget: model.ResourcesAsResourceList(res.Target),
				})
			}
			recommendation := &vpa_types.RecommendedPodResources{containerResources}

			for _, recon := range recommendation.ContainerRecommendations {

				recommendString := fmt.Sprintf("%s,%f",
					recon.Target.Cpu().AsDec().String(), d)
				klog.Info(recommendString)
			}
		}
	}

	// upodate vpa
	// gc
	// maintain checkpoint

}

// RecommenderFactory makes instances of Recommender.
type RecommenderFactory struct {
	PodResourceRecommender logic.PodResourceRecommender
	CheckpointsGCInterval  time.Duration
	UseCheckpoints         bool
}

// Make creates a new recommender instance,
// which can be run in order to provide continuous resource recommendations for containers.
func (c RecommenderFactory) Make() Recommender {
	recommender := &recommender{
		podResourceRecommender: c.PodResourceRecommender,
	}
	klog.V(3).Infof("New Recommender created %+v", recommender)
	return recommender
}

// NewRecommender creates a new recommender instance.
// Dependencies are created automatically.
// Deprecated; use RecommenderFactory instead.
func NewRecommender() Recommender {
	return RecommenderFactory{
		PodResourceRecommender: logic.CreatePodResourceRecommender(),
	}.Make()
}
