package checkpoint

import (
	"context"
	"sort"
	"time"

	vpa_types "github.com/turtacn/cloud-prophet/recommender/types"
	v1 "k8s.io/api/core/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	vpa_api "github.com/turtacn/cloud-prophet/recommender/types"
	// "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/client/clientset/versioned/typed/autoscaling.k8s.io/v1"
	"github.com/turtacn/cloud-prophet/recommender/model"
	api_util "github.com/turtacn/cloud-prophet/recommender/util"
	"k8s.io/klog"
)

// CheckpointWriter persistently stores aggregated historical usage of containers
// controlled by VPA objects. This state can be restored to initialize the model after restart.
type CheckpointWriter interface {
	// StoreCheckpoints writes at least minCheckpoints if there are more checkpoints to write.
	// Checkpoints are written until ctx permits or all checkpoints are written.
	StoreCheckpoints(ctx context.Context, now time.Time, minCheckpoints int) error
}

type checkpointWriter struct {
	vpaCheckpointClient vpa_api.VerticalPodAutoscalerCheckpointsGetter
	cluster             *model.ClusterState
}

// NewCheckpointWriter returns new instance of a CheckpointWriter
func NewCheckpointWriter(cluster *model.ClusterState, vpaCheckpointClient vpa_api.VerticalPodAutoscalerCheckpointsGetter) CheckpointWriter {
	return &checkpointWriter{
		vpaCheckpointClient: vpaCheckpointClient,
		cluster:             cluster,
	}
}

func isFetchingHistory(vpa *model.Vpa) bool {
	condition, found := vpa.Conditions[vpa_types.FetchingHistory]
	if !found {
		return false
	}
	return condition.Status == v1.ConditionTrue
}

func getVpasToCheckpoint(clusterVpas map[model.VpaID]*model.Vpa) []*model.Vpa {
	vpas := make([]*model.Vpa, 0, len(clusterVpas))
	for _, vpa := range clusterVpas {
		if isFetchingHistory(vpa) {
			klog.V(3).Infof("VPA %s/%s is loading history, skipping checkpoints", vpa.ID.Namespace, vpa.ID.VpaName)
			continue
		}
		vpas = append(vpas, vpa)
	}
	sort.Slice(vpas, func(i, j int) bool {
		return vpas[i].CheckpointWritten.Before(vpas[j].CheckpointWritten)
	})
	return vpas
}

func (writer *checkpointWriter) StoreCheckpoints(ctx context.Context, now time.Time, minCheckpoints int) error {
	vpas := getVpasToCheckpoint(writer.cluster.Vpas)
	for _, vpa := range vpas {

		// Draining ctx.Done() channel. ctx.Err() will be checked if timeout occurred, but minCheckpoints have
		// to be written before return from this function.
		select {
		case <-ctx.Done():
		default:
		}

		if ctx.Err() != nil && minCheckpoints <= 0 {
			return ctx.Err()
		}

		aggregateContainerStateMap := buildAggregateContainerStateMap(vpa, writer.cluster, now)
		for container, aggregatedContainerState := range aggregateContainerStateMap {
			containerCheckpoint, err := aggregatedContainerState.SaveToCheckpoint()
			if err != nil {
				klog.Errorf("Cannot serialize checkpoint for vpa %v container %v. Reason: %+v", vpa.ID.VpaName, container, err)
				continue
			}
			//checkpointName := fmt.Sprintf("%s-%s", vpa.ID.VpaName, container)
			vpaCheckpoint := vpa_types.VerticalPodAutoscalerCheckpoint{
				Spec: vpa_types.VerticalPodAutoscalerCheckpointSpec{
					ContainerName: container,
					VPAObjectName: vpa.ID.VpaName,
				},
				Status: *containerCheckpoint,
			}
			err = api_util.CreateOrUpdateVpaCheckpoint(writer.vpaCheckpointClient.VerticalPodAutoscalarCheckpoints(vpa.ID.Namespace), &vpaCheckpoint)
			if err != nil {
				klog.Errorf("Cannot save VPA %s/%s checkpoint for %s. Reason: %+v",
					vpa.ID.Namespace, vpaCheckpoint.Spec.VPAObjectName, vpaCheckpoint.Spec.ContainerName, err)
			} else {
				klog.V(3).Infof("Saved VPA %s/%s checkpoint for %s",
					vpa.ID.Namespace, vpaCheckpoint.Spec.VPAObjectName, vpaCheckpoint.Spec.ContainerName)
				vpa.CheckpointWritten = now
			}
			minCheckpoints--
		}
	}
	return nil
}

// Build the AggregateContainerState for the purpose of the checkpoint. This is an aggregation of state of all
// containers that belong to pods matched by the VPA.
// Note however that we exclude the most recent memory peak for each container (see below).
func buildAggregateContainerStateMap(vpa *model.Vpa, cluster *model.ClusterState, now time.Time) map[string]*model.AggregateContainerState {
	aggregateContainerStateMap := vpa.AggregateStateByContainerName()
	// Note: the memory peak from the current (ongoing) aggregation interval is not included in the
	// checkpoint to avoid having multiple peaks in the same interval after the state is restored from
	// the checkpoint. Therefore we are extracting the current peak from all containers.
	// TODO: Avoid the nested loop over all containers for each VPA.
	for _, pod := range cluster.Pods {
		for containerName, container := range pod.Containers {
			aggregateKey := cluster.MakeAggregateStateKey(pod, containerName)
			if vpa.UsesAggregation(aggregateKey) {
				if aggregateContainerState, exists := aggregateContainerStateMap[containerName]; exists {
					subtractCurrentContainerMemoryPeak(aggregateContainerState, container, now)
				}
			}
		}
	}
	return aggregateContainerStateMap
}

func subtractCurrentContainerMemoryPeak(a *model.AggregateContainerState, container *model.ContainerState, now time.Time) {
	if now.Before(container.WindowEnd) {
		a.AggregateMemoryPeaks.SubtractSample(model.BytesFromMemoryAmount(container.GetMaxMemoryPeak()), 1.0, container.WindowEnd)
	}
}
