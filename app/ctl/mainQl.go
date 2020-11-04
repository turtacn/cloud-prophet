package main

import (
	"fmt"
	"github.com/SOUP-CE-KMITL/Thoth"
	"github.com/SOUP-CE-KMITL/Thoth/learn"
	"github.com/SOUP-CE-KMITL/Thoth/profil"
	"github.com/influxdata/influxdb/client/v2"
	"os"
	"os/signal"
	"time"
)

var username string = "thoth"
var password string = "thoth"

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
		Addr:     thoth.InfluxdbApi,
		Username: username,
		Password: password,
	})

	if err != nil {
		panic(err)
	}

	firstRun := true
	lastState := learn.State{}
	lastAction := 0
	var action int
	for {
		// Get all user RC
		RC := profil.GetUserRC()
		RCLen := len(RC)

		// Getting App Metric
		for i := 0; i < RCLen; i++ {
			// TODO : Use model only with Eight-puzzle
			if RC[i].Namespace == "thoth" && RC[i].Name == "eight-puzzle" {
				replicas, err := profil.GetReplicas(RC[i].Namespace, RC[i].Name)
				if err != nil {
					panic(err)
				}

				res := profil.GetProfilLast(influxDB, RC[i].Namespace, RC[i].Name, "5m")
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
					if _, err := thoth.ScaleOutViaCli(replicas+action, RC[i].Namespace, RC[i].Name); err != nil {
						fmt.Println(err)
					}
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
