package helper

import (
	"github.com/turtacn/cloud-prophet/scheduler/framework/base"
	labels "github.com/turtacn/cloud-prophet/scheduler/helper/label"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
)

func DefaultSelector(pod *v1.Pod, listers ...base.SharedLister) labels.Selector {
	labelSet := make(labels.Set)

	selector := labels.NewSelector()
	if len(labelSet) != 0 {
		selector = labelSet.AsSelector()
	}

	return selector
}
