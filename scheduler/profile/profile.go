package profile

import (
	"errors"
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/turtacn/cloud-prophet/scheduler/apis/config"
	framework "github.com/turtacn/cloud-prophet/scheduler/framework/base"
	frameworkruntime "github.com/turtacn/cloud-prophet/scheduler/framework/runtime"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
)

type FrameworkFactory func(config.KubeSchedulerProfile, ...frameworkruntime.Option) (framework.Framework, error)

type Profile struct {
	framework.Framework
	Name string
}

func NewProfile(cfg config.KubeSchedulerProfile, frameworkFact FrameworkFactory,
	opts ...frameworkruntime.Option) (*Profile, error) {
	opts = append(opts, frameworkruntime.WithProfileName(cfg.SchedulerName))
	fwk, err := frameworkFact(cfg, opts...)
	if err != nil {
		klog.Errorf("famework factory for scheduler %s create error %v", cfg.SchedulerName, err)
		return nil, err
	}
	return &Profile{
		Name:      cfg.SchedulerName,
		Framework: fwk,
	}, nil
}

type Map map[string]*Profile

func NewMap(cfgs []config.KubeSchedulerProfile, frameworkFact FrameworkFactory,
	opts ...frameworkruntime.Option) (Map, error) {
	m := make(Map)
	v := cfgValidator{m: m}

	for _, cfg := range cfgs {
		if err := v.validate(cfg); err != nil {
			klog.Errorf("validate kube schedule profile config error=%v", err)
			return nil, err
		}
		p, err := NewProfile(cfg, frameworkFact, opts...)
		if err != nil {
			klog.Errorf("creating profile for scheduler name %s: %v", cfg.SchedulerName, err)
			return nil, fmt.Errorf("creating profile for scheduler name %s: %v", cfg.SchedulerName, err)
		}
		m[cfg.SchedulerName] = p
	}
	return m, nil
}

func (m Map) HandlesSchedulerName(name string) bool {
	_, ok := m[name]
	return ok
}

type cfgValidator struct {
	m             Map
	queueSort     string
	queueSortArgs runtime.Object
}

func (v *cfgValidator) validate(cfg config.KubeSchedulerProfile) error {
	if len(cfg.SchedulerName) == 0 {
		return errors.New("scheduler name is needed")
	}
	if cfg.Plugins == nil {
		return fmt.Errorf("plugins required for profile with scheduler name %q", cfg.SchedulerName)
	}
	if v.m[cfg.SchedulerName] != nil {
		return fmt.Errorf("duplicate profile with scheduler name %q", cfg.SchedulerName)
	}
	if cfg.Plugins.QueueSort == nil || len(cfg.Plugins.QueueSort.Enabled) != 1 {
		return fmt.Errorf("one queue sort plugin required for profile with scheduler name %q", cfg.SchedulerName)
	}
	queueSort := cfg.Plugins.QueueSort.Enabled[0].Name
	var queueSortArgs runtime.Object
	for _, plCfg := range cfg.PluginConfig {
		if plCfg.Name == queueSort {
			queueSortArgs = plCfg.Args
		}
	}
	if len(v.queueSort) == 0 {
		v.queueSort = queueSort
		v.queueSortArgs = queueSortArgs
		return nil
	}
	if v.queueSort != queueSort {
		return fmt.Errorf("different queue sort plugins for profile %q: %q, first: %q", cfg.SchedulerName, queueSort, v.queueSort)
	}
	if !cmp.Equal(v.queueSortArgs, queueSortArgs) {
		return fmt.Errorf("different queue sort plugin args for profile %q", cfg.SchedulerName)
	}
	return nil
}
