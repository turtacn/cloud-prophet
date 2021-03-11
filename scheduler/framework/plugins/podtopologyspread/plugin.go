package podtopologyspread

import (
	"fmt"

	"github.com/turtacn/cloud-prophet/scheduler/apis/config"
	"github.com/turtacn/cloud-prophet/scheduler/apis/config/validation"
	framework "github.com/turtacn/cloud-prophet/scheduler/framework/base"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	ErrReasonConstraintsNotMatch = "node(s) didn't match pod topology spread constraints"
	ErrReasonNodeLabelNotMatch   = ErrReasonConstraintsNotMatch + " (missing required label)"
)

type PodTopologySpread struct {
	args         config.PodTopologySpreadArgs
	sharedLister framework.SharedLister
}

var _ framework.PreFilterPlugin = &PodTopologySpread{}
var _ framework.FilterPlugin = &PodTopologySpread{}
var _ framework.PreScorePlugin = &PodTopologySpread{}
var _ framework.ScorePlugin = &PodTopologySpread{}

const (
	Name = "PodTopologySpread"
)

func (pl *PodTopologySpread) Name() string {
	return Name
}

func (pl *PodTopologySpread) BuildArgs() interface{} {
	return pl.args
}

func New(plArgs runtime.Object, h framework.FrameworkHandle) (framework.Plugin, error) {
	if h.SnapshotSharedLister() == nil {
		return nil, fmt.Errorf("SnapshotSharedlister is nil")
	}
	args, err := getArgs(plArgs)
	if err != nil {
		return nil, err
	}
	if err := validation.ValidatePodTopologySpreadArgs(&args); err != nil {
		return nil, err
	}
	pl := &PodTopologySpread{
		sharedLister: h.SnapshotSharedLister(),
		args:         args,
	}
	if len(pl.args.DefaultConstraints) != 0 {
	}
	return pl, nil
}

func getArgs(obj runtime.Object) (config.PodTopologySpreadArgs, error) {
	ptr, ok := obj.(*config.PodTopologySpreadArgs)
	if !ok {
		return config.PodTopologySpreadArgs{}, fmt.Errorf("want args to be of type PodTopologySpreadArgs, got %T", obj)
	}
	return *ptr, nil
}
