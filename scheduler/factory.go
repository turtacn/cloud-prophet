//
//
package scheduler

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/turtacn/cloud-prophet/scheduler/algorithmprovider"
	schedulerapi "github.com/turtacn/cloud-prophet/scheduler/apis/config"
	"github.com/turtacn/cloud-prophet/scheduler/core"
	framework "github.com/turtacn/cloud-prophet/scheduler/framework/base"
	frameworkplugins "github.com/turtacn/cloud-prophet/scheduler/framework/plugins"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/noderesources"
	frameworkruntime "github.com/turtacn/cloud-prophet/scheduler/framework/runtime"
	"github.com/turtacn/cloud-prophet/scheduler/helper/sets"
	internalcache "github.com/turtacn/cloud-prophet/scheduler/internal/cache"
	internalqueue "github.com/turtacn/cloud-prophet/scheduler/internal/queue"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	"github.com/turtacn/cloud-prophet/scheduler/profile"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

// Binder knows how to write a binding.
type Binder interface {
	Bind(binding *v1.Binding) error
}

// Configurator defines I/O, caching, and other functionality needed to
// construct a new scheduler.
type Configurator struct {
	client framework.ClientSet

	informerFactory framework.SharedInformer

	podInformer framework.SharedPodsLister

	// Close this to stop all reflectors
	StopEverything <-chan struct{}

	schedulerCache internalcache.Cache

	// Disable pod preemption or not.
	disablePreemption bool

	// Always check all predicates even if the middle of one predicate fails.
	alwaysCheckAllPredicates bool

	// percentageOfNodesToScore specifies percentage of all nodes to score in each scheduling cycle.
	percentageOfNodesToScore int32

	podInitialBackoffSeconds int64

	podMaxBackoffSeconds int64

	profiles          []schedulerapi.KubeSchedulerProfile
	registry          frameworkruntime.Registry
	nodeInfoSnapshot  *internalcache.Snapshot
	extenders         []schedulerapi.Extender
	frameworkCapturer FrameworkCapturer
}

func (c *Configurator) buildFramework(p schedulerapi.KubeSchedulerProfile, opts ...frameworkruntime.Option) (framework.Framework, error) {
	if c.frameworkCapturer != nil {
		c.frameworkCapturer(p)
	}
	opts = append([]frameworkruntime.Option{
		frameworkruntime.WithClientSet(c.client),
		frameworkruntime.WithInformerFactory(c.informerFactory),
		frameworkruntime.WithSnapshotSharedLister(c.nodeInfoSnapshot),
		frameworkruntime.WithRunAllFilters(c.alwaysCheckAllPredicates),
	}, opts...)
	return frameworkruntime.NewFramework(
		c.registry,
		p.Plugins,
		p.PluginConfig,
		opts...,
	)
}

// create a scheduler from a set of registered plugins.
func (c *Configurator) create() (*Scheduler, error) {
	var extenders []framework.Extender
	var ignoredExtendedResources []string
	if len(c.extenders) != 0 {
		var ignorableExtenders []framework.Extender
		for ii := range c.extenders {
			klog.Infof("Creating extender with config %+v", c.extenders[ii])
			extender, err := core.NewHTTPExtender(&c.extenders[ii])
			if err != nil {
				return nil, err
			}
			if !extender.IsIgnorable() {
				extenders = append(extenders, extender)
			} else {
				ignorableExtenders = append(ignorableExtenders, extender)
			}
			for _, r := range c.extenders[ii].ManagedResources {
				if r.IgnoredByScheduler {
					ignoredExtendedResources = append(ignoredExtendedResources, r.Name)
				}
			}
		}
		// place ignorable extenders to the tail of extenders
		extenders = append(extenders, ignorableExtenders...)
	}

	// If there are any extended resources found from the Extenders, append them to the pluginConfig for each profile.
	// This should only have an effect on ComponentConfig v1beta1, where it is possible to configure Extenders and
	// plugin args (and in which case the extender ignored resources take precedence).
	// For earlier versions, using both policy and custom plugin config is disallowed, so this should be the only
	// plugin config for this plugin.
	if len(ignoredExtendedResources) > 0 {
		for i := range c.profiles {
			prof := &c.profiles[i]
			pc := schedulerapi.PluginConfig{
				Name: noderesources.FitName,
				Args: &schedulerapi.NodeResourcesFitArgs{
					IgnoredResources: ignoredExtendedResources,
				},
			}
			prof.PluginConfig = append(prof.PluginConfig, pc)
		}
	}

	// The nominator will be passed all the way to framework instantiation.
	nominator := internalqueue.NewPodNominator()
	profiles, err := profile.NewMap(c.profiles, c.buildFramework,
		frameworkruntime.WithPodNominator(nominator))
	if err != nil {
		klog.Errorf("initializing profiles: %v", err)
		return nil, fmt.Errorf("initializing profiles: %v", err)
	}
	if len(profiles) == 0 {
		klog.Errorf("at least one profile is required")
		return nil, errors.New("at least one profile is required")
	}
	// Profiles are required to have equivalent queue sort plugins.
	lessFn := profiles[c.profiles[0].SchedulerName].Framework.QueueSortFunc()
	podQueue := internalqueue.NewSchedulingQueue(
		lessFn,
		internalqueue.WithPodInitialBackoffDuration(time.Duration(c.podInitialBackoffSeconds)*time.Second),
		internalqueue.WithPodMaxBackoffDuration(time.Duration(c.podMaxBackoffSeconds)*time.Second),
		internalqueue.WithPodNominator(nominator),
	)

	algo := core.NewGenericScheduler(
		c.schedulerCache,
		c.nodeInfoSnapshot,
		extenders,
		c.disablePreemption,
		c.percentageOfNodesToScore,
	)

	return &Scheduler{
		SchedulerCache:  c.schedulerCache,
		Algorithm:       algo,
		Profiles:        profiles,
		NextPod:         internalqueue.MakeNextPodFunc(podQueue),
		Error:           MakeDefaultErrorFunc(c.client, c.podInformer, podQueue, c.schedulerCache),
		StopEverything:  c.StopEverything,
		SchedulingQueue: podQueue,
	}, nil
}

// createFromProvider creates a scheduler from the name of a registered algorithm provider.
func (c *Configurator) createFromProvider(providerName string) (*Scheduler, error) {
	klog.Infof("Creating scheduler from algorithm provider '%v'", providerName)
	r := algorithmprovider.NewRegistry()
	defaultPlugins, exist := r[providerName]
	if !exist {
		klog.Errorf("algorithm provider %q is not registered", providerName)
		return nil, fmt.Errorf("algorithm provider %q is not registered", providerName)
	}

	for i := range c.profiles {
		klog.Infof("meet %d-th(%d) profile", i, len(c.profiles))
		prof := &c.profiles[i]
		plugins := &schedulerapi.Plugins{}
		plugins.Append(defaultPlugins)
		plugins.Apply(prof.Plugins)
		prof.Plugins = plugins
	}
	return c.create()
}

// mergePluginConfigsFromPolicy merges the giving plugin configs ensuring that,
// if a plugin name is repeated, the arguments are the same.
func mergePluginConfigsFromPolicy(pc1, pc2 []schedulerapi.PluginConfig) ([]schedulerapi.PluginConfig, error) {
	args := make(map[string]runtime.Object)
	for _, c := range pc1 {
		args[c.Name] = c.Args
	}
	for _, c := range pc2 {
		if v, ok := args[c.Name]; ok && !cmp.Equal(v, c.Args) {
			// This should be unreachable.
			return nil, fmt.Errorf("inconsistent configuration produced for plugin %s", c.Name)
		}
		args[c.Name] = c.Args
	}
	pc := make([]schedulerapi.PluginConfig, 0, len(args))
	for k, v := range args {
		pc = append(pc, schedulerapi.PluginConfig{
			Name: k,
			Args: v,
		})
	}
	return pc, nil
}

// getPriorityConfigs returns priorities configuration: ones that will run as priorities and ones that will run
// as framework plugins. Specifically, a priority will run as a framework plugin if a plugin config producer was
// registered for that priority.
func getPriorityConfigs(keys map[string]int64, lr *frameworkplugins.LegacyRegistry, args *frameworkplugins.ConfigProducerArgs) (*schedulerapi.Plugins, []schedulerapi.PluginConfig, error) {
	var plugins schedulerapi.Plugins
	var pluginConfig []schedulerapi.PluginConfig

	// Sort the keys so that it is easier for unit tests to do compare.
	var sortedKeys []string
	for k := range keys {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for _, priority := range sortedKeys {
		weight := keys[priority]
		producer, exist := lr.PriorityToConfigProducer[priority]
		if !exist {
			return nil, nil, fmt.Errorf("no config producer registered for %q", priority)
		}
		a := *args
		a.Weight = int32(weight)
		pl, plc := producer(a)
		plugins.Append(&pl)
		pluginConfig = append(pluginConfig, plc...)
	}
	return &plugins, pluginConfig, nil
}

// getPredicateConfigs returns predicates configuration: ones that will run as fitPredicates and ones that will run
// as framework plugins. Specifically, a predicate will run as a framework plugin if a plugin config producer was
// registered for that predicate.
// Note that the framework executes plugins according to their order in the Plugins list, and so predicates run as plugins
// are added to the Plugins list according to the order specified in predicates.Ordering().
func getPredicateConfigs(keys sets.String, lr *frameworkplugins.LegacyRegistry, args *frameworkplugins.ConfigProducerArgs) (*schedulerapi.Plugins, []schedulerapi.PluginConfig, error) {
	allPredicates := keys.Union(lr.MandatoryPredicates)

	// Create the framework plugin configurations, and place them in the order
	// that the corresponding predicates were supposed to run.
	var plugins schedulerapi.Plugins
	var pluginConfig []schedulerapi.PluginConfig

	for _, predicateKey := range frameworkplugins.PredicateOrdering() {
		if allPredicates.Has(predicateKey) {
			producer, exist := lr.PredicateToConfigProducer[predicateKey]
			if !exist {
				return nil, nil, fmt.Errorf("no framework config producer registered for %q", predicateKey)
			}
			pl, plc := producer(*args)
			plugins.Append(&pl)
			pluginConfig = append(pluginConfig, plc...)
			allPredicates.Delete(predicateKey)
		}
	}

	// Third, add the rest in no specific order.
	for predicateKey := range allPredicates {
		producer, exist := lr.PredicateToConfigProducer[predicateKey]
		if !exist {
			return nil, nil, fmt.Errorf("no framework config producer registered for %q", predicateKey)
		}
		pl, plc := producer(*args)
		plugins.Append(&pl)
		pluginConfig = append(pluginConfig, plc...)
	}

	return &plugins, pluginConfig, nil
}

// MakeDefaultErrorFunc construct a function to handle pod scheduler error
func MakeDefaultErrorFunc(client framework.ClientSet, podInformer framework.SharedPodsLister, podQueue internalqueue.SchedulingQueue, schedulerCache internalcache.Cache) func(*framework.QueuedPodInfo, error) {
	return func(podInfo *framework.QueuedPodInfo, err error) {
		pod := podInfo.Pod
		if err == core.ErrNoNodesAvailable {
			klog.InfoS("Unable to schedule pod; no nodes are registered to the cluster; waiting", "pod", pod)
		} else if _, ok := err.(*core.FitError); ok {
			klog.InfoS("Unable to schedule pod; no fit; waiting", "pod", pod, "err", err)
		} else if apierrors.IsNotFound(err) {
			klog.Infof("Unable to schedule %v/%v: possibly due to node not found: %v; waiting", pod.Namespace, pod.Name, err)
			if errStatus, ok := err.(apierrors.APIStatus); ok && errStatus.Status().Details.Kind == "node" {
				nodeName := errStatus.Status().Details.Name
				// when node is not found, We do not remove the node right away. Trying again to get
				// the node and if the node is still not found, then remove it from the scheduler cache.
				if client == nil {

				}
				if err != nil && apierrors.IsNotFound(err) {
					node := v1.Node{ObjectMeta: v1.ObjectMeta{Name: nodeName}}
					if err := schedulerCache.RemoveNode(&node); err != nil {
						klog.Infof("Node %q is not found; failed to remove it from the cache.", node.Name)
					}
				}
			}
		} else {
			klog.ErrorS(err, "Error scheduling pod; retrying", "pod", pod)
		}

		// Check if the Pod exists in informer cache.
		cachedPod, err := podInformer.PodInfos().Get(pod.Name)
		if err != nil {
			klog.Warningf("Pod %v/%v doesn't exist in informer cache: %v", pod.Namespace, pod.Name, err)
			return
		}
		// As <cachedPod> is from SharedInformer, we need to do a DeepCopy() here.
		if cachedPod == nil {
			podInfo.Pod = nil

		}
		if err := podQueue.AddUnschedulableIfNotPresent(podInfo, podQueue.SchedulingCycle()); err != nil {
			klog.Error(err)
		}
	}
}
