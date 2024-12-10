package performance

import (
	"log"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

type Metrics struct {
    CPUUsage     float64         `json:"cpu_usage_percent"`
    MemoryUsage  float64         `json:"memory_usage_percent"`
    DiskIO       DiskIOMetrics   `json:"disk_io"`
    NetworkIO    NetworkIOMetrics `json:"network_io"`
    SystemLoad   []float64       `json:"system_load"`
    Temperatures Temperatures    `json:"temperatures"`
}

type DiskIOMetrics struct {
    ReadBytes  uint64 `json:"read_bytes_per_sec"`
    WriteBytes uint64 `json:"write_bytes_per_sec"`
    IOPSRead   uint64 `json:"iops_read"`
    IOPSWrite  uint64 `json:"iops_write"`
}

type NetworkIOMetrics struct {
    BytesSent   uint64 `json:"bytes_sent_per_sec"`
    BytesRecv   uint64 `json:"bytes_recv_per_sec"`
    PacketsSent uint64 `json:"packets_sent_per_sec"`
    PacketsRecv uint64 `json:"packets_recv_per_sec"`
}

type Temperatures struct {
    CPU  float64   `json:"cpu_celsius"`
    GPU  float64   `json:"gpu_celsius"`
    Disk []float64 `json:"disk_celsius"`
}

func Collect() Metrics {
	var metrics Metrics
	var err error

	metrics.CPUUsage, err = getCPUUsage()
	if err != nil {
		log.Printf("Erro ao coletar uso da CPU: %v", err)
	}

	metrics.MemoryUsage, err = getMemoryUsage()
	if err != nil {
		log.Printf("Erro ao coletar uso de memória: %v", err)
	}

	metrics.DiskIO, err = getDiskIO()
	if err != nil {
		log.Printf("Erro ao coletar I/O de disco: %v", err)
	}

	metrics.NetworkIO, err = getNetworkIO()
	if err != nil {
		log.Printf("Erro ao coletar I/O de rede: %v", err)
	}

	metrics.SystemLoad, err = getSystemLoad()
	if err != nil {
		log.Printf("Erro ao coletar carga do sistema: %v", err)
	}

	metrics.Temperatures, err = getTemperatures()
	if err != nil {
		log.Printf("Erro ao coletar temperaturas: %v", err)
	}

	return metrics
}

func getCPUUsage() (float64, error) {
	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0, err
	}
	return percent[0], nil
}

func getMemoryUsage() (float64, error) {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return 0, err
	}
	return memInfo.UsedPercent, nil
}

func getDiskIO() (DiskIOMetrics, error) {
	before, err := disk.IOCounters()
	if err != nil {
		return DiskIOMetrics{}, err
	}

	time.Sleep(time.Second)

	after, err := disk.IOCounters()
	if err != nil {
		return DiskIOMetrics{}, err
	}

	var totalReadBytes, totalWriteBytes, totalReadCount, totalWriteCount uint64

	for device, afterStat := range after {
		beforeStat, ok := before[device]
		if !ok {
			continue
		}
		totalReadBytes += afterStat.ReadBytes - beforeStat.ReadBytes
		totalWriteBytes += afterStat.WriteBytes - beforeStat.WriteBytes
		totalReadCount += afterStat.ReadCount - beforeStat.ReadCount
		totalWriteCount += afterStat.WriteCount - beforeStat.WriteCount
	}

	return DiskIOMetrics{
		ReadBytes:  totalReadBytes,
		WriteBytes: totalWriteBytes,
		IOPSRead:   totalReadCount,
		IOPSWrite:  totalWriteCount,
	}, nil
}

func getNetworkIO() (NetworkIOMetrics, error) {
	before, err := net.IOCounters(false)
	if err != nil {
		return NetworkIOMetrics{}, err
	}

	time.Sleep(time.Second)

	after, err := net.IOCounters(false)
	if err != nil {
		return NetworkIOMetrics{}, err
	}

	return NetworkIOMetrics{
		BytesSent:   after[0].BytesSent - before[0].BytesSent,
		BytesRecv:   after[0].BytesRecv - before[0].BytesRecv,
		PacketsSent: after[0].PacketsSent - before[0].PacketsSent,
		PacketsRecv: after[0].PacketsRecv - before[0].PacketsRecv,
	}, nil
}

func getSystemLoad() ([]float64, error) {
	loadAvg, err := load.Avg()
	if err != nil {
		return nil, err
	}
	return []float64{loadAvg.Load1, loadAvg.Load5, loadAvg.Load15}, nil
}

func getTemperatures() (Temperatures, error) {
	temps, err := host.SensorsTemperatures()
	if err != nil {
		return Temperatures{}, err
	}

	var cpuTemp, gpuTemp float64
	var diskTemps []float64

	for _, temp := range temps {
		if strings.HasPrefix(temp.SensorKey, "coretemp") {
			cpuTemp = temp.Temperature
		} else if strings.HasPrefix(temp.SensorKey, "acpitz") {
			gpuTemp = temp.Temperature
		} else if strings.HasPrefix(temp.SensorKey, "nvme") {
			diskTemps = append(diskTemps, temp.Temperature)
		}
	}

	return Temperatures{
		CPU:  cpuTemp,
		GPU:  gpuTemp,
		Disk: diskTemps,
	}, nil
}

