// 调度算法的组成,包含哪些Plugin; 打分的Plugin的静态权重
//
package algorithmprovider

import (
	schedulerapi "github.com/turtacn/cloud-prophet/scheduler/apis/config"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/defaultbinder"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/interpodaffinity"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/noderesources"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/podtopologyspread"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/queuesort"
	// 这里扩展调度插件扩展的引用
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
	caConfig := getClusterAutoscalerConfig()

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
				{Name: noderesources.FitName},
				{Name: podtopologyspread.Name},
				{Name: interpodaffinity.Name},
			},
		},
		PostFilter: &schedulerapi.PluginSet{}, // 扩展支持抢占
		PreScore: &schedulerapi.PluginSet{
			Enabled: []schedulerapi.Plugin{
				{Name: interpodaffinity.Name},
				{Name: podtopologyspread.Name},
			},
		},
		// 核心打分算法
		Score: &schedulerapi.PluginSet{
			Enabled: []schedulerapi.Plugin{
				{Name: noderesources.RequestedToCapacityRatioName, Weight: 0}, // 资源请求面向节点可用资源成比例优先
				{Name: noderesources.BalancedAllocationName, Weight: 1},       // 资源请求面向平衡
				//{Name: noderesources.LeastAllocatedName, Weight: 0},             // spread 模式 剩余资源多优先
				{Name: noderesources.MostAllocatedName, Weight: 1}, // binpack模式 剩余资源少优先
				// Weight is doubled because:
				// - This is a score coming from user preference.
				// - It makes its signal comparable to NodeResourcesLeastAllocated.
				{Name: podtopologyspread.Name, Weight: 2},
			},
		},
		Reserve: &schedulerapi.PluginSet{}, // volume, port binding 阶段
		PreBind: &schedulerapi.PluginSet{}, // volume, port binding 阶段
		Bind: &schedulerapi.PluginSet{
			Enabled: []schedulerapi.Plugin{
				{Name: defaultbinder.Name},
			},
		},
		PostBind: &schedulerapi.PluginSet{
			Enabled: []schedulerapi.Plugin{
				{Name: defaultbinder.NameFakeAllocater},
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
