package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetContainersMetricsReturnsEmptyList(t *testing.T) {
	tc := newEmptyMetricsClientTestCase()
	emptyMetricsClient := tc.createFakeMetricsClient()

	containerMetricsSnapshots, err := emptyMetricsClient.GetContainersMetrics()

	assert.NoError(t, err)
	assert.Empty(t, containerMetricsSnapshots, "should be empty for empty MetricsGetter")
}

func TestGetContainersMetricsReturnsResults(t *testing.T) {
	tc := newMetricsClientTestCase()
	fakeMetricsClient := tc.createFakeMetricsClient()

	snapshots, err := fakeMetricsClient.GetContainersMetrics()

	assert.NoError(t, err)
	assert.Len(t, snapshots, len(tc.getAllSnaps()), "It should return right number of snapshots")
	for _, snap := range snapshots {
		assert.Contains(t, tc.getAllSnaps(), snap, "One of returned ContainerMetricsSnapshot is different then expected ")
	}
}
