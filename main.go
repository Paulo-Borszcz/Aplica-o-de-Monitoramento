package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"monitoramento/hardware"
	"monitoramento/network"
	"monitoramento/performance"
	"monitoramento/software"
	"monitoramento/utils"
)

type SystemInfo struct {
	Timestamp   time.Time           `json:"timestamp"`
	Hardware    hardware.Info       `json:"hardware"`
	Software    software.Info       `json:"software"`
	Network     network.Info        `json:"network"`
	Performance performance.Metrics `json:"performance"`
}

func main() {
	// Ler o arquivo .ini
	config, err := utils.ReadINIFile("config.ini")
	if err != nil {
		log.Fatalf("Erro ao ler o arquivo de configuração: %v", err)
	}

	serverAddress, ok := config["server_address"]
	if !ok {
		log.Fatalf("Endereço do servidor não encontrado no arquivo de configuração")
	}

	encryptionKey, ok := config["encryption_key"]
	if !ok {
		log.Fatalf("Chave de criptografia não encontrada no arquivo de configuração")
	}

	// Coletar informações do sistema
	info := collectSystemInfo()

	// Converter para JSON
	jsonData, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		log.Fatalf("Erro ao criar JSON: %v", err)
	}

	// Criptografar o JSON
	encryptedData, err := utils.EncryptJSON(jsonData, encryptionKey)
	if err != nil {
		log.Fatalf("Erro ao criptografar os dados: %v", err)
	}

	// Enviar dados criptografados para o servidor
	err = sendDataToServer(serverAddress, encryptedData)
	if err != nil {
		log.Fatalf("Erro ao enviar dados para o servidor: %v", err)
	}

	fmt.Println("Informações do sistema coletadas, criptografadas e enviadas com sucesso.")
}

func collectSystemInfo() SystemInfo {
	return SystemInfo{
		Timestamp:   time.Now(),
		Hardware:    hardware.Collect(),
		Software:    software.Collect(),
		Network:     network.Collect(),
		Performance: performance.Collect(),
	}
}

func sendDataToServer(serverAddress, encryptedData string) error {
	resp, err := http.Post(serverAddress, "text/plain", strings.NewReader(encryptedData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("servidor retornou status não-OK: %v", resp.Status)
	}

	return nil
}

