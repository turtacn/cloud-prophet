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
	hostInfoFile           = flag.String("host-info", "hosts.csv", `节点元信息csv文件（包含起始状态）`)
	scheduleTraceFile      = flag.String("schedule-trace", "schedule.csv", `调度trace文件`)
	scheduleIntervalSecond = flag.Int("pod-interval", 1, "pod资源请求间隔")
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
			pod := &v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      fmt.Sprintf("pod-%d", i),
					Namespace: "test",
					UID:       fmt.Sprintf("test-%s", fmt.Sprintf("pod-%d", i)),
				},
				Spec: v1.PodSpec{
					SchedulerName: v1.DefaultSchedulerName,
					Containers: append([]v1.Container{}, v1.Container{
						Name: fmt.Sprint("container-%d", i),
						Resources: v1.ResourceRequirements{
							Limits: v1.ResourceList{
								v1.ResourceCPU:    *resource.NewMilliQuantity(80, resource.DecimalSI),
								v1.ResourceMemory: *resource.NewQuantity(80, resource.BinarySI),
							},
							Requests: v1.ResourceList{
								v1.ResourceCPU:    *resource.NewMilliQuantity(80, resource.DecimalSI),
								v1.ResourceMemory: *resource.NewQuantity(80, resource.BinarySI),
							},
						},
					}),
				},
			}
			scheduler.SchedulingQueue.Add(pod)
			scheduler.SchedulingQueue.Delete(pod)
			sleepInterval := *scheduleIntervalSecond
			if sleepInterval != 0 {
				time.Sleep(time.Duration(sleepInterval) * time.Second)
			}
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
				v1.ResourceCPU:              *resource.NewMilliQuantity(milliCPU, resource.DecimalSI),
				v1.ResourceMemory:           *resource.NewQuantity(memory, resource.BinarySI),
				v1.ResourcePods:             *resource.NewQuantity(100, resource.BinarySI),
				v1.ResourceEphemeralStorage: *resource.NewQuantity(10000, resource.BinarySI),
				v1.ResourceStorage:          *resource.NewQuantity(100000, resource.BinarySI),
			},
			Allocatable: v1.ResourceList{
				v1.ResourceCPU:              *resource.NewMilliQuantity(milliCPU, resource.DecimalSI),
				v1.ResourceMemory:           *resource.NewQuantity(memory, resource.BinarySI),
				v1.ResourcePods:             *resource.NewQuantity(100, resource.BinarySI),
				v1.ResourceEphemeralStorage: *resource.NewQuantity(10000, resource.BinarySI),
				v1.ResourceStorage:          *resource.NewQuantity(100000, resource.BinarySI),
			},
			Conditions: append([]v1.NodeCondition{}, v1.NodeCondition{}),
		},
	}
}
