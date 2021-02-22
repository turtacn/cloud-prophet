package main

import (
	"context"
	"flag"
	"github.com/turtacn/cloud-prophet/scheduler"
	"github.com/turtacn/cloud-prophet/scheduler/framework/runtime"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/klog"
)

var (
	hostInfoFile      = flag.String("host-info", "hosts.csv", `节点元信息csv文件（包含起始状态）`)
	scheduleTraceFile = flag.String("schedule-trace", "schedule.csv", `调度trace文件`)
)

type Option func(registry runtime.Registry) error

func main() {
	flag.Parse()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//create a fake client
	client := fake.NewSimpleClientset(
		&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "influxdb-v2",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
		}, &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "chronograf",
				Namespace:   "default",
				Annotations: map[string]string{},
			},
		})

	clusterInformer := informers.NewSharedInformerFactory(client, 0)
	podInformer := scheduler.NewPodInformer(client, 0)
	scheduler, err := scheduler.New(client,
		clusterInformer,
		podInformer,
		ctx.Done(),
		scheduler.WithPodMaxBackoffSeconds(0),
		scheduler.WithPercentageOfNodesToScore(0),
	)

	if err != nil {
		klog.Fatalf("Init uniform scheduler derived from k8s scheduler. error=%+v", err)
		return
	}

	go podInformer.Informer().Run(ctx.Done())

	clusterInformer.Start(ctx.Done())
	clusterInformer.WaitForCacheSync(ctx.Done())
	klog.Infof("begin to run scheduler")
	scheduler.Run(ctx)
}
