package learn

import (
	"fmt"
	"github.com/white-pony/go-fann"
	"os"
)

/*
type TrainingData struct {
	input  []fann.FannType
	output []fann.FannType
}
*/
type Neural struct {
	InputNodes  uint32
	OutputNodes uint32
	Layers      uint
	HiddenNodes uint32
	DesiresErr  float32
	Ann         *fann.Ann
	Avg         map[string]int64
	StdDev      map[string]float64
}

func (n *Neural) Init(path string) {
	// FANN Init
	if _, err := os.Stat("fann.dat"); err == nil {
		// path/to/whatever exists
		fmt.Println("Load fann.dat")
		n.Ann = fann.CreateFromFile("fann.dat")
	} else {
		fmt.Println("Init FANN")
		n.InputNodes = 6
		n.OutputNodes = 3
		n.HiddenNodes = 36
		n.Layers = 3
		n.Ann = fann.CreateStandard(n.Layers, []uint32{n.InputNodes, n.HiddenNodes, n.OutputNodes})
		n.Ann.SetActivationFunctionHidden(fann.SIGMOID_SYMMETRIC)
		n.Ann.SetActivationFunctionOutput(fann.SIGMOID_SYMMETRIC)
	}
}

func (n *Neural) Save(path string) {
	n.Ann.Save(path)
}

func (n Neural) ZScore(field string, value int64) fann.FannType {
	if n.StdDev[field] != 0 {
		return fann.FannType((float64(value - n.Avg[field])) / n.StdDev[field])
	} else {
		return fann.FannType(0)
	}
}

func (n Neural) Run(metric map[string]int64) int {
	input := n.createInput(metric)
	predict := n.Ann.Run(input)
	fmt.Println("Predict: ", predict)
	// +1,0,-0
	// Find Max output nodes
	actionIndex := 0
	for i := 0; i < 3; i++ {
		if predict[actionIndex] < predict[i] {
			actionIndex = i
		}
	}
	if actionIndex == 0 {
		fmt.Println("Predict Result +1")
		return 1
	} else if actionIndex == 2 {
		fmt.Println("Predict Result -1")
		return -1
	} else { // TODO : This may cause bug
		fmt.Println("Predict Result 0")
		return 0
	}
}
func (n Neural) createInput(metric map[string]int64) []fann.FannType {
	return []fann.FannType{
		n.ZScore("cpu", metric["cpu"]),
		n.ZScore("mem", metric["mem"]),
		n.ZScore("rps", metric["rps"]),
		n.ZScore("rtime", metric["rtime"]),
		n.ZScore("r5xx", metric["r5xx"]),
		n.ZScore("replicas", metric["replicas"]),
	}
}

func (n *Neural) Train(metric map[string]int64, class float64) {

	input := n.createInput(metric)
	var plus fann.FannType = 0
	var stay fann.FannType = 0
	var minus fann.FannType = 0
	if class == 1 {
		plus = 1
	} else if class == 0 {
		stay = 1
	} else if class == -1 {
		minus = 1
	}
	output := []fann.FannType{plus, stay, minus}
	n.Ann.Train(input, output)
	fmt.Println("Train Data ", input, output)
	fmt.Printf("MSE : %f\n", n.Ann.GetMSE())
}
