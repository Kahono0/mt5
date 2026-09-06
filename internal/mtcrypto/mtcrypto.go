package mtcrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"

	"github.com/kahono0/mt5/internal/mthelpers"
)

// zeroIV is a 16-byte all-zero IV used for AES-CBC operations.
var zeroIV = make([]byte, 16)

// InitKey creates an AES cipher.Block from raw key bytes. Keys longer
// than 32 bytes are truncated; keys shorter than 16 bytes are rejected.
func InitKey(key []byte) (cipher.Block, error) {
	if len(key) > 32 {
		key = key[:32]
	} else if len(key) < 16 {
		return nil, fmt.Errorf("key must be at least 16 bytes, got %d bytes", len(key))
	}
	return aes.NewCipher(key)
}

// InitKeyFromString creates an AES cipher.Block by deobfuscating and
// hex-decoding the given key string.
func InitKeyFromString(key string) (cipher.Block, error) {
	obs := mthelpers.Obfuscation(key)
	hexBuf, err := mthelpers.HexToBuffer(obs)
	if err != nil {
		return nil, err
	}
	return InitKey(hexBuf)
}

// Encrypt encrypts plaintext using AES-CBC with a zero IV.
// The plaintext is padded with zero bytes to a multiple of the block size.
func Encrypt(block cipher.Block, plaintext []byte) ([]byte, error) {
	blockSize := block.BlockSize()
	plaintext = padZero(plaintext, blockSize)

	mode := cipher.NewCBCEncrypter(block, zeroIV)

	ciphertext := make([]byte, len(plaintext))
	mode.CryptBlocks(ciphertext, plaintext)
	return ciphertext, nil
}

// Decrypt decrypts ciphertext using AES-CBC with a zero IV.
func Decrypt(block cipher.Block, ciphertext []byte) ([]byte, error) {
	blockSize := block.BlockSize()
	if len(ciphertext)%blockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of block size")
	}

	mode := cipher.NewCBCDecrypter(block, zeroIV)

	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	return plaintext, nil
}

// padZero pads data with zero bytes to a multiple of blockSize.
func padZero(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	if padding == blockSize {
		return data
	}
	padded := make([]byte, len(data)+padding)
	copy(padded, data)
	return padded
}
