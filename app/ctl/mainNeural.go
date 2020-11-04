package main

import (
	//"encoding/csv"
	"fmt"
	"github.com/SOUP-CE-KMITL/Thoth"
	"github.com/SOUP-CE-KMITL/Thoth/learn"
	"github.com/SOUP-CE-KMITL/Thoth/profil"
	"github.com/influxdata/influxdb/client/v2"
	//	"io"
	"log"
	"math"
	//	"math/rand"
	"os"
	"os/signal"
	//	"strconv"
	//	"strings"
	"time"
)

var username string = "thoth"
var password string = "thoth"

func main() {
	ann := learn.Neural{}
	ann.Init("fann.dat")
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			// sig is a ^C, handle it
			fmt.Println(sig)
			ann.Save("fann.dat")
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
	round := 0
	correct := 0
	// TODO:it's need to calculate this for all RC
	ann.Avg = profil.GetProfilLast(influxDB, "thoth", "eight-puzzle", "4d")
	ann.StdDev = profil.GetProfilStdLast(influxDB, "thoth", "eight-puzzle", "4d")
	fmt.Println("AVG ", ann.Avg)
	fmt.Println("STDDEV ", ann.StdDev)

	for {
		// Get all user RCLen
		RC := profil.GetUserRC()
		RCLen := len(RC)

		// Getting App Metric
		for i := 0; i < RCLen; i++ {
			round++
			action := 0.0

			replicas, err := profil.GetReplicas(RC[i].Namespace, RC[i].Name)

			if err != nil {
				panic(err)
			}
			fmt.Println("Replicas:", replicas)

			// Check Resposne time & Label & Save WPI
			var responseDay, response10Min float64
			if responseDay, err = profil.GetProfilAvg(influxDB, RC[i].Namespace, RC[i].Name, "rtime", "1d"); err != nil {
				panic(err)
				log.Println(err)
			}
			fmt.Println("resDays : ", responseDay)

			if response10Min, err = profil.GetProfilAvg(influxDB, RC[i].Namespace, RC[i].Name, "rtime", "5m"); err != nil {
				fmt.Println("res10min : ", response10Min)
				panic(err)
				log.Println(err)
			}
			// Floor
			responseDay = math.Floor(responseDay)
			response10Min = math.Floor(response10Min)
			fmt.Println("D", responseDay, " 10M", response10Min)
			metrics := profil.GetAppResource(RC[i].Namespace, RC[i].Name)
			var cpu10Min float64
			if cpu10Min, err = profil.GetProfilAvg(influxDB, RC[i].Namespace, RC[i].Name, "cpu", "5m"); err != nil {
				panic(err)
				log.Println(err)
			}
			fmt.Println("CPU ", cpu10Min)
			if cpu10Min > 70 {
				fmt.Println("Response check")
				//	if response10Min > responseDay { // TODO:Need to check WPI too
				// Save WPI
				fmt.Println("Scale+1")
				if err := profil.WriteRPI(influxDB, RC[i].Namespace, RC[i].Name, metrics.Request, replicas); err != nil {
					panic(err)
					log.Println(err)
				}
				// Scale +1
				// TODO: Limit

				if replicas < 10 {
					action = 1
					//	if _, err := thoth.ScaleOutViaCli(replicas+1, RC[i].Namespace, RC[i].Name); err != nil {
					//		panic(err)
					//	}
				}

				//	}
			} else if replicas > 1 {
				// = rpi/replicas
				var rpiMax float64
				if rpiMax, err = profil.GetAvgRPI(influxDB, RC[i].Namespace, RC[i].Name); err != nil {
					rpiMax = -1
					// TODO:Handler
					//panic(err)
				}
				fmt.Println("WPI", rpiMax)
				if rpiMax > 0 {
					minReplicas := int(metrics.Request / int64(rpiMax)) // TODO: Ceil?

					if minReplicas < replicas {
						// Scale -1
						fmt.Println("Scale-1")
						action = -1
						//if _, err := thoth.ScaleOutViaCli(replicas-1, RC[i].Namespace, RC[i].Name); err != nil {
						//	panic(err)
						//}
					}
				}
			}

			// Normalize
			resUsage10min := profil.GetProfilLast(influxDB, RC[i].Namespace, RC[i].Name, "10min")
			fmt.Println("============================ FANN ============================")
			// Training
			ann.Train(resUsage10min, action)
			// Run (Predict)
			predict := ann.Run(resUsage10min)
			if predict != 0 {
				if predict == 1 {
					if _, err := thoth.ScaleOutViaCli(replicas+1, RC[i].Namespace, RC[i].Name); err != nil {
						fmt.Println(err)
					}
				} else if predict == -1 && replicas > 1 {
					if _, err := thoth.ScaleOutViaCli(replicas-1, RC[i].Namespace, RC[i].Name); err != nil {
						fmt.Println(err)
					}
				}
			}
			if predict == int(action) {
				correct++
			}
			fmt.Println("Stats Round: ", round, " Correct:", correct, " Accuracy: ", (correct*100)/round, "%")
			//-----------
		}

		fmt.Println("Sleep TODO:Change to 5 Minnn")
		time.Sleep(300 * time.Second)
	}
}
