package defaultbinder

import (
	"context"
	framework "github.com/turtacn/cloud-prophet/scheduler/framework/base"
	v1 "github.com/turtacn/cloud-prophet/scheduler/model"
	"k8s.io/apimachinery/pkg/runtime"
)

const NameFakeAllocater = "FakeAllocater"

type FakeAllocater struct {
	handle framework.FrameworkHandle
}

var _ framework.PostBindPlugin = &FakeAllocater{}

func NewFA(_ runtime.Object, handle framework.FrameworkHandle) (framework.Plugin, error) {
	return &FakeAllocater{handle: handle}, nil
}

func (b FakeAllocater) Name() string {
	return NameFakeAllocater
}

func (b FakeAllocater) PostBind(ctx context.Context, state *framework.CycleState, p *v1.Pod, nodeName string) {
}
