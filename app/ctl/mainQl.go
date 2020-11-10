package main

import (
	"fmt"
	"github.com/influxdata/influxdb1-client/v2"
	"github.com/turtacn/cloud-prophet/learn"
	"github.com/turtacn/cloud-prophet/model"
	"github.com/turtacn/cloud-prophet/profil"
	"os"
	"os/signal"
	"time"
)

func main() {
	agent := learn.QLearn{Gamma: 0.3}
	agent.Init()
	if err := agent.Load("ql.da"); err != nil {
		fmt.Println("Load Fail", err)
	}
	agent.Epsilon = 0.0

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			// sig is a ^C, handle it
			fmt.Println(sig)
			agent.Save("ql.da")
			os.Exit(0)
		}
	}()

	// Connect InfluxDB
	influxDB, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     model.InfluxdbApi,
		Username: model.UserName,
		Password: model.PassWord,
	})

	if err != nil {
		fmt.Println(err)
	}

	firstRun := true
	lastState := learn.State{}
	lastAction := 0
	var action int
	for {
		// Get all host
		// Getinstances by hostIp
		// Getting App Metric
		var instanceId string = ""
		for i := 0; i < 10; i++ {
			// TODO : Use model only with Eight-puzzle
			if instanceId == "eight-puzzle" {
				replicas := profil.GetInstanceCount(instanceId, "cpu")

				res := profil.GetProfilLast(influxDB, instanceId, "cpu", "5m")
				fmt.Println(res)

				if !firstRun {
					// Reward Last state
					agent.Reward(lastState, lastAction, res)
				}

				agent.SetCurrentState(res["cpu"], res["memory"], res["rps"], res["rtime"], res["r5xx"], replicas)
				action = agent.ChooseAction()
				lastState = agent.CurrentState
				firstRun = false

				if action+replicas > 0 {
					// 开始操作
				}
				lastAction = action
				fmt.Println(agent)
				fmt.Println("C", agent.CurrentState.CPUH,
					"M", agent.CurrentState.MemH,
					"R", agent.CurrentState.RpsH,
					"T", agent.CurrentState.RtimeH,
					"5", agent.CurrentState.R5xxH)
			}

			fmt.Println(lastState)
			fmt.Println(lastAction)
		}
		//-----------
		fmt.Println("Sleep TODO:Change to 5 Min\n")
		time.Sleep(60 * time.Second)
	}
}
