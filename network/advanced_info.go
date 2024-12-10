package network

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	psnet "github.com/shirou/gopsutil/v3/net"
)

type AdvancedNetworkInfo struct {
	Latency         float64           `json:"latency_ms"`
	PacketLoss      float64           `json:"packet_loss_percent"`
	DownloadSpeed   float64           `json:"download_speed_mbps"`
	UploadSpeed     float64           `json:"upload_speed_mbps"`
	RoutingTable    []RoutingEntry    `json:"routing_table"`
	DNSConfiguration DNSConfig        `json:"dns_configuration"`
	VPNStatus       string            `json:"vpn_status"`
}

type RoutingEntry struct {
	Destination string `json:"destination"`
	Gateway     string `json:"gateway"`
	Interface   string `json:"interface"`
}

type DNSConfig struct {
	Servers []string `json:"servers"`
	Domain  string   `json:"domain"`
}

func GetAdvancedNetworkInfo() (AdvancedNetworkInfo, error) {
	info := AdvancedNetworkInfo{}

	latency, err := measureLatency("8.8.8.8")
	if err == nil {
		info.Latency = latency
	} else {
		fmt.Printf("Erro ao medir latência: %v\n", err)
	}

	packetLoss, err := measurePacketLoss("8.8.8.8")
	if err == nil {
		info.PacketLoss = packetLoss
	} else {
		fmt.Printf("Erro ao medir perda de pacotes: %v\n", err)
	}

	downloadSpeed, uploadSpeed, err := measureNetworkSpeed()
	if err == nil {
		info.DownloadSpeed = downloadSpeed
		info.UploadSpeed = uploadSpeed
	} else {
		fmt.Printf("Erro ao medir velocidade de rede: %v\n", err)
	}

	routingTable, err := getRoutingTable()
	if err == nil {
		info.RoutingTable = routingTable
	} else {
		fmt.Printf("Erro ao obter tabela de roteamento: %v\n", err)
	}

	dnsConfig, err := getDNSConfiguration()
	if err == nil {
		info.DNSConfiguration = dnsConfig
	} else {
		fmt.Printf("Erro ao obter configuração de DNS: %v\n", err)
	}

	vpnStatus, err := getVPNStatus()
	if err == nil {
		info.VPNStatus = vpnStatus
	} else {
		fmt.Printf("Erro ao obter status da VPN: %v\n", err)
	}

	return info, nil
}

func measureLatency(host string) (float64, error) {
	out, err := exec.Command("ping", "-c", "4", host).Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "avg") {
			fields := strings.Fields(line)
			if len(fields) >= 4 {
				avgField := strings.Split(fields[3], "/")
				if len(avgField) >= 2 {
					return strconv.ParseFloat(avgField[1], 64)
				}
			}
		}
	}

	return 0, fmt.Errorf("não foi possível extrair a latência média")
}

func measurePacketLoss(host string) (float64, error) {
	out, err := exec.Command("ping", "-c", "10", host).Output()
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "packet loss") {
			fields := strings.Fields(line)
			for _, field := range fields {
				if strings.HasSuffix(field, "%") {
					return strconv.ParseFloat(strings.TrimSuffix(field, "%"), 64)
				}
			}
		}
	}

	return 0, fmt.Errorf("não foi possível extrair a porcentagem de perda de pacotes")
}

func measureNetworkSpeed() (float64, float64, error) {
	startCounters, err := psnet.IOCounters(false)
	if err != nil {
		return 0, 0, err
	}

	time.Sleep(5 * time.Second)

	endCounters, err := psnet.IOCounters(false)
	if err != nil {
		return 0, 0, err
	}

	bytesSent := endCounters[0].BytesSent - startCounters[0].BytesSent
	bytesRecv := endCounters[0].BytesRecv - startCounters[0].BytesRecv

	downloadSpeed := float64(bytesRecv) / 5 / 1024 / 1024 * 8
	uploadSpeed := float64(bytesSent) / 5 / 1024 / 1024 * 8

	return downloadSpeed, uploadSpeed, nil
}

func getRoutingTable() ([]RoutingEntry, error) {
	out, err := exec.Command("route", "-n").Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(out), "\n")
	var entries []RoutingEntry

	for _, line := range lines[2:] {
		fields := strings.Fields(line)
		if len(fields) >= 8 {
			entries = append(entries, RoutingEntry{
				Destination: fields[0],
				Gateway:     fields[1],
				Interface:   fields[7],
			})
		}
	}

	return entries, nil
}

func getDNSConfiguration() (DNSConfig, error) {
	config := DNSConfig{}

	out, err := exec.Command("cat", "/etc/resolv.conf").Output()
	if err != nil {
		return config, err
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "nameserver") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				config.Servers = append(config.Servers, fields[1])
			}
		} else if strings.HasPrefix(line, "domain") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				config.Domain = fields[1]
			}
		}
	}

	return config, nil
}

func getVPNStatus() (string, error) {
	interfaces, err := psnet.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		if strings.Contains(strings.ToLower(iface.Name), "tun") || strings.Contains(strings.ToLower(iface.Name), "tap") {
			return "Ativo", nil
		}
	}

	return "Inativo", nil
}

