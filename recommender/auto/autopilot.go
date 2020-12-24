package auto

import (
	"context"
	"flag"
	"github.com/turtacn/cloud-prophet/recommender/logic"
	"github.com/turtacn/cloud-prophet/recommender/model"
	vpa_types "github.com/turtacn/cloud-prophet/recommender/types"
	vpa_utils "github.com/turtacn/cloud-prophet/recommender/util"
	"k8s.io/klog"
	"time"
)

// AggregateContainerStateGCInterval defines how often expired AggregateContainerStates are garbage collected.
const AggregateContainerStateGCInterval = 1 * time.Hour

var (
	checkpointsWriteTimeout = flag.Duration("checkpoints-timeout", time.Minute, `Timeout for writing checkpoints since the start of the recommender's main loop`)
	minCheckpointsPerRun    = flag.Int("min-checkpoints", 10, "Minimum number of checkpoints to write per recommender's main loop")
	memorySaver             = flag.Bool("memory-saver", false, `If true, only track pods which have an associated VPA`)
)

// Recommender recommend resources for certain containers, based on utilization periodically got from metrics api.
type Recommender interface {
	// RunOnce performs one iteration of recommender duties followed by update of recommendations in VPA objects.
	RunOnce(string)
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

func (r *recommender) RunOnce(csv string) {

	ctx := context.Background()
	ctx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(*checkpointsWriteTimeout))
	defer cancelFunc()

	klog.V(3).Infof("Recommender Run")

	// load

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
