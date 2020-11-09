package main

import (
	//"bufio"
	"encoding/csv"
	"fmt"
	"github.com/goml/gobrain"
	"io"
	"math/rand"
	"os"

	"log"
	//	"strings"
	"github.com/turtacn/cloud-prophet/util"
)

func main() {
	// Open the file.
	f, _ := os.Open("file.csv")
	r := csv.NewReader(f)
	r.Comma = ' '
	trainSet := [][][]float64{}
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(rec)
		continue

		record := [][]float64{
			{
				util.ParserString2Float(rec[0]), util.ParserString2Float(rec[1]),
				util.ParserString2Float(rec[2]), util.ParserString2Float(rec[3]),
				util.ParserString2Float(rec[4]),
			},
			{
				util.ParserString2Float(rec[5]),
			},
		}
		trainSet = append(trainSet, record)
	}
	fmt.Println(trainSet)
	//-----------
	rand.Seed(0)

	// instantiate the Feed Forward
	ff := &gobrain.FeedForward{}

	// initialize the Neural Network;
	// the networks structure will contain:
	// inputs, hidden nodes and output.
	ff.Init(5, 5, 1)

	// train the network using the XOR patterns
	// the training will run for 1000 epochs
	// the learning rate is set to 0.6 and the momentum factor to 0.4
	// use true in the last parameter to receive reports about the learning error
	ff.Train(trainSet, 1000, 0.6, 0.4, true)

	//Test

	ff.Test([][][]float64{{{40, 70, 2, 1965, 13}, {1}}})
	ff.Test([][][]float64{{{1.54, 41, 1, 150, 5}, {0}}})
}
