package main

import (
	"flag"
	"github.com/turtacn/cloud-prophet/recommender/logic"
	"github.com/turtacn/cloud-prophet/recommender/model"
	vpa_types "github.com/turtacn/cloud-prophet/recommender/types"
	"k8s.io/klog"
	"time"
)

func main() {

	const (
		DefaultHostIp     = "127.0.0.1"
		DefaultVmId       = "i-xxxxxxxxxx"
		DefaultNcId       = "c-xxxxxxxxxx"
		DefaultPodId      = "pod-xxxxxxxxxx"
		DefaultRegionId   = "cn-north-1"
		GuangzhouRegionId = "cn-south-1"
		ShanghaiRegionId  = "cn-east-2"
		SuqianRegionId    = "cn-east-1"
	)

	var (
		anyTime = time.Unix(0, 0)
		hostIp  = flag.String("host-ip", DefaultHostIp, `the host ip of host identification.`)
		vmId    = flag.String("vm-id", DefaultVmId, `the identified string of vm instance.`)
		ncId    = flag.String("nc-id", DefaultNcId, `the identified string of nc instance.`)
		podId   = flag.String("pod-id", DefaultPodId, `the identified string of pod instance.`)
		region  = flag.String("region-id", DefaultRegionId, `region identification what instnaces were belonged to.`)
	)
	klog.InitFlags(nil)

	entityAggregateStateMap := make(model.ContainerNameToAggregateStateMap)

	if *hostIp != "" {
		entityAggregateStateMap[*hostIp] = model.NewAggregateContainerState()
	}
	if *ncId != "" {
		entityAggregateStateMap[*ncId] = model.NewAggregateContainerState()
	}
	if *podId != "" {
		entityAggregateStateMap[*podId] = model.NewAggregateContainerState()
	}
	if *vmId != "" {
		entityAggregateStateMap[*vmId] = model.NewAggregateContainerState()
	}

	if *region != DefaultRegionId || *region != GuangzhouRegionId || *region != ShanghaiRegionId || *region != SuqianRegionId {
		return
	}

	for _, s := range entityAggregateStateMap {
		timestamp := anyTime
		for i := 1; i <= 9; i++ {
			s.AddSample(&model.ContainerUsageSample{
				timestamp, model.CPUAmountFromCores(1.0), model.CPUAmountFromCores(10.0), model.ResourceCPU})
			timestamp = timestamp.Add(time.Minute * 2)
		}
	}

	setResourceRecommender := logic.CreatePodResourceRecommender()

	resources := setResourceRecommender.GetRecommendedPodResources(entityAggregateStateMap)

	containerResources := make([]vpa_types.RecommendedContainerResources, 0, len(resources))
	for containerName, res := range resources {
		containerResources = append(containerResources, vpa_types.RecommendedContainerResources{
			ContainerName:  containerName,
			Target:         model.ResourcesAsResourceList(res.Target),
			LowerBound:     model.ResourcesAsResourceList(res.LowerBound),
			UpperBound:     model.ResourcesAsResourceList(res.UpperBound),
			UncappedTarget: model.ResourcesAsResourceList(res.Target),
		})
	}
	recommendation := &vpa_types.RecommendedPodResources{containerResources}

	for _, recon := range recommendation.ContainerRecommendations {
		klog.Info("%s recommendation resource, target: %+v; upper: %+v, lower: %+v", recon.ContainerName, recon.Target, recon.UpperBound, recon.LowerBound, recon.UncappedTarget)
	}

}
