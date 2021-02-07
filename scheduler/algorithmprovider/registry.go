//
//
package algorithmprovider

import (
	schedulerapi "github.com/turtacn/cloud-prophet/scheduler/apis/config"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/defaultbinder"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/defaultpreemption"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/imagelocality"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/interpodaffinity"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/nodeaffinity"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/nodename"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/noderesources"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/nodeunschedulable"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/podtopologyspread"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/queuesort"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/selectorspread"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/tainttoleration"
	// 这里扩展调度插件扩展的引用
	"k8s.io/klog/v2"
	"sort"
	"strings"
)

// ClusterAutoscalerProvider defines the default autoscaler provider
const ClusterAutoscalerProvider = "ClusterAutoscalerProvider"

// Registry is a collection of all available algorithm providers.
type Registry map[string]*schedulerapi.Plugins

// NewRegistry returns an algorithm provider registry instance.
func NewRegistry() Registry {
	defaultConfig := getDefaultConfig()
	applyFeatureGates(defaultConfig)

	caConfig := getClusterAutoscalerConfig()
	applyFeatureGates(caConfig)

	return Registry{
		schedulerapi.SchedulerDefaultProviderName: defaultConfig,
		ClusterAutoscalerProvider:                 caConfig,
	}
}

// ListAlgorithmProviders lists registered algorithm providers.
func ListAlgorithmProviders() string {
	r := NewRegistry()
	var providers []string
	for k := range r {
		providers = append(providers, k)
	}
	sort.Strings(providers)
	return strings.Join(providers, " | ")
}

func getDefaultConfig() *schedulerapi.Plugins {
	return &schedulerapi.Plugins{
		QueueSort: &schedulerapi.PluginSet{
			Enabled: []schedulerapi.Plugin{
				{Name: queuesort.Name},
			},
		},
		PreFilter: &schedulerapi.PluginSet{
			Enabled: []schedulerapi.Plugin{
				{Name: noderesources.FitName},
				{Name: podtopologyspread.Name},
				{Name: interpodaffinity.Name},
			},
		},
		Filter: &schedulerapi.PluginSet{
			Enabled: []schedulerapi.Plugin{
				{Name: nodeunschedulable.Name},
				{Name: noderesources.FitName},
				{Name: nodename.Name},
				{Name: nodeaffinity.Name},
				{Name: tainttoleration.Name},
				{Name: podtopologyspread.Name},
				{Name: interpodaffinity.Name},
			},
		},
		PostFilter: &schedulerapi.PluginSet{
			Enabled: []schedulerapi.Plugin{
				{Name: defaultpreemption.Name},
			},
		},
		PreScore: &schedulerapi.PluginSet{
			Enabled: []schedulerapi.Plugin{
				{Name: interpodaffinity.Name},
				{Name: podtopologyspread.Name},
				{Name: tainttoleration.Name},
			},
		},
		Score: &schedulerapi.PluginSet{
			Enabled: []schedulerapi.Plugin{
				{Name: noderesources.BalancedAllocationName, Weight: 1},
				{Name: imagelocality.Name, Weight: 1},
				{Name: interpodaffinity.Name, Weight: 1},
				{Name: noderesources.LeastAllocatedName, Weight: 1},
				{Name: nodeaffinity.Name, Weight: 1},
				// Weight is doubled because:
				// - This is a score coming from user preference.
				// - It makes its signal comparable to NodeResourcesLeastAllocated.
				{Name: podtopologyspread.Name, Weight: 2},
				{Name: tainttoleration.Name, Weight: 1},
			},
		},
		Reserve: &schedulerapi.PluginSet{},
		PreBind: &schedulerapi.PluginSet{},
		Bind: &schedulerapi.PluginSet{
			Enabled: []schedulerapi.Plugin{
				{Name: defaultbinder.Name},
			},
		},
	}
}

func getClusterAutoscalerConfig() *schedulerapi.Plugins {
	caConfig := getDefaultConfig()
	// Replace least with most requested.
	for i := range caConfig.Score.Enabled {
		if caConfig.Score.Enabled[i].Name == noderesources.LeastAllocatedName {
			caConfig.Score.Enabled[i].Name = noderesources.MostAllocatedName
		}
	}
	return caConfig
}

func applyFeatureGates(config *schedulerapi.Plugins) {
	// When feature is enabled, the default spreading is done by
	// PodTopologySpread plugin, which is enabled by default.
	klog.Infof("Registering SelectorSpread plugin")
	s := schedulerapi.Plugin{Name: selectorspread.Name}
	config.PreScore.Enabled = append(config.PreScore.Enabled, s)
	s.Weight = 1
	config.Score.Enabled = append(config.Score.Enabled, s)
}
