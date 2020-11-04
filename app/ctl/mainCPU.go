package main

import (
	"encoding/csv"
	"fmt"
	"github.com/SOUP-CE-KMITL/Thoth"
	"github.com/SOUP-CE-KMITL/Thoth/profil"
	"github.com/goml/gobrain"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/white-pony/go-fann"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

var username string = "thoth"
var password string = "thoth"

func main() {
	// Connect InfluxDB
	influxDB, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     thoth.InfluxdbApi,
		Username: username,
		Password: password,
	})

	if err != nil {
		panic(err)
	}

	for {

		// Get all user RC
		RC := profil.GetUserRC()
		RCLen := len(RC)

		// Getting App Metric
		for i := 0; i < RCLen; i++ {
			replicas, err := profil.GetReplicas(RC[i].Namespace, RC[i].Name)

			if err != nil {
				panic(err)
			}
			fmt.Println(replicas)

			// Check Resposne time & Label & Save WPI
			var responseDay, response10Min float64
			if responseDay, err = profil.GetProfilAvg(influxDB, RC[i].Namespace, RC[i].Name, "rtime", "1d"); err != nil {
				panic(err)
				log.Println(err)
			}
			if response10Min, err = profil.GetProfilAvg(influxDB, RC[i].Namespace, RC[i].Name, "rtime", "5m"); err != nil {
				panic(err)
				log.Println(err)
			}
			// Floor
			responseDay = math.Floor(responseDay)
			response10Min = math.Floor(response10Min)
			fmt.Println("D", responseDay, " 10M", response10Min)
			//metrics := profil.GetAppResource(RC[i].Namespace, RC[i].Name)
			var cpu10Min float64
			if cpu10Min, err = profil.GetProfilAvg(influxDB, RC[i].Namespace, RC[i].Name, "cpu", "5m"); err != nil {
				panic(err)
				log.Println(err)
			}
			fmt.Println("CPU ", cpu10Min)
			qRepSpread, err := profil.QueryDB(influxDB, fmt.Sprint("SELECT spread(replicas) FROM "+RC[i].Namespace+" WHERE time > now() - 5m"))
			if err != nil {
				log.Fatal(err)
			}
			repSpread, err := strconv.ParseFloat(fmt.Sprint(qRepSpread[0].Series[0].Values[0][1]), 32)

			if repSpread < 1 {

				if cpu10Min > 70 {

					//fmt.Println("Response check")
					//if response10Min > responseDay { // TODO:Need to check WPI too
					// Save WPI
					fmt.Println("Scale+1")
					//if err := profil.WriteRPI(influxDB, RC[i].Namespace, RC[i].Name, metrics.Request, replicas); err != nil {
					//	panic(err)
					//	log.Println(err)
					//}
					// Scale +1
					// TODO: Limit
					if replicas < 10 {
						if _, err := thoth.ScaleOutViaCli(replicas+1, RC[i].Namespace, RC[i].Name); err != nil {
							panic(err)
						}
					}
					//	}
				} else if replicas > 1 {
					// = rpi/replicas
					//var rpiMax float64
					//if rpiMax, err = profil.GetAvgRPI(influxDB, RC[i].Namespace, RC[i].Name); err != nil {
					//	rpiMax = -1
					// TODO:Handler
					//panic(err)
					//}
					//fmt.Println("WPI", rpiMax)
					//if rpiMax > 0 {
					//	minReplicas := int(metrics.Request / int64(rpiMax)) // TODO: Ceil?

					//	if minReplicas < replicas {
					// Scale -1
					fmt.Println("Scale-1")
					if _, err := thoth.ScaleOutViaCli(replicas-1, RC[i].Namespace, RC[i].Name); err != nil {
						panic(err)
					}
					//	}
					//}
				}
			}
		}
		// -----Prediction-----
		// Normalize
		// Run (Predict)
		// Label
		//runFann()
		//-----------
		fmt.Println("Sleep TODO:Change to 5 Min")
		time.Sleep(1 * time.Minute)
	}
}

func runFann() {
	const num_layers = 3
	const num_neurons_hidden = 10
	const desired_error = 0.001

	train_data := fann.ReadTrainFromFile("file.csv")
	//	test_data := fann.ReadTrainFromFile("../../datasets/robot.test")

	var momentum float32
	//	for momentum = 0.0; momentum < 0.7; momentum += 0.1 {
	fmt.Printf("============= momentum = %f =============\n", momentum)

	ann := fann.CreateStandard(num_layers, []uint32{train_data.GetNumInput(), num_neurons_hidden, train_data.GetNumOutput()})
	/*
		ann.SetTrainingAlgorithm(fann.TRAIN_INCREMENTAL)
				ann.SetLearningMomentum(momentum)
	*/
	ann.SetActivationFunctionHidden(fann.SIGMOID_SYMMETRIC)
	ann.SetActivationFunctionOutput(fann.SIGMOID_SYMMETRIC)
	ann.TrainOnData(train_data, 2000, 500, desired_error)

	fmt.Printf("MSE error on train data: %f\n", ann.TestData(train_data))
	//	fmt.Printf("MSE error on test data : %f\n", ann.TestData(test_data))

	ann.Destroy()
	//	}

	train_data.Destroy()
	//test_data.Destroy()
}

func annGoBrain() {
	// Open the file.
	f, _ := os.Open("file.csv")
	r := csv.NewReader(f)
	trainSet := [][][]float64{}
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		record := [][]float64{
			{parse(rec[5]), parse(rec[6]), parse(rec[7]), parse(rec[8]), parse(rec[9])}, {parse(rec[10])},
		}
		fmt.Println(strings.Join(rec[5:], " "))
		fmt.Println(record)
		trainSet = append(trainSet, record)
	}
	//-----------
	rand.Seed(0)

	// instantiate the Feed Forward
	ff := &gobrain.FeedForward{}

	// initialize the Neural Network;
	// inputs, hidden nodes and output.
	ff.Init(5, 10, 1)

	// train the network using the XOR patterns,1000 epochs,learning rate 0.6,momentum factor 0.4,receive reports about error
	ff.Train(trainSet, 1000, 0.6, 0.4, true)

	//Test

	//	ff.Test([][][]float64{{{40, 70, 2, 1965, 13}, {1}}})
}

func parse(str string) float64 {
	f, _ := strconv.ParseFloat(str, 64)
	return f
}
