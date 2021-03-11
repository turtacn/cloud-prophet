package runtime

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/turtacn/cloud-prophet/scheduler/apis/config"
	"github.com/turtacn/cloud-prophet/scheduler/apis/config/scheme"
	framework "github.com/turtacn/cloud-prophet/scheduler/framework/base"
	"github.com/turtacn/cloud-prophet/scheduler/helper/sets"
	"github.com/turtacn/cloud-prophet/scheduler/internal/parallelize"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

const (
	Filter                                    = "Filter"
	maxTimeout                  time.Duration = 15 * time.Minute
	preFilter                                 = "PreFilter"
	preFilterExtensionAddPod                  = "PreFilterExtensionAddPod"
	preFilterExtensionRemovePod               = "PreFilterExtensionRemovePod"
	postFilter                                = "PostFilter"
	preScore                                  = "PreScore"
	score                                     = "Score"
	scoreExtensionNormalize                   = "ScoreExtensionNormalize"
	preBind                                   = "PreBind"
	bind                                      = "Bind"
	postBind                                  = "PostBind"
	reserve                                   = "Reserve"
	unreserve                                 = "Unreserve"
	permit                                    = "Permit"
)

type frameworkImpl struct {
	registry              Registry
	snapshotSharedLister  framework.SharedLister
	waitingPods           *waitingPodsMap
	pluginNameToWeightMap map[string]int
	queueSortPlugins      []framework.QueueSortPlugin
	preFilterPlugins      []framework.PreFilterPlugin
	filterPlugins         []framework.FilterPlugin
	postFilterPlugins     []framework.PostFilterPlugin
	preScorePlugins       []framework.PreScorePlugin
	scorePlugins          []framework.ScorePlugin
	reservePlugins        []framework.ReservePlugin
	preBindPlugins        []framework.PreBindPlugin
	bindPlugins           []framework.BindPlugin
	postBindPlugins       []framework.PostBindPlugin
	permitPlugins         []framework.PermitPlugin

	clientSet       framework.ClientSet
	informerFactory framework.SharedInformer

	profileName string

	preemptHandle framework.PreemptHandle

	runAllFilters bool
}

type extensionPoint struct {
	plugins  *config.PluginSet
	slicePtr interface{}
}

func (f *frameworkImpl) getExtensionPoints(plugins *config.Plugins) []extensionPoint {
	return []extensionPoint{
		{plugins.PreFilter, &f.preFilterPlugins},
		{plugins.Filter, &f.filterPlugins},
		{plugins.PostFilter, &f.postFilterPlugins},
		{plugins.Reserve, &f.reservePlugins},
		{plugins.PreScore, &f.preScorePlugins},
		{plugins.Score, &f.scorePlugins},
		{plugins.PreBind, &f.preBindPlugins},
		{plugins.Bind, &f.bindPlugins},
		{plugins.PostBind, &f.postBindPlugins},
		{plugins.Permit, &f.permitPlugins},
		{plugins.QueueSort, &f.queueSortPlugins},
	}
}

type frameworkOptions struct {
	clientSet            framework.ClientSet
	informerFactory      framework.SharedInformer
	snapshotSharedLister framework.SharedLister
	profileName          string
	podNominator         framework.PodNominator
	runAllFilters        bool
}

type Option func(*frameworkOptions)

func WithClientSet(clientSet framework.ClientSet) Option {
	return func(o *frameworkOptions) {
		o.clientSet = clientSet
	}
}

func WithInformerFactory(informerFactory framework.SharedInformer) Option {
	return func(o *frameworkOptions) {
		o.informerFactory = informerFactory
	}
}

func WithSnapshotSharedLister(snapshotSharedLister framework.SharedLister) Option {
	return func(o *frameworkOptions) {
		o.snapshotSharedLister = snapshotSharedLister
	}
}

func WithRunAllFilters(runAllFilters bool) Option {
	return func(o *frameworkOptions) {
		o.runAllFilters = runAllFilters
	}
}

func WithProfileName(name string) Option {
	return func(o *frameworkOptions) {
		o.profileName = name
	}
}

func WithPodNominator(nominator framework.PodNominator) Option {
	return func(o *frameworkOptions) {
		o.podNominator = nominator
	}
}

var defaultFrameworkOptions = frameworkOptions{}

var _ framework.PreemptHandle = &preemptHandle{}

type preemptHandle struct {
	framework.PodNominator
	framework.PluginsRunner
}

var _ framework.Framework = &frameworkImpl{}

func NewFramework(r Registry, plugins *config.Plugins, args []config.PluginConfig, opts ...Option) (framework.Framework, error) {
	options := defaultFrameworkOptions
	for _, opt := range opts {
		opt(&options)
	}

	f := &frameworkImpl{
		registry:              r,
		snapshotSharedLister:  options.snapshotSharedLister,
		pluginNameToWeightMap: make(map[string]int),
		waitingPods:           newWaitingPodsMap(),
		clientSet:             options.clientSet,
		informerFactory:       options.informerFactory,
		profileName:           options.profileName,
		runAllFilters:         options.runAllFilters,
	}
	f.preemptHandle = &preemptHandle{
		PodNominator:  options.podNominator,
		PluginsRunner: f,
	}
	if plugins == nil {
		return f, nil
	}

	pg := f.pluginsNeeded(plugins)

	pluginConfig := make(map[string]runtime.Object, len(args))
	klog.Infof("length of plugin config is %d", len(args))
	for i := range args {
		name := args[i].Name
		klog.Infof("list plugin %s config args", name)
		if _, ok := pluginConfig[name]; ok {
			klog.Errorf("repeated config for plugin %s", name)
			return nil, fmt.Errorf("repeated config for plugin %s", name)
		}
		pluginConfig[name] = args[i].Args
	}

	pluginsMap := make(map[string]framework.Plugin)
	var totalPriority int64
	for name, factory := range r {
		klog.Infof("list plugin of framework %s", name)
		if _, ok := pg[name]; !ok {
			klog.Warningf("not found plugin name %s", name)
			continue
		}

		args, err := getPluginArgsOrDefault(pluginConfig, name)
		if err != nil {
			klog.Errorf("getting args for Plugin %q: %w", name, err)
			return nil, fmt.Errorf("getting args for Plugin %q: %w", name, err)
		}
		p, err := factory(args, f)
		if err != nil {
			return nil, fmt.Errorf("error initializing plugin %q: %v", name, err)
		}
		pluginsMap[name] = p

		f.pluginNameToWeightMap[name] = int(pg[name].Weight)
		if f.pluginNameToWeightMap[name] == 0 {
			f.pluginNameToWeightMap[name] = 1
		}
		if int64(f.pluginNameToWeightMap[name])*framework.MaxNodeScore > framework.MaxTotalScore-totalPriority {
			return nil, fmt.Errorf("total score of Score plugins could overflow")
		}
		totalPriority += int64(f.pluginNameToWeightMap[name]) * framework.MaxNodeScore
	}

	for _, e := range f.getExtensionPoints(plugins) {
		if err := updatePluginList(e.slicePtr, e.plugins, pluginsMap); err != nil {
			return nil, err
		}
	}

	for _, scorePlugin := range f.scorePlugins {
		if f.pluginNameToWeightMap[scorePlugin.Name()] == 0 {
			return nil, fmt.Errorf("score plugin %q is not configured with weight", scorePlugin.Name())
		}
	}

	if len(f.queueSortPlugins) == 0 {
		return nil, fmt.Errorf("no queue sort plugin is enabled")
	}
	if len(f.queueSortPlugins) > 1 {
		return nil, fmt.Errorf("only one queue sort plugin can be enabled")
	}
	if len(f.bindPlugins) == 0 {
		return nil, fmt.Errorf("at least one bind plugin is needed")
	}

	return f, nil
}

func getPluginArgsOrDefault(pluginConfig map[string]runtime.Object, name string) (runtime.Object, error) {
	res, ok := pluginConfig[name]
	if ok {
		return res, nil
	}
	klog.Infof("not found plugin config name %s, we create", name)
	return scheme.NewFromSchemeByName(name), nil
}

func updatePluginList(pluginList interface{}, pluginSet *config.PluginSet, pluginsMap map[string]framework.Plugin) error {
	if pluginSet == nil {
		return nil
	}

	plugins := reflect.ValueOf(pluginList).Elem()
	pluginType := plugins.Type().Elem()
	set := sets.NewString()
	for _, ep := range pluginSet.Enabled {
		pg, ok := pluginsMap[ep.Name]
		if !ok {
			return fmt.Errorf("%s %q does not exist", pluginType.Name(), ep.Name)
		}

		if !reflect.TypeOf(pg).Implements(pluginType) {
			return fmt.Errorf("plugin %q does not extend %s plugin", ep.Name, pluginType.Name())
		}

		if set.Has(ep.Name) {
			return fmt.Errorf("plugin %q already registered as %q", ep.Name, pluginType.Name())
		}

		set.Insert(ep.Name)

		newPlugins := reflect.Append(plugins, reflect.ValueOf(pg))
		plugins.Set(newPlugins)
	}
	return nil
}

func (f *frameworkImpl) QueueSortFunc() framework.LessFunc {
	if f == nil {
		return func(_, _ *framework.QueuedPodInfo) bool { return false }
	}

	if len(f.queueSortPlugins) == 0 {
		panic("No QueueSort plugin is registered in the frameworkImpl.")
	}

	return f.queueSortPlugins[0].Less
}

func (f *frameworkImpl) RunPreFilterPlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod) (status *framework.Status) {
	klog.Infof("To run %d prefilter plugin", len(f.preFilterPlugins))
	for _, pl := range f.preFilterPlugins {
		status = f.runPreFilterPlugin(ctx, pl, state, pod)
		if !status.IsSuccess() {
			if status.IsUnschedulable() {
				return status
			}
			msg := fmt.Sprintf("prefilter plugin %q failed for pod %q: %v", pl.Name(), pod.Name, status.Message())
			klog.Error(msg)
			return framework.NewStatus(framework.Error, msg)
		}
	}

	return nil
}

func (f *frameworkImpl) runPreFilterPlugin(ctx context.Context, pl framework.PreFilterPlugin, state *framework.CycleState, pod *v1.Pod) *framework.Status {
	return pl.PreFilter(ctx, state, pod)
}

func (f *frameworkImpl) RunPreFilterExtensionAddPod(
	ctx context.Context,
	state *framework.CycleState,
	podToSchedule *v1.Pod,
	podToAdd *v1.Pod,
	nodeInfo *framework.NodeInfo,
) (status *framework.Status) {
	for _, pl := range f.preFilterPlugins {
		if pl.PreFilterExtensions() == nil {
			continue
		}
		status = f.runPreFilterExtensionAddPod(ctx, pl, state, podToSchedule, podToAdd, nodeInfo)
		if !status.IsSuccess() {
			msg := fmt.Sprintf("error while running AddPod for plugin %q while scheduling pod %q: %v",
				pl.Name(), podToSchedule.Name, status.Message())
			klog.Error(msg)
			return framework.NewStatus(framework.Error, msg)
		}
	}

	return nil
}

func (f *frameworkImpl) runPreFilterExtensionAddPod(ctx context.Context, pl framework.PreFilterPlugin, state *framework.CycleState, podToSchedule *v1.Pod, podToAdd *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	return pl.PreFilterExtensions().AddPod(ctx, state, podToSchedule, podToAdd, nodeInfo)
}

func (f *frameworkImpl) RunPreFilterExtensionRemovePod(
	ctx context.Context,
	state *framework.CycleState,
	podToSchedule *v1.Pod,
	podToRemove *v1.Pod,
	nodeInfo *framework.NodeInfo,
) (status *framework.Status) {
	for _, pl := range f.preFilterPlugins {
		if pl.PreFilterExtensions() == nil {
			continue
		}
		status = f.runPreFilterExtensionRemovePod(ctx, pl, state, podToSchedule, podToRemove, nodeInfo)
		if !status.IsSuccess() {
			msg := fmt.Sprintf("error while running RemovePod for plugin %q while scheduling pod %q: %v",
				pl.Name(), podToSchedule.Name, status.Message())
			klog.Error(msg)
			return framework.NewStatus(framework.Error, msg)
		}
	}

	return nil
}

func (f *frameworkImpl) runPreFilterExtensionRemovePod(ctx context.Context, pl framework.PreFilterPlugin, state *framework.CycleState, podToSchedule *v1.Pod, podToAdd *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	status := pl.PreFilterExtensions().RemovePod(ctx, state, podToSchedule, podToAdd, nodeInfo)
	return status
}

func (f *frameworkImpl) RunFilterPlugins(
	ctx context.Context,
	state *framework.CycleState,
	pod *v1.Pod,
	nodeInfo *framework.NodeInfo,
) framework.PluginToStatus {
	statuses := make(framework.PluginToStatus)
	for _, pl := range f.filterPlugins {
		pluginStatus := f.runFilterPlugin(ctx, pl, state, pod, nodeInfo)
		if !pluginStatus.IsSuccess() {
			if !pluginStatus.IsUnschedulable() {
				errStatus := framework.NewStatus(framework.Error, fmt.Sprintf("running %q filter plugin for pod %q: %v", pl.Name(), pod.Name, pluginStatus.Message()))
				return map[string]*framework.Status{pl.Name(): errStatus}
			}
			statuses[pl.Name()] = pluginStatus
			if !f.runAllFilters {
				return statuses
			}
		}
	}

	return statuses
}

func (f *frameworkImpl) runFilterPlugin(ctx context.Context, pl framework.FilterPlugin, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {
	status := pl.Filter(ctx, state, pod, nodeInfo)
	return status
}

func (f *frameworkImpl) RunPostFilterPlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod, filteredNodeStatusMap framework.NodeToStatusMap) (_ *framework.PostFilterResult, status *framework.Status) {

	statuses := make(framework.PluginToStatus)
	for _, pl := range f.postFilterPlugins {
		r, s := f.runPostFilterPlugin(ctx, pl, state, pod, filteredNodeStatusMap)
		if s.IsSuccess() {
			return r, s
		} else if !s.IsUnschedulable() {
			return nil, framework.NewStatus(framework.Error, s.Message())
		}
		statuses[pl.Name()] = s
	}

	return nil, statuses.Merge()
}

func (f *frameworkImpl) runPostFilterPlugin(ctx context.Context, pl framework.PostFilterPlugin, state *framework.CycleState, pod *v1.Pod, filteredNodeStatusMap framework.NodeToStatusMap) (*framework.PostFilterResult, *framework.Status) {
	r, s := pl.PostFilter(ctx, state, pod, filteredNodeStatusMap)
	return r, s
}

func (f *frameworkImpl) RunPreScorePlugins(
	ctx context.Context,
	state *framework.CycleState,
	pod *v1.Pod,
	nodes []*v1.Node,
) (status *framework.Status) {
	for _, pl := range f.preScorePlugins {
		status = f.runPreScorePlugin(ctx, pl, state, pod, nodes)
		if !status.IsSuccess() {
			msg := fmt.Sprintf("error while running %q prescore plugin for pod %q: %v", pl.Name(), pod.Name, status.Message())
			klog.Error(msg)
			return framework.NewStatus(framework.Error, msg)
		}
	}

	return nil
}

func (f *frameworkImpl) runPreScorePlugin(ctx context.Context, pl framework.PreScorePlugin, state *framework.CycleState, pod *v1.Pod, nodes []*v1.Node) *framework.Status {
	return pl.PreScore(ctx, state, pod, nodes)
}

func (f *frameworkImpl) RunScorePlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodes []*v1.Node) (ps framework.PluginToNodeScores, status *framework.Status) {
	pluginToNodeScores := make(framework.PluginToNodeScores, len(f.scorePlugins))
	for _, pl := range f.scorePlugins {
		pluginToNodeScores[pl.Name()] = make(framework.NodeScoreList, len(nodes))
	}
	ctx, cancel := context.WithCancel(ctx)
	errCh := parallelize.NewErrorChannel()

	parallelize.Until(ctx, len(nodes), func(index int) {
		for _, pl := range f.scorePlugins {
			nodeName := nodes[index].Name
			s, status := f.runScorePlugin(ctx, pl, state, pod, nodeName)
			if !status.IsSuccess() {
				errCh.SendErrorWithCancel(fmt.Errorf(status.Message()), cancel)
				return
			}
			pluginToNodeScores[pl.Name()][index] = framework.NodeScore{
				Name:  nodeName,
				Score: int64(s),
			}
		}
	})
	if err := errCh.ReceiveError(); err != nil {
		msg := fmt.Sprintf("error while running score plugin for pod %q: %v", pod.Name, err)
		klog.Error(msg)
		return nil, framework.NewStatus(framework.Error, msg)
	}

	parallelize.Until(ctx, len(f.scorePlugins), func(index int) {
		pl := f.scorePlugins[index]
		nodeScoreList := pluginToNodeScores[pl.Name()]
		if pl.ScoreExtensions() == nil {
			return
		}
		status := f.runScoreExtension(ctx, pl, state, pod, nodeScoreList)
		if !status.IsSuccess() {
			err := fmt.Errorf("normalize score plugin %q failed with error %v", pl.Name(), status.Message())
			errCh.SendErrorWithCancel(err, cancel)
			return
		}
	})
	if err := errCh.ReceiveError(); err != nil {
		msg := fmt.Sprintf("error while running normalize score plugin for pod %q: %v", pod.Name, err)
		klog.Error(msg)
		return nil, framework.NewStatus(framework.Error, msg)
	}

	parallelize.Until(ctx, len(f.scorePlugins), func(index int) {
		pl := f.scorePlugins[index]
		weight := f.pluginNameToWeightMap[pl.Name()]
		nodeScoreList := pluginToNodeScores[pl.Name()]

		for i, nodeScore := range nodeScoreList {
			if nodeScore.Score > int64(framework.MaxNodeScore) || nodeScore.Score < int64(framework.MinNodeScore) {
				err := fmt.Errorf("score plugin %q returns an invalid score %v, it should in the range of [%v, %v] after normalizing", pl.Name(), nodeScore.Score, framework.MinNodeScore, framework.MaxNodeScore)
				errCh.SendErrorWithCancel(err, cancel)
				return
			}
			nodeScoreList[i].Score = nodeScore.Score * int64(weight)
		}
	})
	if err := errCh.ReceiveError(); err != nil {
		msg := fmt.Sprintf("error while applying score defaultWeights for pod %q: %v", pod.Name, err)
		klog.Error(msg)
		return nil, framework.NewStatus(framework.Error, msg)
	}

	return pluginToNodeScores, nil
}

func (f *frameworkImpl) runScorePlugin(ctx context.Context, pl framework.ScorePlugin, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	return pl.Score(ctx, state, pod, nodeName)
}

func (f *frameworkImpl) runScoreExtension(ctx context.Context, pl framework.ScorePlugin, state *framework.CycleState, pod *v1.Pod, nodeScoreList framework.NodeScoreList) *framework.Status {
	return pl.ScoreExtensions().NormalizeScore(ctx, state, pod, nodeScoreList)
}

func (f *frameworkImpl) RunPreBindPlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (status *framework.Status) {
	for _, pl := range f.preBindPlugins {
		status = f.runPreBindPlugin(ctx, pl, state, pod, nodeName)
		if !status.IsSuccess() {
			msg := fmt.Sprintf("error while running %q prebind plugin for pod %q: %v", pl.Name(), pod.Name, status.Message())
			klog.Error(msg)
			return framework.NewStatus(framework.Error, msg)
		}
	}
	return nil
}

func (f *frameworkImpl) runPreBindPlugin(ctx context.Context, pl framework.PreBindPlugin, state *framework.CycleState, pod *v1.Pod, nodeName string) *framework.Status {
	return pl.PreBind(ctx, state, pod, nodeName)
}

func (f *frameworkImpl) RunBindPlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (status *framework.Status) {
	if len(f.bindPlugins) == 0 {
		return framework.NewStatus(framework.Skip, "")
	}
	for _, bp := range f.bindPlugins {
		status = f.runBindPlugin(ctx, bp, state, pod, nodeName)
		if status != nil && status.Code() == framework.Skip {
			continue
		}
		if !status.IsSuccess() {
			msg := fmt.Sprintf("plugin %q failed to bind pod \"%v/%v\": %v", bp.Name(), pod.Namespace, pod.Name, status.Message())
			klog.Error(msg)
			return framework.NewStatus(framework.Error, msg)
		}
		return status
	}
	return status
}

func (f *frameworkImpl) runBindPlugin(ctx context.Context, bp framework.BindPlugin, state *framework.CycleState, pod *v1.Pod, nodeName string) *framework.Status {
	return bp.Bind(ctx, state, pod, nodeName)
}

func (f *frameworkImpl) RunPostBindPlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) {
	for _, pl := range f.postBindPlugins {
		f.runPostBindPlugin(ctx, pl, state, pod, nodeName)
	}
}

func (f *frameworkImpl) runPostBindPlugin(ctx context.Context, pl framework.PostBindPlugin, state *framework.CycleState, pod *v1.Pod, nodeName string) {
	pl.PostBind(ctx, state, pod, nodeName)
}

func (f *frameworkImpl) RunReservePluginsReserve(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (status *framework.Status) {
	for _, pl := range f.reservePlugins {
		status = f.runReservePluginReserve(ctx, pl, state, pod, nodeName)
		if !status.IsSuccess() {
			msg := fmt.Sprintf("error while running Reserve in %q reserve plugin for pod %q: %v", pl.Name(), pod.Name, status.Message())
			klog.Error(msg)
			return framework.NewStatus(framework.Error, msg)
		}
	}
	return nil
}

func (f *frameworkImpl) runReservePluginReserve(ctx context.Context, pl framework.ReservePlugin, state *framework.CycleState, pod *v1.Pod, nodeName string) *framework.Status {
	return pl.Reserve(ctx, state, pod, nodeName)
}

func (f *frameworkImpl) RunReservePluginsUnreserve(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) {
	for i := len(f.reservePlugins) - 1; i >= 0; i-- {
		f.runReservePluginUnreserve(ctx, f.reservePlugins[i], state, pod, nodeName)
	}
}

func (f *frameworkImpl) runReservePluginUnreserve(ctx context.Context, pl framework.ReservePlugin, state *framework.CycleState, pod *v1.Pod, nodeName string) {
	pl.Unreserve(ctx, state, pod, nodeName)
}

func (f *frameworkImpl) RunPermitPlugins(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (status *framework.Status) {
	pluginsWaitTime := make(map[string]time.Duration)
	statusCode := framework.Success
	for _, pl := range f.permitPlugins {
		status, timeout := f.runPermitPlugin(ctx, pl, state, pod, nodeName)
		if !status.IsSuccess() {
			if status.IsUnschedulable() {
				msg := fmt.Sprintf("rejected pod %q by permit plugin %q: %v", pod.Name, pl.Name(), status.Message())
				klog.Infof(msg)
				return framework.NewStatus(status.Code(), msg)
			}
			if status.Code() == framework.Wait {
				if timeout > maxTimeout {
					timeout = maxTimeout
				}
				pluginsWaitTime[pl.Name()] = timeout
				statusCode = framework.Wait
			} else {
				msg := fmt.Sprintf("error while running %q permit plugin for pod %q: %v", pl.Name(), pod.Name, status.Message())
				klog.Error(msg)
				return framework.NewStatus(framework.Error, msg)
			}
		}
	}
	if statusCode == framework.Wait {
		waitingPod := newWaitingPod(pod, pluginsWaitTime)
		f.waitingPods.add(waitingPod)
		msg := fmt.Sprintf("one or more plugins asked to wait and no plugin rejected pod %q", pod.Name)
		klog.Infof(msg)
		return framework.NewStatus(framework.Wait, msg)
	}
	return nil
}

func (f *frameworkImpl) runPermitPlugin(ctx context.Context, pl framework.PermitPlugin, state *framework.CycleState, pod *v1.Pod, nodeName string) (*framework.Status, time.Duration) {
	return pl.Permit(ctx, state, pod, nodeName)
}

func (f *frameworkImpl) WaitOnPermit(ctx context.Context, pod *v1.Pod) (status *framework.Status) {
	waitingPod := f.waitingPods.get(pod.UID)
	if waitingPod == nil {
		return nil
	}
	defer f.waitingPods.remove(pod.UID)
	klog.Infof("pod %q waiting on permit", pod.Name)

	s := <-waitingPod.s

	if !s.IsSuccess() {
		if s.IsUnschedulable() {
			msg := fmt.Sprintf("pod %q rejected while waiting on permit: %v", pod.Name, s.Message())
			klog.Infof(msg)
			return framework.NewStatus(s.Code(), msg)
		}
		msg := fmt.Sprintf("error received while waiting on permit for pod %q: %v", pod.Name, s.Message())
		klog.Error(msg)
		return framework.NewStatus(framework.Error, msg)
	}
	return nil
}

func (f *frameworkImpl) SnapshotSharedLister() framework.SharedLister {
	return f.snapshotSharedLister
}

func (f *frameworkImpl) IterateOverWaitingPods(callback func(framework.WaitingPod)) {
	f.waitingPods.iterate(callback)
}

func (f *frameworkImpl) GetWaitingPod(uid string) framework.WaitingPod {
	if wp := f.waitingPods.get(uid); wp != nil {
		return wp
	}
	return nil // Returning nil instead of *waitingPod(nil).
}

func (f *frameworkImpl) RejectWaitingPod(uid string) {
	waitingPod := f.waitingPods.get(uid)
	if waitingPod != nil {
		waitingPod.Reject("removed")
	}
}

func (f *frameworkImpl) HasFilterPlugins() bool {
	return len(f.filterPlugins) > 0
}

func (f *frameworkImpl) HasPostFilterPlugins() bool {
	return len(f.postFilterPlugins) > 0
}

func (f *frameworkImpl) HasScorePlugins() bool {
	return len(f.scorePlugins) > 0
}

func (f *frameworkImpl) ListPlugins() map[string][]config.Plugin {
	m := make(map[string][]config.Plugin)

	for _, e := range f.getExtensionPoints(&config.Plugins{}) {
		plugins := reflect.ValueOf(e.slicePtr).Elem()
		extName := plugins.Type().Elem().Name()
		var cfgs []config.Plugin
		for i := 0; i < plugins.Len(); i++ {
			name := plugins.Index(i).Interface().(framework.Plugin).Name()
			p := config.Plugin{Name: name}
			if extName == "ScorePlugin" {
				p.Weight = int32(f.pluginNameToWeightMap[name])
			}
			cfgs = append(cfgs, p)
		}
		if len(cfgs) > 0 {
			m[extName] = cfgs
		}
	}
	if len(m) > 0 {
		return m
	}
	return nil
}

func (f *frameworkImpl) ClientSet() framework.ClientSet {
	return f.clientSet
}

func (f *frameworkImpl) SharedInformerFactory() framework.SharedInformer {
	return f.informerFactory
}

func (f *frameworkImpl) pluginsNeeded(plugins *config.Plugins) map[string]config.Plugin {
	pgMap := make(map[string]config.Plugin)

	if plugins == nil {
		return pgMap
	}

	find := func(pgs *config.PluginSet) {
		if pgs == nil {
			return
		}
		for _, pg := range pgs.Enabled {
			pgMap[pg.Name] = pg
		}
	}
	for _, e := range f.getExtensionPoints(plugins) {
		find(e.plugins)
	}
	return pgMap
}

func (f *frameworkImpl) PreemptHandle() framework.PreemptHandle {
	return f.preemptHandle
}
