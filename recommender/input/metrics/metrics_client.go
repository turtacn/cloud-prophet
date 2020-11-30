package metrics

import (
	"time"

	"github.com/turtacn/cloud-prophet/recommender/model"
	k8sapiv1 "k8s.io/api/core/v1"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// ContainerMetricsSnapshot contains information about usage of certain container within defined time window.
type ContainerMetricsSnapshot struct {
	// ID identifies a specific container those metrics are coming from.
	ID model.ContainerID
	// End time of the measurement interval.
	SnapshotTime time.Time
	// Duration of the measurement interval, which is [SnapshotTime - SnapshotWindow, SnapshotTime].
	SnapshotWindow time.Duration
	// Actual usage of the resources over the measurement interval.
	Usage model.Resources
}

// MetricsClient provides simple metrics on resources usage on containter level.
type MetricsClient interface {
	// GetContainersMetrics returns an array of ContainerMetricsSnapshots,
	// representing resource usage for every running container in the cluster
	GetContainersMetrics() ([]*ContainerMetricsSnapshot, error)
}

type metricsClient struct {
	namespace string
}

// NewMetricsClient creates new instance of MetricsClient, which is used by recommender.
// It requires an instance of PodMetricsesGetter, which is used for underlying communication with metrics server.
// namespace limits queries to particular namespace, use k8sapiv1.NamespaceAll to select all namespaces.
func NewMetricsClient(namespace string) MetricsClient {
	return &metricsClient{
		namespace: namespace,
	}
}

func (c *metricsClient) GetContainersMetrics() ([]*ContainerMetricsSnapshot, error) {
	var metricsSnapshots []*ContainerMetricsSnapshot

	return metricsSnapshots, nil
}

func createContainerMetricsSnapshots(podMetrics v1beta1.PodMetrics) []*ContainerMetricsSnapshot {
	snapshots := make([]*ContainerMetricsSnapshot, len(podMetrics.Containers))
	return snapshots
}

func newContainerMetricsSnapshot() *ContainerMetricsSnapshot {

	return &ContainerMetricsSnapshot{
		ID: model.ContainerID{},
	}
}

func calculateUsage(containerUsage k8sapiv1.ResourceList) model.Resources {
	cpuQuantity := containerUsage[k8sapiv1.ResourceCPU]
	cpuMillicores := cpuQuantity.MilliValue()

	memoryQuantity := containerUsage[k8sapiv1.ResourceMemory]
	memoryBytes := memoryQuantity.Value()

	return model.Resources{
		model.ResourceCPU:    model.ResourceAmount(cpuMillicores),
		model.ResourceMemory: model.ResourceAmount(memoryBytes),
	}
}
