package test

import (
	"bufio"
	"encoding/csv"
	"io"
	"k8s.io/klog"
	"os"
	"strconv"
)

type JvirtHost struct {
	Aid                string `json:"aid"`                  //0
	Id                 string `json:"id"`                   //1
	ClusterId          string `json:"cluster_id"`           //2
	Tag                string `json:"tag"`                  //3
	HostName           string `json:"host_name"`            //4
	Az                 string `json:"az"`                   //5
	DataCenter         string `json:"data_center"`          //6
	Rack               string `json:"rack"`                 //7
	Machine            string `json:"machine"`              //8
	ServiceType        string `json:"service_type"`         //9
	HostIp             string `json:"host_ip"`              //10
	State              string `json:"state"`                //11
	ErrorCount         string `json:"error_count"`          //12
	CreatedTime        string `json:"created_time"`         //13
	UpdateTime         string `json:"update_time"`          //14
	Status             string `json:"status"`               //15
	RackId             string `json:"rack_id"`              //16
	HostGroupId        string `json:"host_group_id"`        //17
	DisabledAppId      string `json:"disabled_app_id"`      //18
	Reserved           string `json:"reserved"`             //19
	IsDpdk             string `json:"is_dpdk"`              //20
	XcGroup            string `json:"xc_group"`             //21
	DhId               string `json:"dh_id"`                //22
	CpuModel           string `json:"cpu_model"`            //23
	DataIp             string `json:"data_ip"`              //24
	SnicIp             string `json:"snic_ip"`              //25
	SnicDataIp         string `json:"snic_data_ip"`         //26
	SnicHostName       string `json:"snic_host_name"`       //27
	Baid               string `json:"baid"`                 //28
	Bid                string `json:"bid"`                  //29
	BclusterId         string `json:"bcluster_id"`          //30
	BhostId            string `json:"bhost_id"`             //31
	Bvcpus             string `json:"bvcpus"`               //32
	Bmemory            string `json:"bmemory"`              //33
	Bdisk              string `json:"bdisk"`                //34
	BvcpusUsed         string `json:"bvcpus_used"`          //35
	BmemoryUsed        string `json:"bmemory_used"`         //36
	BdiskUsed          string `json:"bdisk_used"`           //37
	BvcpusReserved     string `json:"bvcpus_reserved"`      //38
	BvcpusReservedInfo string `json:"bvcpus_reserved_info"` //39
	BmemoryReserved    string `json:"bmemory_reserved"`     //40
	BdiskReserved      string `json:"bdisk_reserved"`       //41
	BcpuMode           string `json:"bcpu_mode"`            //42
	BcpuallocRatio     string `json:"bcpualloc_ratio"`      //43
	BrunVms            string `json:"brun_vms"`             //44
	BcreatedTime       string `json:"bcreated_time"`        //45
	BupdateTime        string `json:"bupdate_time"`         //46
	Bstatus            string `json:"bstatus"`              //47
	Bversion           string `json:"bversion"`             //48
	BgpuCard           string `json:"bgpu_card"`            //49
	BdiskCount         string `json:"bdisk_count"`          //50
	BgpuCardRatio      string `json:"bgpu_card_ratio"`      //51
	BgpuCardUsed       string `json:"bgpu_card_used"`       //52
	BdiskCountUsed     string `json:"bdisk_count_used"`     //53
	BshareVms          string `json:"bshare_vms"`           //54
	BshareVcpus        string `json:"bshare_vcpus"`         //55
	BmixMode           string `json:"bmix_mode"`            //56
	BmemoryRatio       string `json:"bmemory_ratio"`        //57
}

type JvirtInstanceTrace struct {
	OpTime       string `json:"op_time"`
	OpAction     string `json:"op_action"`
	InstanceId   string `json:"instance_id"`
	RequestedCpu string `json:"requested_cpu"`
	RequestedMem string `json:"requested_mem"`
}

type JvirtInstance struct {
	InstanceId  string `json:"instance_id"`
	InstanceCpu string `json:"instance_cpu"`
	InstanceMem string `json:"instance_mem"`
}

func LoadHostInfo(file string) []JvirtHost {
	var hosts []JvirtHost
	csvFile, err := os.Open(file)
	if err != nil {
		klog.Fatal(err)
	}
	reader := csv.NewReader(bufio.NewReader(csvFile))
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			klog.Errorf("Read csv file line data error %v", err)
		}
		hosts = append(hosts, JvirtHost{
			Aid:                line[0+1],
			Id:                 line[1+1],
			ClusterId:          line[2+1],
			Tag:                line[3+1],
			HostName:           line[4+1],
			Az:                 line[5+1],
			DataCenter:         line[6+1],
			Rack:               line[7+1],
			Machine:            line[8+1],
			ServiceType:        line[9+1],
			HostIp:             line[10+1],
			State:              line[11+1],
			ErrorCount:         line[12+1],
			CreatedTime:        line[13+1],
			UpdateTime:         line[14+1],
			Status:             line[15+1],
			RackId:             line[16+1],
			HostGroupId:        line[17+1],
			DisabledAppId:      line[18+1],
			Reserved:           line[19+1],
			IsDpdk:             line[20+1],
			XcGroup:            line[21+1],
			DhId:               line[22+1],
			CpuModel:           line[23+1],
			DataIp:             line[24+1],
			SnicIp:             line[25+1],
			SnicDataIp:         line[26+1],
			SnicHostName:       line[27+1],
			Baid:               line[28+1],
			Bid:                line[29+1],
			BclusterId:         line[30+1],
			BhostId:            line[31+1],
			Bvcpus:             line[32+1],
			Bmemory:            line[33+1],
			Bdisk:              line[34+1],
			BvcpusUsed:         line[35+1],
			BmemoryUsed:        line[36+1],
			BdiskUsed:          line[37+1],
			BvcpusReserved:     line[38+1],
			BvcpusReservedInfo: line[39+1],
			BmemoryReserved:    line[40+1],
			BdiskReserved:      line[41+1],
			BcpuMode:           line[42+1],
			BcpuallocRatio:     line[43+1],
			BrunVms:            line[44+1],
			BcreatedTime:       line[45+1],
			BupdateTime:        line[46+1],
			Bstatus:            line[47+1],
			Bversion:           line[48+1],
			BgpuCard:           line[49+1],
			BdiskCount:         line[50+1],
			BgpuCardRatio:      line[51+1],
			BgpuCardUsed:       line[52+1],
			BdiskCountUsed:     line[53+1],
			BshareVms:          line[54+1],
			BshareVcpus:        line[55+1],
			BmixMode:           line[56+1],
			BmemoryRatio:       line[57+1],
		})
	}
	return hosts
}

func (h JvirtHost) AvailableCpu() float64 {
	a, _ := strconv.Atoi(h.Bvcpus)
	r, _ := strconv.Atoi(h.BvcpusReserved)
	rt, _ := strconv.ParseFloat(h.BcpuallocRatio, 32)
	return float64(a-r) * rt * 1000
}
func (h JvirtHost) AvailableMemory() float64 {
	a, _ := strconv.Atoi(h.Bmemory)
	r, _ := strconv.Atoi(h.BmemoryReserved)
	rt, _ := strconv.ParseFloat(h.BmemoryRatio, 32)
	return float64(a-r) * rt
}

func LoadIntanceOpsTrace(file string) []JvirtInstanceTrace {
	var opsTraces []JvirtInstanceTrace
	csvFile, err := os.Open(file)
	if err != nil {
		klog.Fatal(err)
	}
	reader := csv.NewReader(bufio.NewReader(csvFile))
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			klog.Errorf("Read csv file line data error %v", err)
		}
		opsTraces = append(opsTraces, JvirtInstanceTrace{
			OpTime:       line[0],
			OpAction:     line[1],
			InstanceId:   line[2],
			RequestedCpu: line[3],
			RequestedMem: line[4],
		})
	}
	return opsTraces
}
func (trace JvirtInstanceTrace) RequestCpu() float64 {
	a, _ := strconv.ParseFloat(trace.RequestedCpu, 32)
	return a * 1000
}
func (trace JvirtInstanceTrace) RequestMem() float64 {
	a, _ := strconv.ParseFloat(trace.RequestedMem, 32)
	return a
}

func readLines(file string) [][]string {
	var lines [][]string
	csvFile, err := os.Open(file)
	if err != nil {
		klog.Fatal(err)
	}
	reader := csv.NewReader(bufio.NewReader(csvFile))
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			klog.Errorf("Read csv file line data error %v", err)
		}
		lines = append(lines, line)
	}
	return lines
}

func FillTrace(trace, instance, newtrace string) {
	var opsTraces []JvirtInstanceTrace
	for _, line := range readLines(trace) {
		opsTraces = append(opsTraces, JvirtInstanceTrace{
			OpTime:     line[0],
			OpAction:   line[1],
			InstanceId: line[2],
		})
	}
	var instancesMap = make(map[string]JvirtInstance)

	for _, line := range readLines(instance) {
		instancesMap[line[0]] = JvirtInstance{
			InstanceId:  line[0],
			InstanceCpu: line[1],
			InstanceMem: line[2],
		}
	}
	traceFile, _ := os.Create(newtrace)
	writer := csv.NewWriter(traceFile)
	defer writer.Flush()
	for _, t := range opsTraces {
		if i, ok := instancesMap[t.InstanceId]; ok {
			t.RequestedCpu = instancesMap[i.InstanceId].InstanceCpu
			t.RequestedMem = instancesMap[i.InstanceId].InstanceMem
		} else {
			klog.Fatal("instnace not found!!!!")
		}
		writer.Write([]string{t.OpTime, t.OpAction, t.InstanceId, t.RequestedCpu, t.RequestedMem})
	}
}
