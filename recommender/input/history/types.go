package history

import (
	"time"
)

// Sample is a single timestamped value of the metric.
type Sample struct {
	Value     float64
	Timestamp time.Time
}

// Timeseries represents a metric with given labels, with its values possibly changing in time.
type Timeseries struct {
	Labels  map[string]string
	Samples []Sample
}
