package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/turtacn/cloud-prophet/scheduler"
	"github.com/turtacn/cloud-prophet/scheduler/framework/runtime"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"k8s.io/klog"
	"time"
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
	client := fake.NewSimpleClientset()

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
	//go podInformer.Informer().Run(ctx.Done())
	//clusterInformer.Start(ctx.Done())
	//clusterInformer.WaitForCacheSync(ctx.Done())

	// make nodes
	for i := 1; i <= 200; i++ {
		node := makeNode(fmt.Sprintf("node-%d", i), 8000, 80000)
		if err := scheduler.SchedulerCache.AddNode(node); err != nil {
			klog.Warningf("scheduler cache add node failed %v", err)
		}
	}

	go func() {
		for i := 1; true; i++ {
			scheduler.SchedulingQueue.Add(&v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      fmt.Sprintf("pod-%d", i),
					Namespace: "test",
					UID:       fmt.Sprint("test-pod-%s", fmt.Sprintf("pod-%d", i)),
				},
				Spec: v1.PodSpec{
					SchedulerName: v1.DefaultSchedulerName,
				},
			})
			time.Sleep(1 * time.Second)
		}

	}()

	klog.Infof("begin to run scheduler")
	scheduler.Run(ctx)
}

func makeNode(node string, milliCPU, memory int64) *v1.Node {
	return &v1.Node{
		Spec: v1.NodeSpec{
			Unschedulable: false,
		},
		ObjectMeta: v1.ObjectMeta{Name: node},
		Status: v1.NodeStatus{
			Phase: v1.NodeRunning,
			Capacity: v1.ResourceList{
				v1.ResourceCPU:    *resource.NewMilliQuantity(milliCPU, resource.DecimalSI),
				v1.ResourceMemory: *resource.NewQuantity(memory, resource.BinarySI),
			},
			Allocatable: v1.ResourceList{
				v1.ResourceCPU:    *resource.NewMilliQuantity(milliCPU, resource.DecimalSI),
				v1.ResourceMemory: *resource.NewQuantity(memory, resource.BinarySI),
			},
		},
	}
}
