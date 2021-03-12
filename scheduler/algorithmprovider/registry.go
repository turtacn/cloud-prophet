//
package algorithmprovider

import (
	schedulerapi "github.com/turtacn/cloud-prophet/scheduler/apis/config"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/defaultbinder"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/interpodaffinity"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/noderesources"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/podtopologyspread"
	"github.com/turtacn/cloud-prophet/scheduler/framework/plugins/queuesort"
	"sort"
	"strings"
)

const ClusterK8sGeneralForJvirt = "ClusterK8sGeneralForJvirt"

type Registry map[string]*schedulerapi.Plugins

func NewRegistry() Registry {
	defaultConfig := getDefaultConfig()
	caConfig := getClusterAutoscalerConfig()

	return Registry{
		schedulerapi.SchedulerDefaultProviderName: defaultConfig,
		ClusterK8sGeneralForJvirt:                 caConfig,
	}
}

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
		Score: &schedulerapi.PluginSet{ // 核心打分逻辑
			Enabled: []schedulerapi.Plugin{
				// {Name: noderesources.RequestedToCapacityRatioName, Weight: 1},
				// {Name: noderesources.BalancedAllocationName, Weight: 1}, // 资源请求面向平衡
				//// 以下二选一， binpacking and spread strategies
				{Name: noderesources.MostAllocatedName, Weight: 1}, // binpack模式 剩余资源少优先
				// {Name: noderesources.LeastAllocatedName, Weight: 1},
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
	for i := range caConfig.Score.Enabled {
		if caConfig.Score.Enabled[i].Name == noderesources.LeastAllocatedName {
			caConfig.Score.Enabled[i].Name = noderesources.MostAllocatedName
		}
	}
	return caConfig
}
