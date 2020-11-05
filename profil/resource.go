package profil

import (
	/*
		"github.com/shirou/gopsutil/cpu"
		"github.com/shirou/gopsutil/disk"
	*/
	/*
		"github.com/shirou/gopsutil/mem"
		"github.com/shirou/gopsutil/net"
	*/

	"github.com/turtacn/cloud-prophet/model"
)

/**
	read node resource usage
**/
/*
func GetNodeResource(w http.ResponseWriter, r *http.Request) {
	// get this node memory
	memory, _ := mem.VirtualMemory()
	// get this node cpu percent usage
	cpu_percent, _ := cpu.CPUPercent(time.Duration(1)*time.Second, false)
	// Disk mount Point
	disk_partitions, _ := disk.DiskPartitions(true)
	// Disk usage
	var disk_usages []*disk.DiskUsageStat
	for _, disk_partition := range disk_partitions {
		if disk_partition.Mountpoint == "/" || disk_partition.Mountpoint == "/home" {
			disk_stat, _ := disk.DiskUsage(disk_partition.Device)
			disk_usages = append(disk_usages, disk_stat)
		}
	}
	// Network
	network, _ := net.NetIOCounters(false)

	// create new node obj with resource usage information
	node_metric := thoth.NodeMetric{
		Cpu:       cpu_percent,
		Memory:    memory,
		DiskUsage: disk_usages,
		Network:   network,
	}

	node_json, err := json.MarshalIndent(node_metric, "", "\t")
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Fprint(w, string(node_json))
}
*/

// podlist
func GetAllPod() error {
	return nil
}

func GetAllRunningPod() map[string]bool {
	return nil
}

func GetRunningPodStatus(region string) []string {
	return nil
}

func GetInstanceCount(region, hostIp string) int {
	return 0
}
func GetAppMetrics() *model.AppMetric {
	return nil
}
