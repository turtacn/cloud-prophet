// 维护各种维度到实例的Lister操作映射
//
package podtopologyspread

import (
	"fmt"

	"github.com/turtacn/cloud-prophet/scheduler/apis/config"
	"github.com/turtacn/cloud-prophet/scheduler/apis/config/validation"
	framework "github.com/turtacn/cloud-prophet/scheduler/framework/base"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// ErrReasonConstraintsNotMatch is used for PodTopologySpread filter error.
	ErrReasonConstraintsNotMatch = "node(s) didn't match pod topology spread constraints"
	// ErrReasonNodeLabelNotMatch is used when the node doesn't hold the required label.
	ErrReasonNodeLabelNotMatch = ErrReasonConstraintsNotMatch + " (missing required label)"
)

// PodTopologySpread is a plugin that ensures pod's topologySpreadConstraints is satisfied.
type PodTopologySpread struct {
	args         config.PodTopologySpreadArgs
	sharedLister framework.SharedLister
}

var _ framework.PreFilterPlugin = &PodTopologySpread{}
var _ framework.FilterPlugin = &PodTopologySpread{}
var _ framework.PreScorePlugin = &PodTopologySpread{}
var _ framework.ScorePlugin = &PodTopologySpread{}

const (
	// Name is the name of the plugin used in the plugin registry and configurations.
	Name = "PodTopologySpread"
)

// Name returns name of the plugin. It is used in logs, etc.
func (pl *PodTopologySpread) Name() string {
	return Name
}

// BuildArgs returns the arguments used to build the plugin.
func (pl *PodTopologySpread) BuildArgs() interface{} {
	return pl.args
}

// New initializes a new plugin and returns it.
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
