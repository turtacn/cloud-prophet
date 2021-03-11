package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/turtacn/cloud-prophet/scheduler"
	"github.com/turtacn/cloud-prophet/scheduler/framework/runtime"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	"github.com/turtacn/cloud-prophet/scheduler/test"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
	"time"
)

var (
	hostInfoFile           = flag.String("host-info", "hosts-info.csv", `节点元信息csv文件（包含起始状态）`)
	hostUtilityFile        = flag.String("host-util", "utility.csv", `节点CPU利用率csv文件，带时间戳`)
	scheduleTraceFile      = flag.String("schedule-trace", "schedule.csv", `调度trace文件`)
	scheduleIntervalSecond = flag.Int("pod-interval", 100, "资源请求间隔")
	printableHostFlag      = flag.Bool("print-host", false, "是否打印出候选节点的调度详情（默认false）")
)

type Option func(registry runtime.Registry) error

func main() {
	flag.Parse()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// scheduler with a fake client
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

	// Load host info csv file
	for round := 0; round < 2; round++ {
		for i, h := range test.LoadHostInfo(*hostInfoFile) {
			if i == 0 {
				continue
			}
			klog.Infof("Node %s cpu %f memory %f", h.HostIp, h.AvailableCpu(), h.AvailableMemory())
			scheduler.AddNode(makeNode(fmt.Sprintf("%s-%d", h.HostIp, round), int64(h.AvailableCpu()), int64(h.AvailableMemory())))
		}
	}
	// 生成Trace
	// test.FillTrace("trace.csv","instance.csv","schedule.csv")
	go func() {
		Traces := test.LoadIntanceOpsTrace(*scheduleTraceFile)
		for i, jvirt := range Traces {
			if i == 0 {
				continue
			}
			pod := &v1.Pod{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Pod",
					APIVersion: "v1",
				},
				ObjectMeta: v1.ObjectMeta{
					Name:      jvirt.InstanceId,
					Namespace: "jvirt/bj02/general",
					UID:       fmt.Sprintf("jvirt/bj02/general/%s", jvirt.InstanceId),
				},
				Spec: v1.PodSpec{
					SchedulerName: v1.DefaultSchedulerName,
					Containers: append([]v1.Container{}, v1.Container{
						Name: fmt.Sprint("%s-%d", jvirt.InstanceId, i),
						Resources: v1.ResourceRequirements{
							Limits: v1.ResourceList{
								v1.ResourceCPU:    *resource.NewMilliQuantity(int64(jvirt.RequestCpu()), resource.DecimalSI),
								v1.ResourceMemory: *resource.NewQuantity(int64(jvirt.RequestMem()), resource.BinarySI),
							},
							Requests: v1.ResourceList{
								v1.ResourceCPU:    *resource.NewMilliQuantity(int64(jvirt.RequestCpu()), resource.DecimalSI),
								v1.ResourceMemory: *resource.NewQuantity(int64(jvirt.RequestMem()), resource.BinarySI),
							},
						},
					}),
				},
			}
			switch jvirt.OpAction {
			case "alloc":
				scheduler.AddPod(pod)

			case "free":
				p, err := scheduler.SchedulerCache.GetPod(pod)
				if err == nil {
					scheduler.DeletePod(p)
				} else {
					klog.Errorf("=====>delete pod failed %v", err)
				}
			}
			sleepInterval := *scheduleIntervalSecond
			if sleepInterval != 0 {
				time.Sleep(time.Duration(sleepInterval) * time.Millisecond)
			}
			klog.Infof("Process %d/%d", i, len(Traces))
		}
		klog.Infof("===================================================\nTrace Replay was finished!!!")
	}()
	klog.Infof("begin to run scheduler")
	scheduler.Run(ctx)
}

func makeNode(node string, milliCPU, memory int64) *v1.Node {
	return &v1.Node{
		Spec: v1.NodeSpec{
			Unschedulable: false,
		},
		ObjectMeta: v1.ObjectMeta{
			Name: node,
		},
		Status: v1.NodeStatus{
			Phase: v1.NodeRunning,
			Capacity: v1.ResourceList{
				v1.ResourceCPU:    *resource.NewMilliQuantity(milliCPU, resource.DecimalSI),
				v1.ResourceMemory: *resource.NewQuantity(memory, resource.BinarySI),
				v1.ResourcePods:   *resource.NewQuantity(200, resource.BinarySI), // 线上最大120个实例
				//v1.ResourceEphemeralStorage: *resource.NewQuantity(10000, resource.BinarySI),
				//v1.ResourceStorage:          *resource.NewQuantity(100000, resource.BinarySI),
			},
			Allocatable: v1.ResourceList{
				v1.ResourceCPU:    *resource.NewMilliQuantity(milliCPU, resource.DecimalSI),
				v1.ResourceMemory: *resource.NewQuantity(memory, resource.BinarySI),
				v1.ResourcePods:   *resource.NewQuantity(200, resource.BinarySI),
				//v1.ResourceEphemeralStorage: *resource.NewQuantity(10000, resource.BinarySI),
				//v1.ResourceStorage:          *resource.NewQuantity(100000, resource.BinarySI),
			},
		},
	}
}
