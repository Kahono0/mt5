package mt5

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
	"sync"

	"github.com/kahono0/mt5/internal/mthelpers"
)

var zeroIV = make([]byte, 16)

// DefaultObfuscatedKey is the default obfuscated AES key used by MT5 servers.
const DefaultObfuscatedKey = "13ef13b2b76dd8:5795gdcfb2fdc1ge85bf768f54773d22fff996e3ge75g5:75"

// Crypto defines the interface for encrypting and decrypting MT5 protocol
// payloads. Implementations must be safe for concurrent use.
type Crypto interface {
	Encrypt(plain []byte) ([]byte, error)
	Decrypt(cipherText []byte) ([]byte, error)
	SetKey(key []byte) error
	SetKeyFromString(key string) error
	ResetKey()
}

// Codec defines the interface for encoding commands, framing packets, and
// parsing received messages. Implementations must be safe for concurrent use.
type Codec interface {
	EncodeCommand(command uint16, body []byte) ([]byte, error)
	Pack(payload []byte) ([]byte, error)
	Unpack(frame []byte) ([]byte, error)
	ParseMessage(plain []byte) (Message, error)
}

// DefaultCrypto implements the [Crypto] interface using AES-CBC encryption
// with the standard MT5 key derivation scheme.
type DefaultCrypto struct {
	mu           sync.RWMutex
	block        cipher.Block
	defaultBlock cipher.Block
}

// NewDefaultCrypto creates a [DefaultCrypto] initialized with the given
// obfuscated key string. Use [DefaultObfuscatedKey] for the standard key.
func NewDefaultCrypto(obfuscatedKey string) (*DefaultCrypto, error) {
	block, err := initKeyFromString(obfuscatedKey)
	if err != nil {
		return nil, err
	}

	return &DefaultCrypto{
		block:        block,
		defaultBlock: block,
	}, nil
}

// Encrypt encrypts plain using AES-CBC with zero IV.
func (c *DefaultCrypto) Encrypt(plaintext []byte) ([]byte, error) {
	c.mu.RLock()
	block := c.block
	c.mu.RUnlock()

	blockSize := block.BlockSize()
	plaintext = padZero(plaintext, blockSize)

	ciphertext := make([]byte, len(plaintext))
	encrypter := cipher.NewCBCEncrypter(block, zeroIV)

	encrypter.CryptBlocks(ciphertext, plaintext)
	return ciphertext, nil
}

// Decrypt decrypts cipherText using AES-CBC with zero IV.
func (c *DefaultCrypto) Decrypt(ciphertext []byte) ([]byte, error) {
	c.mu.RLock()
	block := c.block
	c.mu.RUnlock()

	if block == nil {
		return nil, fmt.Errorf("encryption key not set")
	}

	blockSize := block.BlockSize()
	if len(ciphertext)%blockSize != 0 {
		return nil, errors.New("ciphertext is not a multiple of block size")
	}

	mode := cipher.NewCBCDecrypter(block, zeroIV)

	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	return plaintext, nil

}

// SetKey replaces the encryption key with raw key bytes.
func (c *DefaultCrypto) SetKey(key []byte) error {
	block, err := initKey(key)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.block = block
	c.mu.Unlock()
	return nil
}

// ResetKey resets the encryption key to the default MT5 key.
func (c *DefaultCrypto) ResetKey() {
	c.mu.Lock()
	c.block = c.defaultBlock
	c.mu.Unlock()
}

// SetKeyFromString replaces the encryption key by decoding an obfuscated
// key string.
func (c *DefaultCrypto) SetKeyFromString(key string) error {
	block, err := initKeyFromString(key)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.block = block
	c.mu.Unlock()
	return nil
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

// initKey creates an AES cipher.Block from raw key bytes. Keys longer
// than 32 bytes are truncated; keys shorter than 16 bytes are rejected.
func initKey(key []byte) (cipher.Block, error) {
	if len(key) > 32 {
		key = key[:32]
	} else if len(key) < 16 {
		return nil, fmt.Errorf("key must be at least 16 bytes, got %d bytes", len(key))
	}
	return aes.NewCipher(key)
}

// initKeyFromString creates an AES cipher.Block by deobfuscating and
// hex-decoding the given key string.
func initKeyFromString(key string) (cipher.Block, error) {
	obs := mthelpers.Obfuscation(key)
	hexBuf, err := mthelpers.HexToBuffer(obs)
	if err != nil {
		return nil, err
	}
	return initKey(hexBuf)
}
