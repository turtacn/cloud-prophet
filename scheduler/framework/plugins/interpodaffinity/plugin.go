package interpodaffinity

import (
	"fmt"

	"github.com/turtacn/cloud-prophet/scheduler/apis/config"
	"github.com/turtacn/cloud-prophet/scheduler/apis/config/validation"
	framework "github.com/turtacn/cloud-prophet/scheduler/framework/base"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	Name = "InterPodAffinity"
)

var _ framework.PreFilterPlugin = &InterPodAffinity{}
var _ framework.FilterPlugin = &InterPodAffinity{}
var _ framework.PreScorePlugin = &InterPodAffinity{}
var _ framework.ScorePlugin = &InterPodAffinity{}

type InterPodAffinity struct {
	args         config.InterPodAffinityArgs
	sharedLister framework.SharedLister
}

func (pl *InterPodAffinity) Name() string {
	return Name
}

func (pl *InterPodAffinity) BuildArgs() interface{} {
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
	if err := validation.ValidateInterPodAffinityArgs(args); err != nil {
		return nil, err
	}
	return &InterPodAffinity{
		args:         args,
		sharedLister: h.SnapshotSharedLister(),
	}, nil
}

func getArgs(obj runtime.Object) (config.InterPodAffinityArgs, error) {
	ptr, ok := obj.(*config.InterPodAffinityArgs)
	if !ok {
		return config.InterPodAffinityArgs{}, fmt.Errorf("want args to be of type InterPodAffinityArgs, got %T", obj)
	}
	return *ptr, nil
}
