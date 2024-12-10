package software

import (
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/process"
	"golang.org/x/sys/windows/registry"
)

type Info struct {
	OS               OSInfo         `json:"os"`
	Kernel           string         `json:"kernel"`
	InstalledApps    []InstalledApp `json:"installed_apps"`
	RunningProcesses []Process      `json:"running_processes"`
	SystemServices   []Service      `json:"system_services"`
}

type OSInfo struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	Architecture string `json:"architecture"`
	Hostname     string `json:"hostname"`
}

type InstalledApp struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	InstallDate string `json:"install_date"`
}

type Process struct {
	Name     string  `json:"name"`
	PID      int     `json:"pid"`
	CPUUsage float64 `json:"cpu_usage_percent"`
	MemUsage uint64  `json:"memory_usage_bytes"`
}

type Service struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

func Collect() Info {
	var info Info
	var err error

	info.OS, err = getOSInfo()
	if err != nil {
		log.Printf("Erro ao coletar informações do sistema operacional: %v", err)
	}

	info.Kernel, err = getKernelVersion()
	if err != nil {
		log.Printf("Erro ao coletar versão do kernel: %v", err)
	}

	info.InstalledApps, err = getInstalledApps()
	if err != nil {
		log.Printf("Erro ao coletar aplicativos instalados: %v", err)
	}

	info.RunningProcesses, err = getRunningProcesses()
	if err != nil {
		log.Printf("Erro ao coletar processos em execução: %v", err)
	}

	info.SystemServices, err = getSystemServices()
	if err != nil {
		log.Printf("Erro ao coletar serviços do sistema: %v", err)
	}

	return info
}

func getOSInfo() (OSInfo, error) {
	hostInfo, err := host.Info()
	if err != nil {
		return OSInfo{}, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("Erro ao obter o hostname: %v", err)
		hostname = "Desconhecido"
	}

	return OSInfo{
		Name:         hostInfo.Platform,
		Version:      hostInfo.PlatformVersion,
		Architecture: runtime.GOARCH,
		Hostname:     hostname,
	}, nil
}

func getKernelVersion() (string, error) {
	hostInfo, err := host.Info()
	if err != nil {
		return "", err
	}
	return hostInfo.KernelVersion, nil
}

func getInstalledApps() ([]InstalledApp, error) {
	var apps []InstalledApp

	if runtime.GOOS == "windows" {
		key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`, registry.ENUMERATE_SUB_KEYS)
		if err != nil {
			return nil, err
		}
		defer key.Close()

		subkeys, err := key.ReadSubKeyNames(-1)
		if err != nil {
			return nil, err
		}

		for _, subkeyName := range subkeys {
			subkey, err := registry.OpenKey(key, subkeyName, registry.QUERY_VALUE)
			if err != nil {
				continue
			}
			defer subkey.Close()

			displayName, _, _ := subkey.GetStringValue("DisplayName")
			displayVersion, _, _ := subkey.GetStringValue("DisplayVersion")
			installDate, _, _ := subkey.GetStringValue("InstallDate")

			if displayName != "" {
				apps = append(apps, InstalledApp{
					Name:        displayName,
					Version:     displayVersion,
					InstallDate: installDate,
				})
			}
		}
	} else if runtime.GOOS == "linux" {
		cmd := exec.Command("dpkg", "-l")
		output, err := cmd.Output()
		if err != nil {
			return nil, err
		}

		lines := strings.Split(string(output), "\n")
		for _, line := range lines[5:] {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				apps = append(apps, InstalledApp{
					Name:    fields[1],
					Version: fields[2],
				})
			}
		}
	}

	return apps, nil
}

func getRunningProcesses() ([]Process, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var runningProcesses []Process

	for _, p := range processes {
		name, err := p.Name()
		if err != nil {
			continue
		}

		pid := p.Pid

		cpuPercent, err := p.CPUPercent()
		if err != nil {
			cpuPercent = 0
		}

		memInfo, err := p.MemoryInfo()
		if err != nil {
			memInfo = &process.MemoryInfoStat{}
		}

		runningProcesses = append(runningProcesses, Process{
			Name:     name,
			PID:      int(pid),
			CPUUsage: cpuPercent,
			MemUsage: memInfo.RSS,
		})
	}

	return runningProcesses, nil
}

func getSystemServices() ([]Service, error) {
	var services []Service

	if runtime.GOOS == "windows" {
		cmd := exec.Command("sc", "query", "type=", "service", "state=", "all")
		output, err := cmd.Output()
		if err != nil {
			return nil, err
		}

		lines := strings.Split(string(output), "\n")
		for i := 0; i < len(lines); i++ {
			if strings.HasPrefix(lines[i], "SERVICE_NAME:") {
				name := strings.TrimSpace(strings.TrimPrefix(lines[i], "SERVICE_NAME:"))
				status := "Unknown"
				for j := i + 1; j < len(lines) && j < i+5; j++ {
					if strings.HasPrefix(lines[j], "        STATE") {
						status = strings.TrimSpace(strings.TrimPrefix(lines[j], "        STATE              :"))
						break
					}
				}
				services = append(services, Service{Name: name, Status: status})
			}
		}
	} else if runtime.GOOS == "linux" {
		cmd := exec.Command("systemctl", "list-units", "--type=service", "--all", "--no-pager", "--no-legend")
		output, err := cmd.Output()
		if err != nil {
			return nil, err
		}

		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				services = append(services, Service{
					Name:   fields[0],
					Status: fields[2],
				})
			}
		}
	}

	return services, nil
}

