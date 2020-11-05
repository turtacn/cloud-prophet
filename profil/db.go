package profil

import (
	"github.com/influxdata/influxdb1-client/v2"
	"time"
)

var myDB string = "prophet"

//func main2() {
// Make client
/*
	c, _ := client.NewHTTPClient(client.HTTPConfig{
		Addr:     "http://localhost:8086",
		Username: username,
		Password: password,
	})
*/

// batch write
//	writePoints(c)
// query
//res, err := queryDB(c, fmt.Sprintf("SELECT count(busy) FROM cpu_usage"))
//if err != nil {
//	log.Fatal(err)
//}
//fmt.Print(res)
//}

func WritePoints(clnt client.Client, measurement string, precision string, tags map[string]string, fields map[string]interface{}) error {

	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  myDB,
		Precision: precision,
	})

	pt, _ := client.NewPoint(measurement, tags, fields, time.Now())
	bp.AddPoint(pt)

	return clnt.Write(bp)
}

func QueryDB(clnt client.Client, cmd string) (res []client.Result, err error) {
	q := client.Query{
		Command:  cmd,
		Database: myDB,
	}
	if response, err := clnt.Query(q); err == nil {
		if response.Error() != nil {
			return res, response.Error()
		}
		res = response.Results
	} else {
		return res, err
	}
	return res, nil
}
