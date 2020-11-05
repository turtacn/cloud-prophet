package profil

import (
	"fmt"
	influx "github.com/influxdata/influxdb1-client/v2"
	"strconv"
)

func GetProfilAvg(conn influx.Client, region, instanceId, metric, timeLength string) (float64, error) {
	query := fmt.Sprint("SELECT MEAN(" + metric + ") FROM " + region + " WHERE app =~ /" + instanceId + "/ AND time > now() - " + timeLength)
	dbResult, err := QueryDB(conn, query)
	if err != nil {
		panic(err)
		return -1.0, err
	}
	//	fmt.Println(dbResult)
	return strconv.ParseFloat(fmt.Sprint(dbResult[0].Series[0].Values[0][1]), 32)
}

func GetProfilLast(conn influx.Client, region, instanceId, timeLength string) map[string]int64 {
	//mean(cpu),mean(memory),mean(request),mean(response),stddev(code5xx)
	query := fmt.Sprint("SELECT mean(cpu) as cpu,mean(memory) as memory,mean(rps) as rps,mean(rtime) as rtime,stddev(r2xx) as r2xx,stddev(r5xx) as r5xx,last(replicas) as replicas FROM " + region + " WHERE app =~ /" + instanceId + "/ AND time > now() - " + timeLength)
	dbResult, err := QueryDB(conn, query)
	if err != nil {
		return nil
	}
	res := make(map[string]int64)
	dbRes1, _ := strconv.ParseFloat(fmt.Sprint(dbResult[0].Series[0].Values[0][1]), 32)
	dbRes2, _ := strconv.ParseFloat(fmt.Sprint(dbResult[0].Series[0].Values[0][2]), 32)
	dbRes3, _ := strconv.ParseFloat(fmt.Sprint(dbResult[0].Series[0].Values[0][3]), 32)
	dbRes4, _ := strconv.ParseFloat(fmt.Sprint(dbResult[0].Series[0].Values[0][4]), 32)
	dbRes5, _ := strconv.ParseFloat(fmt.Sprint(dbResult[0].Series[0].Values[0][5]), 32)
	dbRes6, _ := strconv.ParseFloat(fmt.Sprint(dbResult[0].Series[0].Values[0][6]), 32)
	dbRes7, _ := strconv.ParseFloat(fmt.Sprint(dbResult[0].Series[0].Values[0][7]), 32)

	res[fmt.Sprint(dbResult[0].Series[0].Columns[1])] = int64(dbRes1)
	res[fmt.Sprint(dbResult[0].Series[0].Columns[2])] = int64(dbRes2)
	res[fmt.Sprint(dbResult[0].Series[0].Columns[3])] = int64(dbRes3)
	res[fmt.Sprint(dbResult[0].Series[0].Columns[4])] = int64(dbRes4)
	res[fmt.Sprint(dbResult[0].Series[0].Columns[5])] = int64(dbRes5)
	res[fmt.Sprint(dbResult[0].Series[0].Columns[6])] = int64(dbRes6)
	res[fmt.Sprint(dbResult[0].Series[0].Columns[7])] = int64(dbRes7)
	return res
}

func GetProfilStdLast(conn influx.Client, region, instanceId, timeLength string) map[string]float64 {
	//mean(cpu),mean(memory),mean(request),mean(response),stddev(code5xx)
	query := fmt.Sprint("SELECT STDDEV(cpu) as cpu,STDDEV(memory) as memory,STDDEV(rps) as rps,STDDEV(rtime) as rtime,STDDEV(r2xx) as r2xx,STDDEV(r5xx) as r5xx,last(replicas) as replicas FROM " + region + " WHERE app =~ /" + instanceId + "/ AND time > now() - " + timeLength)
	dbResult, err := QueryDB(conn, query)
	if err != nil {
		return nil
	}
	res := make(map[string]float64)
	dbRes1, _ := strconv.ParseFloat(fmt.Sprint(dbResult[0].Series[0].Values[0][1]), 32)
	dbRes2, _ := strconv.ParseFloat(fmt.Sprint(dbResult[0].Series[0].Values[0][2]), 32)
	dbRes3, _ := strconv.ParseFloat(fmt.Sprint(dbResult[0].Series[0].Values[0][3]), 32)
	dbRes4, _ := strconv.ParseFloat(fmt.Sprint(dbResult[0].Series[0].Values[0][4]), 32)
	dbRes5, _ := strconv.ParseFloat(fmt.Sprint(dbResult[0].Series[0].Values[0][5]), 32)
	dbRes6, _ := strconv.ParseFloat(fmt.Sprint(dbResult[0].Series[0].Values[0][6]), 32)
	dbRes7, _ := strconv.ParseFloat(fmt.Sprint(dbResult[0].Series[0].Values[0][7]), 32)

	res[fmt.Sprint(dbResult[0].Series[0].Columns[1])] = dbRes1
	res[fmt.Sprint(dbResult[0].Series[0].Columns[2])] = dbRes2
	res[fmt.Sprint(dbResult[0].Series[0].Columns[3])] = dbRes3
	res[fmt.Sprint(dbResult[0].Series[0].Columns[4])] = dbRes4
	res[fmt.Sprint(dbResult[0].Series[0].Columns[5])] = dbRes5
	res[fmt.Sprint(dbResult[0].Series[0].Columns[6])] = dbRes6
	res[fmt.Sprint(dbResult[0].Series[0].Columns[7])] = dbRes7
	return res
}

func WriteRPI(conn influx.Client, region, instanceId string, request int64, replicas int) error {

	tags := map[string]string{
		"app": instanceId,
	}

	metrics := map[string]interface{}{
		"rps":      request,
		"replicas": replicas,
		"rpi":      request / int64(replicas), // TODO:remove?
	}
	fmt.Println(metrics)
	if err := WritePoints(conn, region+"_rpi", "s", tags, metrics); err != nil {
		return err
	}
	return nil
}

func GetAvgRPI(conn influx.Client, region, instanceId string) (float64, error) {
	query := fmt.Sprint("SELECT MEAN(rpi) FROM " + region + "_rpi WHERE app =~ /" + instanceId + "/ limit 20")
	dbResult, err := QueryDB(conn, query)
	if err != nil {
		panic(err)
		return -1.0, err
	}
	//	fmt.Println(dbResult)
	return strconv.ParseFloat(fmt.Sprint(dbResult[0].Series[0].Values[0][1]), 32)
}
