package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func EncryptJSON(jsonData []byte, hexKey string) (string, error) {
	log.Printf("Tamanho dos dados JSON: %d bytes", len(jsonData))
	log.Printf("Tamanho da chave hexadecimal: %d caracteres", len(hexKey))

	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return "", fmt.Errorf("erro ao decodificar a chave: %v", err)
	}

	log.Printf("Tamanho da chave decodificada: %d bytes", len(key))

	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return "", fmt.Errorf("tamanho de chave inválido: %d bytes. Deve ser 16, 24 ou 32 bytes", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("erro ao criar cifra: %v", err)
	}

	ciphertext := make([]byte, aes.BlockSize+len(jsonData))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", fmt.Errorf("erro ao gerar IV: %v", err)
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], jsonData)

	encodedData := base64.URLEncoding.EncodeToString(ciphertext)
	log.Printf("Tamanho dos dados criptografados e codificados: %d caracteres", len(encodedData))

	return encodedData, nil
}

func ReadINIFile(filename string) (map[string]string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	config := make(map[string]string)
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		config[key] = value
	}

	return config, nil
}

