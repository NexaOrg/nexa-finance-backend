package security

import (
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
)

func Base64Decode(input string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(input)
}

func LoadPrivateKeyFromEnv() (interface{}, error) {
	b64 := os.Getenv("PRIVATE_KEY")
	if b64 == "" {
		return nil, fmt.Errorf("env PRIVATE_KEY não definida")
	}

	// Base64 → PEM bytes
	pemData, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, fmt.Errorf("falha no Base64: %w", err)
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, fmt.Errorf("bloco PEM inválido ou vazio")
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		// Parse PKCS#1 RSA
		rsaKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("erro ao parsear RSA: %w", err)
		}
		return rsaKey, nil

	case "PRIVATE KEY":
		// PKCS#8: pode ser ECDH ou RSA
		parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("erro PKCS#8: %w", err)
		}
		switch key := parsed.(type) {
		case *rsa.PrivateKey:
			return key, nil
		case *ecdsa.PrivateKey:
			curve := ecdh.P256()
			return curve.NewPrivateKey(key.D.Bytes())
		case *ecdh.PrivateKey:
			return key, nil
		default:
			return nil, fmt.Errorf("PKCS#8 não é RSA, ECDSA nem ECDH")
		}

	case "EC PRIVATE KEY":
		// Formato tradicional EC
		ecdsaKey, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("erro ao parsear EC: %w", err)
		}
		curve := ecdh.P256()
		return curve.NewPrivateKey(ecdsaKey.D.Bytes())

	default:
		return nil, fmt.Errorf("tipo PEM inesperado: %s", block.Type)
	}
}

func LoadPublicKeyPEMFlatString() (string, error) {
	b64 := os.Getenv("PUBLIC_KEY")
	if b64 == "" {
		return "", fmt.Errorf("variável de ambiente PUBLIC_KEY não definida")
	}

	pemBytes, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return "", fmt.Errorf("erro ao decodificar base64: %w", err)
	}

	pemStr := string(pemBytes)
	// Remove todas as quebras de linha (\n, \r\n)
	flat := strings.ReplaceAll(pemStr, "\n", "")
	flat = strings.ReplaceAll(flat, "\r", "")

	return flat, nil
}
