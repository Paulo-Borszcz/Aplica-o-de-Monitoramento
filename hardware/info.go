package hardware

import (
	"fmt"
	"log"
	"time"

	"github.com/jaypipes/ghw"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

type Info struct {
	CPU           CPUInfo     `json:"cpu"`
	Memory        MemoryInfo  `json:"memory"`
	Disk          []DiskInfo  `json:"disk"`
	GPU           []GPUInfo   `json:"gpu"`
	Motherboard   Motherboard `json:"motherboard"`
	BIOS          BIOSInfo    `json:"bios"`
	USB           []USBDevice `json:"usb_devices"`
}

type CPUInfo struct {
	Model       string   `json:"model"`
	Cores       int      `json:"cores"`
	Threads     int      `json:"threads"`
	Frequency   float64  `json:"frequency_ghz"`
	Temperature float64  `json:"temperature_celsius"`
	Usage       float64  `json:"usage_percent"`
}

type MemoryInfo struct {
	Total        uint64  `json:"total_bytes"`
	Used         uint64  `json:"used_bytes"`
	Free         uint64  `json:"free_bytes"`
	UsagePercent float64 `json:"usage_percent"`
}

type DiskInfo struct {
	Device       string  `json:"device"`
	Type         string  `json:"type"`
	Total        uint64  `json:"total_bytes"`
	Used         uint64  `json:"used_bytes"`
	Free         uint64  `json:"free_bytes"`
	UsagePercent float64 `json:"usage_percent"`
}

type GPUInfo struct {
	Model       string  `json:"model"`
	Memory      uint64  `json:"memory_bytes"`
	Temperature float64 `json:"temperature_celsius"`
	Usage       float64 `json:"usage_percent"`
}

type Motherboard struct {
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	SerialNumber string `json:"serial_number"`
}

type BIOSInfo struct {
	Vendor      string `json:"vendor"`
	Version     string `json:"version"`
	ReleaseDate string `json:"release_date"`
}

type USBDevice struct {
	Name         string `json:"name"`
	VendorID     string `json:"vendor_id"`
	ProductID    string `json:"product_id"`
	SerialNumber string `json:"serial_number"`
}

func Collect() Info {
	var info Info
	var err error

	info.CPU, err = getCPUInfo()
	if err != nil {
		log.Printf("Erro ao coletar informações da CPU: %v", err)
	}

	info.Memory, err = getMemoryInfo()
	if err != nil {
		log.Printf("Erro ao coletar informações da memória: %v", err)
	}

	info.Disk, err = getDiskInfo()
	if err != nil {
		log.Printf("Erro ao coletar informações do disco: %v", err)
	}

	info.GPU, err = getGPUInfo()
	if err != nil {
		log.Printf("Erro ao coletar informações da GPU: %v", err)
	}

	info.Motherboard, err = getMotherboardInfo()
	if err != nil {
		log.Printf("Erro ao coletar informações da placa-mãe: %v", err)
	}

	info.BIOS, err = getBIOSInfo()
	if err != nil {
		log.Printf("Erro ao coletar informações da BIOS: %v", err)
	}

	info.USB, err = getUSBInfo()
	if err != nil {
		log.Printf("Erro ao coletar informações dos dispositivos USB: %v", err)
	}

	return info
}

func getCPUInfo() (CPUInfo, error) {
	cpuInfo, err := cpu.Info()
	if err != nil {
		return CPUInfo{}, err
	}

	if len(cpuInfo) == 0 {
		return CPUInfo{}, fmt.Errorf("nenhuma informação de CPU encontrada")
	}

	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		log.Printf("Erro ao obter uso da CPU: %v", err)
	}

	temperature := 0.0

	return CPUInfo{
		Model:       cpuInfo[0].ModelName,
		Cores:       int(cpuInfo[0].Cores),
		Threads:     int(cpuInfo[0].Cores * 2),
		Frequency:   cpuInfo[0].Mhz / 1000,
		Temperature: temperature,
		Usage:       percent[0],
	}, nil
}

func getMemoryInfo() (MemoryInfo, error) {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return MemoryInfo{}, err
	}

	return MemoryInfo{
		Total:        memInfo.Total,
		Used:         memInfo.Used,
		Free:         memInfo.Free,
		UsagePercent: memInfo.UsedPercent,
	}, nil
}

func getDiskInfo() ([]DiskInfo, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}

	var disks []DiskInfo

	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			log.Printf("Erro ao obter uso do disco %s: %v", partition.Device, err)
			continue
		}

		disks = append(disks, DiskInfo{
			Device:       partition.Device,
			Type:         partition.Fstype,
			Total:        usage.Total,
			Used:         usage.Used,
			Free:         usage.Free,
			UsagePercent: usage.UsedPercent,
		})
	}

	return disks, nil
}

func getGPUInfo() ([]GPUInfo, error) {
	gpu, err := ghw.GPU()
	if err != nil {
		return nil, err
	}

	var gpus []GPUInfo

	for _, card := range gpu.GraphicsCards {
		gpus = append(gpus, GPUInfo{
			Model:  card.DeviceInfo.Product.Name,
			Memory: 0,
		})
	}

	return gpus, nil
}

func getMotherboardInfo() (Motherboard, error) {
	product, err := ghw.Product()
	if err != nil {
		return Motherboard{}, err
	}

	return Motherboard{
		Manufacturer: product.Vendor,
		Model:        product.Name,
		SerialNumber: product.SerialNumber,
	}, nil
}

func getBIOSInfo() (BIOSInfo, error) {
	bios, err := ghw.BIOS()
	if err != nil {
		return BIOSInfo{}, err
	}

	return BIOSInfo{
		Vendor:      bios.Vendor,
		Version:     bios.Version,
		ReleaseDate: bios.Date,
	}, nil
}

func getUSBInfo() ([]USBDevice, error) {
	option := ghw.WithChroot("/")
	block, err := ghw.Block(option)
	if err != nil {
		return nil, fmt.Errorf("error getting block info: %v", err)
	}

	var devices []USBDevice

	for _, disk := range block.Disks {
		if disk.IsRemovable {
			devices = append(devices, USBDevice{
				Name:         disk.Model,
				VendorID:     disk.Vendor,
				ProductID:    disk.Model,
				SerialNumber: disk.SerialNumber,
			})
		}
	}

	return devices, nil
}

