package scheduler

import (
	"errors"
	"fmt"
	"github.com/turtacn/cloud-prophet/scheduler/algorithmprovider"
	schedulerapi "github.com/turtacn/cloud-prophet/scheduler/apis/config"
	"github.com/turtacn/cloud-prophet/scheduler/core"
	framework "github.com/turtacn/cloud-prophet/scheduler/framework/base"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/noderesources"
	frameworkruntime "github.com/turtacn/cloud-prophet/scheduler/framework/runtime"
	internalcache "github.com/turtacn/cloud-prophet/scheduler/internal/cache"
	internalqueue "github.com/turtacn/cloud-prophet/scheduler/internal/queue"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	"github.com/turtacn/cloud-prophet/scheduler/profile"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/klog/v2"
	"time"
)

type Binder interface {
	Bind(binding *v1.Binding) error
}

type Configurator struct {
	client framework.ClientSet

	informerFactory framework.SharedInformer

	podInformer framework.SharedPodsLister

	StopEverything <-chan struct{}

	schedulerCache internalcache.Cache

	alwaysCheckAllPredicates bool

	percentageOfNodesToScore int32

	podInitialBackoffSeconds int64

	podMaxBackoffSeconds int64

	hostPrintedScheduleTrace bool

	profiles          []schedulerapi.KubeSchedulerProfile
	registry          frameworkruntime.Registry
	nodeInfoSnapshot  *internalcache.Snapshot
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

func (c *Configurator) create() (*Scheduler, error) {
	var extenders []framework.Extender
	var ignoredExtendedResources []string
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
	lessFn := profiles[c.profiles[0].SchedulerName].Framework.QueueSortFunc()
	podQueue := internalqueue.NewSchedulingQueue(
		lessFn,
		internalqueue.WithPodInitialBackoffDuration(time.Duration(c.podInitialBackoffSeconds)*time.Second), // pod 初始化补偿时间间隔
		internalqueue.WithPodMaxBackoffDuration(time.Duration(c.podMaxBackoffSeconds)*time.Second),         // pod 最大补偿时间间隔
		internalqueue.WithPodNominator(nominator),
	)

	algo := core.NewGenericScheduler(
		c.schedulerCache,
		c.nodeInfoSnapshot,
		extenders,
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

		if podInformer != nil {
			cachedPod, err := podInformer.PodInfos().Get(pod.Name)
			if err != nil {
				klog.Warningf("Pod %v/%v doesn't exist in informer cache: %v", pod.Namespace, pod.Name, err)
				return
			}
			if cachedPod == nil {
				podInfo.Pod = nil

			}
		}
		if err := podQueue.AddUnschedulableIfNotPresent(podInfo, podQueue.SchedulingCycle()); err != nil {
			klog.Error(err)
		}
	}
}
