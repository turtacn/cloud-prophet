package util

import (
	"fmt"
	"math"
	"time"
)

var (
	// When the decay factor exceeds 2^maxDecayExponent the histogram is
	// renormalized by shifting the decay start time forward.
	maxDecayExponent = 100
)

// A histogram that gives newer samples a higher weight than the old samples,
// gradually decaying ("forgetting") the past samples. The weight of each sample
// is multiplied by the factor of 2^((sampleTime - referenceTimestamp) / halfLife).
// This means that the sample loses half of its weight ("importance") with
// each halfLife period.
// Since only relative (and not absolute) weights of samples matter, the
// referenceTimestamp can be shifted at any time, which is equivalent to multiplying all
// weights by a constant. In practice the referenceTimestamp is shifted forward whenever
// the exponents become too large, to avoid floating point arithmetics overflow.
type decayingHistogram struct {
	histogram
	// Decay half life period.
	halfLife time.Duration
	// Reference time for determining the relative age of samples.
	// It is always an integer multiple of halfLife.
	referenceTimestamp time.Time
}

// NewDecayingHistogram returns a new DecayingHistogram instance using given options.
func NewDecayingHistogram(options HistogramOptions, halfLife time.Duration) Histogram {
	return &decayingHistogram{
		histogram:          *NewHistogram(options).(*histogram),
		halfLife:           halfLife,
		referenceTimestamp: time.Time{},
	}
}

func (h *decayingHistogram) Percentile(percentile float64) float64 {
	return h.histogram.Percentile(percentile)
}

func (h *decayingHistogram) AddSample(value float64, weight float64, time time.Time) {
	h.histogram.AddSample(value, weight*h.decayFactor(time), time)
}

func (h *decayingHistogram) SubtractSample(value float64, weight float64, time time.Time) {
	h.histogram.SubtractSample(value, weight*h.decayFactor(time), time)
}

func (h *decayingHistogram) Merge(other Histogram) {
	o := other.(*decayingHistogram)
	if h.halfLife != o.halfLife {
		panic("can't merge decaying histograms with different half life periods")
	}
	// Align the older referenceTimestamp with the younger one.
	if h.referenceTimestamp.Before(o.referenceTimestamp) {
		h.shiftReferenceTimestamp(o.referenceTimestamp)
	} else if o.referenceTimestamp.Before(h.referenceTimestamp) {
		o.shiftReferenceTimestamp(h.referenceTimestamp)
	}
	h.histogram.Merge(&o.histogram)
}

func (h *decayingHistogram) Equals(other Histogram) bool {
	h2, typesMatch := (other).(*decayingHistogram)
	return typesMatch && h.halfLife == h2.halfLife && h.referenceTimestamp == h2.referenceTimestamp && h.histogram.Equals(&h2.histogram)
}

func (h *decayingHistogram) IsEmpty() bool {
	return h.histogram.IsEmpty()
}

func (h *decayingHistogram) String() string {
	return fmt.Sprintf("referenceTimestamp: %v, halfLife: %v\n%s", h.referenceTimestamp, h.halfLife, h.histogram.String())
}

func (h *decayingHistogram) shiftReferenceTimestamp(newreferenceTimestamp time.Time) {
	// Make sure the decay start is an integer multiple of halfLife.
	newreferenceTimestamp = newreferenceTimestamp.Round(h.halfLife)
	exponent := round(float64(h.referenceTimestamp.Sub(newreferenceTimestamp)) / float64(h.halfLife))
	h.histogram.scale(math.Ldexp(1., exponent)) // Scale all weights by 2^exponent.
	h.referenceTimestamp = newreferenceTimestamp
}

func (h *decayingHistogram) decayFactor(timestamp time.Time) float64 {
	// Max timestamp before the exponent grows too large.
	maxAllowedTimestamp := h.referenceTimestamp.Add(
		time.Duration(int64(h.halfLife) * int64(maxDecayExponent)))
	if timestamp.After(maxAllowedTimestamp) {
		// The exponent has grown too large. Renormalize the histogram by
		// shifting the referenceTimestamp to the current timestamp and rescaling
		// the weights accordingly.
		h.shiftReferenceTimestamp(timestamp)
	}
	return math.Exp2(float64(timestamp.Sub(h.referenceTimestamp)) / float64(h.halfLife))
}

func round(x float64) int {
	return int(math.Floor(x + 0.5))
}
