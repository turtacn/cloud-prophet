package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	epsilon = 0.001
)

// Test all methods of LinearHistogramOptions using a sample bucketing scheme.
func TestLinearHistogramOptions(t *testing.T) {
	o, err := NewLinearHistogramOptions(5.0, 0.3, epsilon)
	assert.Nil(t, err)
	assert.Equal(t, epsilon, o.Epsilon())
	assert.Equal(t, 18, o.NumBuckets())

	assert.Equal(t, 0.0, o.GetBucketStart(0))
	assert.Equal(t, 5.1, o.GetBucketStart(17))

	assert.Equal(t, 0, o.FindBucket(-1.0))
	assert.Equal(t, 0, o.FindBucket(0.0))
	assert.Equal(t, 4, o.FindBucket(1.3))
	assert.Equal(t, 17, o.FindBucket(100.0))
}

// Test all methods of ExponentialHistogramOptions using a sample bucketing scheme.
func TestExponentialHistogramOptions(t *testing.T) {
	o, err := NewExponentialHistogramOptions(500.0, 40.0, 1.5, epsilon)
	assert.Nil(t, err)
	assert.Equal(t, epsilon, o.Epsilon())
	assert.Equal(t, 6, o.NumBuckets())

	assert.Equal(t, 0.0, o.GetBucketStart(0))
	assert.Equal(t, 40.0, o.GetBucketStart(1))
	assert.Equal(t, 100.0, o.GetBucketStart(2))
	assert.Equal(t, 190.0, o.GetBucketStart(3))
	assert.Equal(t, 325.0, o.GetBucketStart(4))
	assert.Equal(t, 527.5, o.GetBucketStart(5))

	assert.Equal(t, 0, o.FindBucket(-1.0))
	assert.Equal(t, 0, o.FindBucket(39.99))
	assert.Equal(t, 1, o.FindBucket(40.0))
	assert.Equal(t, 2, o.FindBucket(100.0))
	assert.Equal(t, 5, o.FindBucket(900.0))
}
