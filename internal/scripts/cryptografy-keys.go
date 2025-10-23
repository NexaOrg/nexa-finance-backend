package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
)

func main() {
	// 1. Gere a chave RSA-2048
	privRSA, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}

	// 2. Serializa como PKCS#1 PEM
	privBytes := x509.MarshalPKCS1PrivateKey(privRSA)
	privPem := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	})
	os.WriteFile("private_rsa.pem", privPem, 0600)

	// 3. Serializa a public key como PKIX PEM
	pubBytes, err := x509.MarshalPKIXPublicKey(&privRSA.PublicKey)
	if err != nil {
		panic(err)
	}
	pubPem := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})
	os.WriteFile("public_rsa.pem", pubPem, 0644)
}
