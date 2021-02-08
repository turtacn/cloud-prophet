package scheduler

import (
	"context"
	"flag"
	"github.com/turtacn/cloud-prophet/scheduler"
	"github.com/turtacn/cloud-prophet/scheduler/apis/config"
	"github.com/turtacn/cloud-prophet/scheduler/framework/runtime"
)

type Option func(registry runtime.Registry) error

func main() {
	flag.Parse()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	scheduler, err := scheduler.New(nil,
		nil,
		nil,
		ctx.Done(),
		scheduler.WithAlgorithmSource(config.SchedulerAlgorithmSource{}),
		scheduler.WithExtenders(config.Extender{}),
	)

	if err != nil {
		return
	}

	scheduler.Run(ctx)

}
