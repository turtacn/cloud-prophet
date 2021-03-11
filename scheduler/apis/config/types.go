package config

import (
	"github.com/turtacn/cloud-prophet/scheduler/helper/sets"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"math"
	"time"
)

const (
	SchedulerDefaultLockObjectNamespace string = "kube-system"

	SchedulerDefaultLockObjectName = "kube-scheduler"

	SchedulerPolicyConfigMapKey = "policy.cfg"

	SchedulerDefaultProviderName = "DefaultProvider"
)

type KubeSchedulerProfile struct {
	SchedulerName string

	Plugins *Plugins

	PluginConfig []PluginConfig
}

type SchedulerAlgorithmSource struct {
	Provider *string
}

type Plugins struct {
	QueueSort *PluginSet

	PreFilter *PluginSet

	Filter *PluginSet

	PostFilter *PluginSet

	PreScore *PluginSet

	Score *PluginSet

	Reserve *PluginSet

	Permit *PluginSet

	PreBind *PluginSet

	Bind *PluginSet

	PostBind *PluginSet
}

type PluginSet struct {
	Enabled  []Plugin
	Disabled []Plugin
}

type Plugin struct {
	Name   string
	Weight int32
}

type PluginConfig struct {
	Name string
	Args runtime.Object
}

/*
 * NOTE: The following variables and methods are intentionally left out of the staging mirror.
 */
const (
	DefaultPercentageOfNodesToScore = 0

	MaxCustomPriorityScore int64 = 10

	MaxTotalScore int64 = math.MaxInt64

	MaxWeight = MaxTotalScore / MaxCustomPriorityScore

	MaxDuration = (1<<32 - 1) * time.Second
)

func appendPluginSet(dst *PluginSet, src *PluginSet) *PluginSet {
	if dst == nil {
		dst = &PluginSet{}
	}
	if src != nil {
		dst.Enabled = append(dst.Enabled, src.Enabled...)
		dst.Disabled = append(dst.Disabled, src.Disabled...)
	}
	return dst
}

func (p *Plugins) Append(src *Plugins) {
	if p == nil || src == nil {
		return
	}
	p.QueueSort = appendPluginSet(p.QueueSort, src.QueueSort)
	p.PreFilter = appendPluginSet(p.PreFilter, src.PreFilter)
	p.Filter = appendPluginSet(p.Filter, src.Filter)
	p.PostFilter = appendPluginSet(p.PostFilter, src.PostFilter)
	p.PreScore = appendPluginSet(p.PreScore, src.PreScore)
	p.Score = appendPluginSet(p.Score, src.Score)
	p.Reserve = appendPluginSet(p.Reserve, src.Reserve)
	p.Permit = appendPluginSet(p.Permit, src.Permit)
	p.PreBind = appendPluginSet(p.PreBind, src.PreBind)
	p.Bind = appendPluginSet(p.Bind, src.Bind)
	p.PostBind = appendPluginSet(p.PostBind, src.PostBind)
}

func (p *Plugins) Apply(customPlugins *Plugins) {
	if customPlugins == nil {
		return
	}

	p.QueueSort = mergePluginSets(p.QueueSort, customPlugins.QueueSort)
	p.PreFilter = mergePluginSets(p.PreFilter, customPlugins.PreFilter)
	p.Filter = mergePluginSets(p.Filter, customPlugins.Filter)
	p.PostFilter = mergePluginSets(p.PostFilter, customPlugins.PostFilter)
	p.PreScore = mergePluginSets(p.PreScore, customPlugins.PreScore)
	p.Score = mergePluginSets(p.Score, customPlugins.Score)
	p.Reserve = mergePluginSets(p.Reserve, customPlugins.Reserve)
	p.Permit = mergePluginSets(p.Permit, customPlugins.Permit)
	p.PreBind = mergePluginSets(p.PreBind, customPlugins.PreBind)
	p.Bind = mergePluginSets(p.Bind, customPlugins.Bind)
	p.PostBind = mergePluginSets(p.PostBind, customPlugins.PostBind)
}

func mergePluginSets(defaultPluginSet, customPluginSet *PluginSet) *PluginSet {
	if customPluginSet == nil {
		customPluginSet = &PluginSet{}
	}

	if defaultPluginSet == nil {
		defaultPluginSet = &PluginSet{}
	}

	disabledPlugins := sets.NewString()
	for _, disabledPlugin := range customPluginSet.Disabled {
		disabledPlugins.Insert(disabledPlugin.Name)
	}

	enabledPlugins := []Plugin{}
	if !disabledPlugins.Has("*") {
		for _, defaultEnabledPlugin := range defaultPluginSet.Enabled {
			if disabledPlugins.Has(defaultEnabledPlugin.Name) {
				continue
			}

			enabledPlugins = append(enabledPlugins, defaultEnabledPlugin)
		}
	}

	enabledPlugins = append(enabledPlugins, customPluginSet.Enabled...)

	return &PluginSet{Enabled: enabledPlugins}
}

type Extender struct {
	URLPrefix        string
	FilterVerb       string
	PreemptVerb      string
	PrioritizeVerb   string
	Weight           int64
	BindVerb         string
	EnableHTTPS      bool
	TLSConfig        *ExtenderTLSConfig
	HTTPTimeout      metav1.Duration
	NodeCacheCapable bool
	ManagedResources []ExtenderManagedResource
	Ignorable        bool
}

type ExtenderManagedResource struct {
	Name               string
	IgnoredByScheduler bool
}

type ExtenderTLSConfig struct {
	Insecure   bool
	ServerName string

	CertFile string
	KeyFile  string
	CAFile   string

	CertData []byte
	KeyData  []byte
	CAData   []byte
}
