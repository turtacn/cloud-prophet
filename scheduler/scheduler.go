//
package scheduler

import (
	"context"
	"fmt"
	"strconv"
	"time"

	schedulerapi "github.com/turtacn/cloud-prophet/scheduler/apis/config"
	"github.com/turtacn/cloud-prophet/scheduler/core"
	framework "github.com/turtacn/cloud-prophet/scheduler/framework/base"
	frameworkplugins "github.com/turtacn/cloud-prophet/scheduler/framework/plugins"
	frameworkruntime "github.com/turtacn/cloud-prophet/scheduler/framework/runtime"
	podutil "github.com/turtacn/cloud-prophet/scheduler/helper/pod"
	internalcache "github.com/turtacn/cloud-prophet/scheduler/internal/cache"
	internalqueue "github.com/turtacn/cloud-prophet/scheduler/internal/queue"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	"github.com/turtacn/cloud-prophet/scheduler/profile"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

const (
	SchedulerError = "SchedulerError"
)

type Scheduler struct {
	SchedulerCache internalcache.Cache

	Algorithm core.ScheduleAlgorithm

	NextPod func() *framework.QueuedPodInfo

	Error func(*framework.QueuedPodInfo, error)

	StopEverything <-chan struct{}

	SchedulingQueue internalqueue.SchedulingQueue

	Profiles profile.Map

	scheduledPodsHasSynced func() bool

	client framework.ClientSet
}

func (sched *Scheduler) Cache() internalcache.Cache {
	return sched.SchedulerCache
}

type schedulerOptions struct {
	schedulerAlgorithmSource schedulerapi.SchedulerAlgorithmSource
	percentageOfNodesToScore int32
	podInitialBackoffSeconds int64
	podMaxBackoffSeconds     int64
	printHostScore           bool

	frameworkOutOfTreeRegistry frameworkruntime.Registry
	profiles                   []schedulerapi.KubeSchedulerProfile
	extenders                  []schedulerapi.Extender
	frameworkCapturer          FrameworkCapturer
}

type Option func(*schedulerOptions)

func WithProfiles(p ...schedulerapi.KubeSchedulerProfile) Option {
	return func(o *schedulerOptions) {
		o.profiles = p
	}
}

func WithAlgorithmSource(source schedulerapi.SchedulerAlgorithmSource) Option {
	return func(o *schedulerOptions) {
		o.schedulerAlgorithmSource = source
	}
}

func WithPrintHostScheduleDetail(printHostScore bool) Option {
	return func(o *schedulerOptions) {
		o.printHostScore = printHostScore
	}
}

func WithPercentageOfNodesToScore(percentageOfNodesToScore int32) Option {
	return func(o *schedulerOptions) {
		o.percentageOfNodesToScore = percentageOfNodesToScore
	}
}

func WithFrameworkOutOfTreeRegistry(registry frameworkruntime.Registry) Option {
	return func(o *schedulerOptions) {
		o.frameworkOutOfTreeRegistry = registry
	}
}

func WithPodInitialBackoffSeconds(podInitialBackoffSeconds int64) Option {
	return func(o *schedulerOptions) {
		o.podInitialBackoffSeconds = podInitialBackoffSeconds
	}
}

func WithPodMaxBackoffSeconds(podMaxBackoffSeconds int64) Option {
	return func(o *schedulerOptions) {
		o.podMaxBackoffSeconds = podMaxBackoffSeconds
	}
}

type FrameworkCapturer func(schedulerapi.KubeSchedulerProfile)

func WithBuildFrameworkCapturer(fc FrameworkCapturer) Option {
	return func(o *schedulerOptions) {
		o.frameworkCapturer = fc
	}
}

var defaultSchedulerOptions = schedulerOptions{
	profiles: []schedulerapi.KubeSchedulerProfile{
		{SchedulerName: v1.DefaultSchedulerName},
	},
	schedulerAlgorithmSource: schedulerapi.SchedulerAlgorithmSource{
		Provider: defaultAlgorithmSourceProviderName(),
	},
	percentageOfNodesToScore: schedulerapi.DefaultPercentageOfNodesToScore,
	podInitialBackoffSeconds: int64(internalqueue.DefaultPodInitialBackoffDuration.Seconds()),
	podMaxBackoffSeconds:     int64(internalqueue.DefaultPodMaxBackoffDuration.Seconds()),
}

func New(client framework.ClientSet,
	informerFactory framework.SharedInformer,
	podInformer framework.SharedInformer,
	stopCh <-chan struct{},
	opts ...Option) (*Scheduler, error) {

	stopEverything := stopCh
	if stopEverything == nil {
		stopEverything = wait.NeverStop
	}

	options := defaultSchedulerOptions
	for _, opt := range opts {
		opt(&options)
	}

	schedulerCache := internalcache.New(schedulerapi.MaxDuration, stopEverything)

	registry := frameworkplugins.NewInTreeRegistry()
	if err := registry.Merge(options.frameworkOutOfTreeRegistry); err != nil {
		klog.Errorf("registry merge out of tree failed, error=%v", err)
		return nil, err
	}

	snapshot := internalcache.NewEmptySnapshot()

	configurator := &Configurator{
		client:                   client,
		schedulerCache:           schedulerCache,
		StopEverything:           stopEverything,
		percentageOfNodesToScore: options.percentageOfNodesToScore,
		podInitialBackoffSeconds: options.podInitialBackoffSeconds,
		podMaxBackoffSeconds:     options.podMaxBackoffSeconds,
		hostPrintedScheduleTrace: options.printHostScore,
		profiles:                 append([]schedulerapi.KubeSchedulerProfile(nil), options.profiles...),
		registry:                 registry,
		nodeInfoSnapshot:         snapshot,
		frameworkCapturer:        options.frameworkCapturer,
	}
	if podInformer != nil {
		configurator.podInformer = podInformer.PodLister()
	}
	if informerFactory != nil {
		configurator.informerFactory = informerFactory
	}

	var sched *Scheduler
	source := options.schedulerAlgorithmSource
	switch {
	case source.Provider != nil:
		sc, err := configurator.createFromProvider(*source.Provider)
		if err != nil {
			return nil, fmt.Errorf("couldn't create scheduler using provider %q: %v", *source.Provider, err)
		}
		sched = sc
	default:
		return nil, fmt.Errorf("unsupported algorithm source: %v", source)
	}
	sched.StopEverything = stopEverything
	sched.client = client

	if podInformer != nil || informerFactory != nil {
		sched.scheduledPodsHasSynced = podInformer.HasSynced
		addAllEventHandlers(sched, informerFactory, podInformer)
	}
	return sched, nil
}

func (sched *Scheduler) Run(ctx context.Context) {
	sched.SchedulingQueue.Run()
	klog.Infof("scheduling queue was ready for in-coming pod.")
	wait.UntilWithContext(ctx, sched.scheduleOne, 0)
	sched.SchedulingQueue.Close()
	klog.Infof("scheduling queue was closed.")
}

func (sched *Scheduler) recordSchedulingFailure(prof *profile.Profile, podInfo *framework.QueuedPodInfo, err error, reason string, nominatedNode string) {
	sched.Error(podInfo, err)

	if sched.SchedulingQueue != nil {
		sched.SchedulingQueue.AddNominatedPod(podInfo.Pod, nominatedNode)
	}

	pod := podInfo.Pod
	if err := updatePod(sched.client, pod, &v1.PodCondition{
		Type:    v1.PodScheduled,
		Status:  v1.ConditionFalse,
		Reason:  reason,
		Message: err.Error(),
	}, nominatedNode); err != nil {
		klog.Errorf("Error updating pod %s/%s: %v", pod.Namespace, pod.Name, err)
	}
}

func updatePod(client framework.ClientSet, pod *v1.Pod, condition *v1.PodCondition, nominatedNode string) error {
	klog.Infof("Updating pod condition for %s/%s to (%s==%s, Reason=%s)", pod.Namespace, pod.Name, condition.Type, condition.Status, condition.Reason)
	podCopy := pod.DeepCopy()
	if !podutil.UpdatePodCondition(&podCopy.Status, condition) &&
		(len(nominatedNode) == 0 || pod.Status.NominatedNodeName == nominatedNode) {
		return nil
	}
	if nominatedNode != "" {
		podCopy.Status.NominatedNodeName = nominatedNode
	}
	return nil
}

func (sched *Scheduler) assume(assumed *v1.Pod, host string) error {
	assumed.Spec.NodeName = host

	if err := sched.SchedulerCache.AssumePod(assumed); err != nil {
		klog.Errorf("scheduler cache AssumePod failed: %v", err)
		return err
	}
	if sched.SchedulingQueue != nil {
		sched.SchedulingQueue.DeleteNominatedPodIfExists(assumed)
	}

	return nil
}

func (sched *Scheduler) bind(ctx context.Context, prof *profile.Profile, assumed *v1.Pod, targetNode string, state *framework.CycleState) (err error) {
	start := time.Now()
	defer func() {
		sched.finishBinding(prof, assumed, targetNode, start, err)
	}()

	bindStatus := prof.RunBindPlugins(ctx, state, assumed, targetNode)
	if bindStatus.IsSuccess() {
		return nil
	}
	if bindStatus.Code() == framework.Error {
		return bindStatus.AsError()
	}
	return fmt.Errorf("bind status: %s, %v", bindStatus.Code().String(), bindStatus.Message())
}

func (sched *Scheduler) finishBinding(prof *profile.Profile, assumed *v1.Pod, targetNode string, start time.Time, err error) {
	if finErr := sched.SchedulerCache.FinishBinding(assumed); finErr != nil {
		klog.Errorf("scheduler cache FinishBinding failed: %v", finErr)
	}
	if err != nil {
		klog.Infof("Failed to bind pod: %v/%v", assumed.Namespace, assumed.Name)
		return
	}
}

func (sched *Scheduler) scheduleOne(ctx context.Context) {
	klog.Infof("prepare next pod for scheduler")
	podInfo := sched.NextPod()
	if podInfo == nil || podInfo.Pod == nil {
		return
	}
	pod := podInfo.Pod
	prof, err := sched.profileForPod(pod)
	if err != nil {
		klog.Error(err)
		return
	}
	if sched.skipPodSchedule(prof, pod) {
		return
	}

	klog.Infof("Attempting to schedule pod: %v/%v", pod.Namespace, pod.Name)

	state := framework.NewCycleState()
	schedulingCycleCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	scheduleResult, err := sched.Algorithm.Schedule(schedulingCycleCtx, prof, state, pod)
	if err != nil {
		nominatedNode := ""
		if fitError, ok := err.(*core.FitError); ok {
			if !prof.HasPostFilterPlugins() {
				klog.Infof("No PostFilter plugins are registered, so no preemption will be performed.")
			} else {
				result, status := prof.RunPostFilterPlugins(ctx, state, pod, fitError.FilteredNodesStatuses)
				if status.Code() == framework.Error {
					klog.Errorf("Status after running PostFilter plugins for pod %v/%v: %v", pod.Namespace, pod.Name, status)
				} else {
					klog.Infof("Status after running PostFilter plugins for pod %v/%v: %v", pod.Namespace, pod.Name, status)
				}
				if status.IsSuccess() && result != nil {
					nominatedNode = result.NominatedNodeName
				}
			}
		} else if err == core.ErrNoNodesAvailable {
			klog.Infof("No node available for pod %v error=%v", pod, err)
		} else {
			klog.ErrorS(err, "Error selecting node for pod", "pod", pod)
		}
		sched.recordSchedulingFailure(prof, podInfo, err, v1.PodReasonUnschedulable, nominatedNode)
		return
	}
	assumedPodInfo := podInfo.DeepCopy()
	assumedPod := assumedPodInfo.Pod
	err = sched.assume(assumedPod, scheduleResult.SuggestedHost)
	if err != nil {
		sched.recordSchedulingFailure(prof, assumedPodInfo, err, SchedulerError, "")
		return
	}

	if sts := prof.RunReservePluginsReserve(schedulingCycleCtx, state, assumedPod, scheduleResult.SuggestedHost); !sts.IsSuccess() {
		prof.RunReservePluginsUnreserve(schedulingCycleCtx, state, assumedPod, scheduleResult.SuggestedHost)
		if forgetErr := sched.Cache().ForgetPod(assumedPod); forgetErr != nil {
			klog.Errorf("scheduler cache ForgetPod failed: %v", forgetErr)
		}
		sched.recordSchedulingFailure(prof, assumedPodInfo, sts.AsError(), SchedulerError, "")
		return
	}

	runPermitStatus := prof.RunPermitPlugins(schedulingCycleCtx, state, assumedPod, scheduleResult.SuggestedHost)
	if runPermitStatus.Code() != framework.Wait && !runPermitStatus.IsSuccess() {
		var reason string
		if runPermitStatus.IsUnschedulable() {
			reason = v1.PodReasonUnschedulable
		} else {
			reason = SchedulerError
		}
		prof.RunReservePluginsUnreserve(schedulingCycleCtx, state, assumedPod, scheduleResult.SuggestedHost)
		if forgetErr := sched.Cache().ForgetPod(assumedPod); forgetErr != nil {
			klog.Errorf("scheduler cache ForgetPod failed: %v", forgetErr)
		}
		sched.recordSchedulingFailure(prof, assumedPodInfo, runPermitStatus.AsError(), reason, "")
		return
	}

	go func() {
		bindingCycleCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		waitOnPermitStatus := prof.WaitOnPermit(bindingCycleCtx, assumedPod)
		if !waitOnPermitStatus.IsSuccess() {
			var reason string
			if waitOnPermitStatus.IsUnschedulable() {
				reason = v1.PodReasonUnschedulable
			} else {
				reason = SchedulerError
			}
			prof.RunReservePluginsUnreserve(bindingCycleCtx, state, assumedPod, scheduleResult.SuggestedHost)
			if forgetErr := sched.Cache().ForgetPod(assumedPod); forgetErr != nil {
				klog.Errorf("scheduler cache ForgetPod failed: %v", forgetErr)
			}
			sched.recordSchedulingFailure(prof, assumedPodInfo, waitOnPermitStatus.AsError(), reason, "")
			return
		}

		preBindStatus := prof.RunPreBindPlugins(bindingCycleCtx, state, assumedPod, scheduleResult.SuggestedHost)
		if !preBindStatus.IsSuccess() {
			prof.RunReservePluginsUnreserve(bindingCycleCtx, state, assumedPod, scheduleResult.SuggestedHost)
			if forgetErr := sched.Cache().ForgetPod(assumedPod); forgetErr != nil {
				klog.Errorf("scheduler cache ForgetPod failed: %v", forgetErr)
			}
			sched.recordSchedulingFailure(prof, assumedPodInfo, preBindStatus.AsError(), SchedulerError, "")
			return
		}

		err := sched.bind(bindingCycleCtx, prof, assumedPod, scheduleResult.SuggestedHost, state)
		if err != nil {
			prof.RunReservePluginsUnreserve(bindingCycleCtx, state, assumedPod, scheduleResult.SuggestedHost)
			if err := sched.SchedulerCache.ForgetPod(assumedPod); err != nil {
				klog.Errorf("scheduler cache ForgetPod failed: %v", err)
			}
			sched.recordSchedulingFailure(prof, assumedPodInfo, fmt.Errorf("Binding rejected: %v", err), SchedulerError, "")
		} else {
			if true {
				klog.InfoS("Successfully bound pod to node", "pod", pod.Name, "node", scheduleResult.SuggestedHost, "evaluatedNodes", scheduleResult.EvaluatedNodes, "feasibleNodes", scheduleResult.FeasibleNodes)
			}

			prof.RunPostBindPlugins(bindingCycleCtx, state, assumedPod, scheduleResult.SuggestedHost)

			targetNode, _ := prof.SnapshotSharedLister().NodeInfos().Get(scheduleResult.SuggestedHost)
			schedRequest := assumedPod.Spec.Containers[0].Resources.Requests
			newNodeInfo := targetNode.Clone()
			newNode := newNodeInfo.Node()
			newNode.Status.Allocatable.CpuSub(schedRequest.Cpu())
			newNode.Status.Allocatable.MemSub(schedRequest.Memory())
			pc, _ := sched.SchedulerCache.PodCount()
			newNodeInfo.AddPod(assumedPod)
			//assumedPod.Spec.NodeName = newNode.Name
			sched.updateNodeInCache(targetNode.Node(), newNode)
			klog.Infof("Node %s; has pods %d/%d", newNode.Name, len(newNodeInfo.Pods), pc)

		}
	}()

}

func getAttemptsLabel(p *framework.QueuedPodInfo) string {
	if p.Attempts >= 15 {
		return "15+"
	}
	return strconv.Itoa(p.Attempts)
}

func (sched *Scheduler) profileForPod(pod *v1.Pod) (*profile.Profile, error) {
	prof, ok := sched.Profiles[pod.Spec.SchedulerName]
	if !ok {
		return nil, fmt.Errorf("profile not found for scheduler name %q", pod.Spec.SchedulerName)
	}
	return prof, nil
}

func (sched *Scheduler) skipPodSchedule(prof *profile.Profile, pod *v1.Pod) bool {
	if pod.DeletionTimestamp != nil {
		klog.Infof("Skip schedule deleting pod: %v/%v", pod.Namespace, pod.Name)
		return true
	}

	if sched.skipPodUpdate(pod) {
		return true
	}

	return false
}

func defaultAlgorithmSourceProviderName() *string {
	provider := schedulerapi.SchedulerDefaultProviderName
	return &provider
}
