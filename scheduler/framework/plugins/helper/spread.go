//
//
package helper

import (
	"github.com/turtacn/cloud-prophet/scheduler/framework/base"
	labels "github.com/turtacn/cloud-prophet/scheduler/helper/label"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
)

// DefaultSelector returns a selector deduced from the Services, Replication
// Controllers, Replica Sets, and Stateful Sets matching the given pod.
func DefaultSelector(pod *v1.Pod, listers ...base.SharedLister) labels.Selector {
	labelSet := make(labels.Set)
	// Since services, RCs, RSs and SSs match the pod, they won't have conflicting
	// labels. Merging is safe.

	selector := labels.NewSelector()
	if len(labelSet) != 0 {
		selector = labelSet.AsSelector()
	}

	return selector
}
