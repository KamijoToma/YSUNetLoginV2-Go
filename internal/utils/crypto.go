package utils

import (
	"crypto/aes"
	"encoding/base64"
	"fmt"
)

// AESEncryptECB encrypts plaintext using AES-ECB with a base64-encoded key (cas-sso login)
func AESEncryptECB(keyB64, plaintext string) (string, error) {
	key, err := base64.StdEncoding.DecodeString(keyB64)
	if err != nil {
		return "", fmt.Errorf("failed to decode AES key: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	data := pkcs7Pad([]byte(plaintext), aes.BlockSize)
	encrypted := make([]byte, len(data))

	// ECB mode: encrypt each block independently
	for i := 0; i < len(data); i += aes.BlockSize {
		block.Encrypt(encrypted[i:i+aes.BlockSize], data[i:i+aes.BlockSize])
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// pkcs7Pad applies PKCS7 padding to data
func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	return append(data, padtext...)
}
