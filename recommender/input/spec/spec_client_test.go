package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPodSpecsReturnsNoResults(t *testing.T) {
	// given
	tc := newEmptySpecClientTestCase()
	client := tc.createFakeSpecClient()

	// when
	podSpecs, err := client.GetPodSpecs()

	// then
	assert.NoError(t, err)
	assert.Empty(t, podSpecs)
}

func TestGetPodSpecsReturnsSpecs(t *testing.T) {
	// given
	tc := newSpecClientTestCase()
	client := tc.createFakeSpecClient()

	// when
	podSpecs, err := client.GetPodSpecs()

	//then
	assert.NoError(t, err)
	assert.Equal(t, len(tc.podSpecs), len(podSpecs), "SpecClient returned different number of results then expected")
	for _, podSpec := range podSpecs {
		assert.Contains(t, tc.podSpecs, podSpec, "One of returned BasicPodSpcec is different than expected")
	}
}
