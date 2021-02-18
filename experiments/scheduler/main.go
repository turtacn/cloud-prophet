package main

import (
	"context"
	"flag"
	"github.com/turtacn/cloud-prophet/scheduler"
	"github.com/turtacn/cloud-prophet/scheduler/apis/config"
	"github.com/turtacn/cloud-prophet/scheduler/framework/runtime"
	"k8s.io/client-go/informers"
	kubernetes "k8s.io/client-go/kubernetes/fake"
	"k8s.io/klog"
)

type Option func(registry runtime.Registry) error

func main() {
	flag.Parse()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//create a fake client
	client := kubernetes.NewSimpleClientset()

	clusterInformer := informers.NewSharedInformerFactory(client, 0)
	podInformer := scheduler.NewPodInformer(client, 0)
	schedulerAlgorithmProvider := config.SchedulerDefaultProviderName
	scheduler, err := scheduler.New(client,
		clusterInformer,
		podInformer,
		ctx.Done(),
		scheduler.WithAlgorithmSource(config.SchedulerAlgorithmSource{
			Provider: &schedulerAlgorithmProvider,
		}),
		scheduler.WithExtenders(config.Extender{}),
	)

	if err != nil {
		klog.Fatalf("Init uniform scheduler derived from k8s scheduler. error=%+v", err)
		return
	}

	scheduler.Run(ctx)

}
