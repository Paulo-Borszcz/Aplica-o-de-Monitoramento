package network

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/shirou/gopsutil/v3/net"
)

type Info struct {
	Interfaces        []Interface         `json:"interfaces"`
	Connections       []Connection        `json:"connections"`
	DNSServers        []string            `json:"dns_servers"`
	PublicIP          string              `json:"public_ip"`
	AdvancedInfo      AdvancedNetworkInfo `json:"advanced_info"`
}

type Interface struct {
	Name        string   `json:"name"`
	MACAddress  string   `json:"mac_address"`
	IPAddresses []string `json:"ip_addresses"`
	Status      string   `json:"status"`
	Speed       uint64   `json:"speed_mbps"`
	BytesSent   uint64   `json:"bytes_sent"`
	BytesRecv   uint64   `json:"bytes_recv"`
}

type Connection struct {
	LocalAddr  string `json:"local_address"`
	LocalPort  int    `json:"local_port"`
	RemoteAddr string `json:"remote_address"`
	RemotePort int    `json:"remote_port"`
	State      string `json:"state"`
	Process    string `json:"process"`
}

func Collect() Info {
	var info Info
	var err error

	info.Interfaces, err = getNetworkInterfaces()
	if err != nil {
		log.Printf("Erro ao coletar informações das interfaces de rede: %v", err)
	}

	info.Connections, err = getNetworkConnections()
	if err != nil {
		log.Printf("Erro ao coletar conexões de rede: %v", err)
	}

	info.DNSServers, err = getDNSServers()
	if err != nil {
		log.Printf("Erro ao coletar servidores DNS: %v", err)
	}

	info.PublicIP, err = getPublicIP()
	if err != nil {
		log.Printf("Erro ao coletar IP público: %v", err)
	}

	info.AdvancedInfo, err = GetAdvancedNetworkInfo()
	if err != nil {
		log.Printf("Erro ao coletar informações avançadas de rede: %v", err)
	}

	return info
}

func getNetworkInterfaces() ([]Interface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var networkInterfaces []Interface

	for _, iface := range interfaces {
		addrs := iface.Addrs

		var ipAddresses []string
		for _, addr := range addrs {
			ipAddresses = append(ipAddresses, addr.Addr)
		}

		// Get IO counters for this interface
		ioCounters, err := net.IOCounters(true)
		if err != nil {
			log.Printf("Erro ao obter contadores de IO para interface %s: %v", iface.Name, err)
			continue
		}

		var bytesSent, bytesRecv uint64
		for _, counter := range ioCounters {
			if counter.Name == iface.Name {
				bytesSent = counter.BytesSent
				bytesRecv = counter.BytesRecv
				break
			}
		}

		networkInterfaces = append(networkInterfaces, Interface{
			Name:        iface.Name,
			MACAddress:  iface.HardwareAddr,
			IPAddresses: ipAddresses,
			Status:      strings.Join(iface.Flags, ", "),
			Speed:       uint64(iface.MTU),
			BytesSent:   bytesSent,
			BytesRecv:   bytesRecv,
		})
	}

	return networkInterfaces, nil
}

func getNetworkConnections() ([]Connection, error) {
	connections, err := net.Connections("tcp")
	if err != nil {
		return nil, err
	}

	var networkConnections []Connection

	for _, conn := range connections {
		networkConnections = append(networkConnections, Connection{
			LocalAddr:  conn.Laddr.IP,
			LocalPort:  int(conn.Laddr.Port),
			RemoteAddr: conn.Raddr.IP,
			RemotePort: int(conn.Raddr.Port),
			State:      conn.Status,
			Process:    string(conn.Pid),
		})
	}

	return networkConnections, nil
}

func getDNSServers() ([]string, error) {
	file, err := ioutil.ReadFile("/etc/resolv.conf")
	if err != nil {
		return nil, err
	}

	var servers []string
	lines := strings.Split(string(file), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "nameserver") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				servers = append(servers, fields[1])
			}
		}
	}

	return servers, nil
}

func getPublicIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(ip), nil
}

