package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"
)

const charset = "ABCDEFGHJKMNPQRSTWXYZabcdefhijkmnprstwxyz2345678"

// RandomString generates a random string of specified length
func RandomString(length int) (string, error) {
	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		result[i] = charset[num.Int64()]
	}
	return string(result), nil
}

// EncryptPassword encrypts password using AES/CBC/PKCS7 to match Python implementation
func EncryptPassword(password, salt string) (string, error) {
	// Generate random prefix (64 characters)
	prefix, err := RandomString(64)
	if err != nil {
		return "", fmt.Errorf("failed to generate random prefix: %w", err)
	}

	// Generate random IV (16 bytes)
	iv, err := RandomString(16)
	if err != nil {
		return "", fmt.Errorf("failed to generate IV: %w", err)
	}

	// Prepare key and data
	key := []byte(salt)
	dataToEncrypt := []byte(prefix + password)

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Apply PKCS7 padding
	paddedData := pkcs7Pad(dataToEncrypt, aes.BlockSize)

	// Encrypt using CBC mode
	mode := cipher.NewCBCEncrypter(block, []byte(iv))
	encrypted := make([]byte, len(paddedData))
	mode.CryptBlocks(encrypted, paddedData)

	// Return base64 encoded result
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
