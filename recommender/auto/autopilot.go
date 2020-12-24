package auto

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"github.com/turtacn/cloud-prophet/recommender/logic"
	"github.com/turtacn/cloud-prophet/recommender/model"
	vpa_types "github.com/turtacn/cloud-prophet/recommender/types"
	vpa_utils "github.com/turtacn/cloud-prophet/recommender/util"
	"k8s.io/klog"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	runOnceTimeout       = flag.Duration("runonce-timeout", time.Hour, `一次运行的最大超时时间`)
	sampleSecondInterval = flag.Int("sample-second-interval", 60, `样本的采样间隔，单位(秒)，整型`)
)

// Recommender recommend resources for certain containers, based on utilization periodically got from metrics api.
type Recommender interface {
	// RunOnce performs one iteration of recommender
	// elementid , csv data
	RunOnce(string, string)
}

type recommender struct {
	podResourceRecommender logic.PodResourceRecommender
}

// getCappedRecommendation creates a recommendation based on recommended pod
// resources, setting the UncappedTarget to the calculated recommended target
// and if necessary, capping the Target, LowerBound and UpperBound according
// to the ResourcePolicy.
func getCappedRecommendation(vpaID model.VpaID, resources logic.RecommendedPodResources,
	policy *vpa_types.PodResourcePolicy) *vpa_types.RecommendedPodResources {
	containerResources := make([]vpa_types.RecommendedContainerResources, 0, len(resources))
	for containerName, res := range resources {
		containerResources = append(containerResources, vpa_types.RecommendedContainerResources{
			ContainerName:  containerName,
			Target:         model.ResourcesAsResourceList(res.Target),
			LowerBound:     model.ResourcesAsResourceList(res.LowerBound),
			UpperBound:     model.ResourcesAsResourceList(res.UpperBound),
			UncappedTarget: model.ResourcesAsResourceList(res.Target),
		})
	}
	recommendation := &vpa_types.RecommendedPodResources{containerResources}
	cappedRecommendation, err := vpa_utils.ApplyVPAPolicy(recommendation, policy)
	if err != nil {
		klog.Errorf("Failed to apply policy for VPA %v/%v: %v", vpaID.Namespace, vpaID.VpaName, err)
		return recommendation
	}
	return cappedRecommendation
}

// make2dFloatArray makes a new 2d array of float64s based on the
// rowCount and colCount provided as arguments
func make2dFloatArray(rowCount int, colCount int) [][]float64 {
	values := make([][]float64, rowCount)
	for rowIndex := range values {
		values[rowIndex] = make([]float64, colCount)
	}

	return values
}

// stringValuesToFloats converts a 2d array of strings into a 2d array
// of float64s.
func stringValuesToFloats(stringValues [][]string) ([][]float64, error) {
	values := make2dFloatArray(len(stringValues), len(stringValues[0]))

	for rowIndex, _ := range values {
		for colIndex, _ := range values[rowIndex] {
			var err error = nil

			trimString :=
				strings.TrimSpace(stringValues[rowIndex][colIndex])

			values[rowIndex][colIndex], err =
				strconv.ParseFloat(trimString, 64)

			if err != nil {
				fmt.Println(err)
				return values, err
			}
		}
	}

	return values, nil
}

// ReadFromCsv will read the csv file at filePath and return its
// contents as a 2d array of floats
func ReadFromCsv(filePath string) ([][]float64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(file)
	stringValues, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	values, err := stringValuesToFloats(stringValues)
	if err != nil {
		return nil, err
	}

	return values, nil
}

func (r *recommender) RunOnce(element, csvfile string) {

	ctx := context.Background()
	ctx, cancelFunc := context.WithDeadline(ctx, time.Now().Add(*runOnceTimeout))
	defer cancelFunc()

	klog.V(0).Infof("Recommender Run")
	// load

	var anyTime = time.Unix(0, 0)
	entityAggregateStateMap := make(model.ContainerNameToAggregateStateMap)
	entityAggregateStateMap[element] = model.NewAggregateContainerState()

	data, err := ReadFromCsv(csvfile)
	if err != nil {
		klog.Errorf("load data failed")
	}

	for elem, s := range entityAggregateStateMap {

		file, err := os.Create(fmt.Sprintf("%s-predict.csv", elem))
		checkError("Cannot create file", err)
		defer file.Close()
		writer := csv.NewWriter(file)
		defer writer.Flush()

		timestamp := anyTime
		for i := 0; i < len(data); i++ {
			d := data[i][1] / 100
			s.AddSample(&model.ContainerUsageSample{
				timestamp, model.CPUAmountFromCores(d), model.CPUAmountFromCores(1), model.ResourceCPU})

			timestamp = timestamp.Add(time.Duration(*sampleSecondInterval) * time.Second)

			resources := r.podResourceRecommender.GetRecommendedPodResources(entityAggregateStateMap)
			containerResources := make([]vpa_types.RecommendedContainerResources, 0, len(resources))
			for containerName, res := range resources {
				containerResources = append(containerResources, vpa_types.RecommendedContainerResources{
					ContainerName:  containerName,
					Target:         model.ResourcesAsResourceList(res.Target),
					LowerBound:     model.ResourcesAsResourceList(res.LowerBound),
					UpperBound:     model.ResourcesAsResourceList(res.UpperBound),
					UncappedTarget: model.ResourcesAsResourceList(res.Target),
				})
			}
			recommendation := &vpa_types.RecommendedPodResources{containerResources}

			for _, recon := range recommendation.ContainerRecommendations {

				recommendString := fmt.Sprintf("%s,%f",
					recon.Target.Cpu().AsDec().String(), d)
				klog.Info(recommendString)
				err := writer.Write([]string{recon.Target.Cpu().AsDec().String(), fmt.Sprintf("%f", d)})
				checkError("Cannot write to file", err)
			}
		}
	}

	// upodate vpa
	// gc
	// maintain checkpoint

}

func checkError(message string, err error) {
	if err != nil {
		klog.Fatalf("message:%s, error=%+v", message, err)
	}
}

// RecommenderFactory makes instances of Recommender.
type RecommenderFactory struct {
	PodResourceRecommender logic.PodResourceRecommender
	CheckpointsGCInterval  time.Duration
	UseCheckpoints         bool
}

// Make creates a new recommender instance,
// which can be run in order to provide continuous resource recommendations for containers.
func (c RecommenderFactory) Make() Recommender {
	recommender := &recommender{
		podResourceRecommender: c.PodResourceRecommender,
	}
	klog.Infof("New Recommender created %+v", recommender)
	return recommender
}

// NewRecommender creates a new recommender instance.
// Dependencies are created automatically.
// Deprecated; use RecommenderFactory instead.
func NewRecommender() Recommender {
	return RecommenderFactory{
		PodResourceRecommender: logic.CreatePodResourceRecommender(),
	}.Make()
}
