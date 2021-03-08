package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/turtacn/cloud-prophet/scheduler"
	"github.com/turtacn/cloud-prophet/scheduler/framework/runtime"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	//"io/ioutil"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"time"
)

var (
	hostInfoFile           = flag.String("host-info", "hosts.csv", `节点元信息csv文件（包含起始状态）`)
	hostUtilityFile        = flag.String("host-util", "utility.csv", `节点CPU利用率csv文件，带时间戳`)
	scheduleTraceFile      = flag.String("schedule-trace", "schedule.csv", `调度trace文件`)
	scheduleIntervalSecond = flag.Int("pod-interval", 1, "pod资源请求间隔")
	printableHostFlag      = flag.Bool("print-host", false, "是否打印出候选节点的调度详情（默认false）")
)

type Option func(registry runtime.Registry) error

func main() {
	//  增加、去掉 log 相关的命令行参数
	//klog.InitFlags(flag.CommandLine)
	flag.Parse()
	//flag.Set("logtostderr", "false")
	//flag.Set("alsologtostderr", "false")
	//klog.SetOutput(ioutil.Discard)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//create a fake client
	scheduler, err := scheduler.New(nil,
		nil,
		nil,
		ctx.Done(),
		scheduler.WithPodMaxBackoffSeconds(0),
		scheduler.WithPercentageOfNodesToScore(100),
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
		scheduler.AddNode(makeNode(fmt.Sprintf("node-%d", i), 8000, 80000))
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
			scheduler.AddPod(pod)
			sleepInterval := *scheduleIntervalSecond
			if sleepInterval != 0 {
				time.Sleep(time.Duration(sleepInterval) * time.Second)
			}
		}

	}()

	go func() {
		for i := 1; true; i++ {
			sleepInterval := *scheduleIntervalSecond
			time.Sleep(time.Duration(5*sleepInterval) * time.Second)
			pod := &v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      fmt.Sprintf("pod-%d", i),
					Namespace: "test",
					UID:       fmt.Sprintf("test-%s", fmt.Sprintf("pod-%d", i)),
				}}
			p, e := scheduler.SchedulerCache.GetPod(pod)
			if e != nil {
				klog.Errorf("scheduler get pod failed error %v", e)
			}
			if p != nil {
				scheduler.Cache().RemovePod(p)
				klog.Infof("scheduler cache remove pod %v", p)
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
