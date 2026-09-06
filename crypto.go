package mt5

import (
	"crypto/cipher"
	"fmt"
	"sync"

	"github.com/kahono0/mt5/internal/mtcrypto"
)

// DefaultObfuscatedKey is the default obfuscated AES key used by MT5 servers.
const DefaultObfuscatedKey = "13ef13b2b76dd8:5795gdcfb2fdc1ge85bf768f54773d22fff996e3ge75g5:75"

// Crypto defines the interface for encrypting and decrypting MT5 protocol
// payloads. Implementations must be safe for concurrent use.
type Crypto interface {
	Encrypt(plain []byte) ([]byte, error)
	Decrypt(cipherText []byte) ([]byte, error)
	SetKey(key []byte) error
	SetKeyFromString(key string) error
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
	mu    sync.RWMutex
	block cipher.Block
}

// NewDefaultCrypto creates a [DefaultCrypto] initialized with the given
// obfuscated key string. Use [DefaultObfuscatedKey] for the standard key.
func NewDefaultCrypto(obfuscatedKey string) (*DefaultCrypto, error) {
	block, err := mtcrypto.InitKeyFromString(obfuscatedKey)
	if err != nil {
		return nil, err
	}

	return &DefaultCrypto{block: block}, nil
}

// Encrypt encrypts plain using AES-CBC with zero IV.
func (c *DefaultCrypto) Encrypt(plain []byte) ([]byte, error) {
	c.mu.RLock()
	block := c.block
	c.mu.RUnlock()

	if block == nil {
		return nil, fmt.Errorf("encryption key not set")
	}

	return mtcrypto.Encrypt(block, plain)
}

// Decrypt decrypts cipherText using AES-CBC with zero IV.
func (c *DefaultCrypto) Decrypt(cipherText []byte) ([]byte, error) {
	c.mu.RLock()
	block := c.block
	c.mu.RUnlock()

	if block == nil {
		return nil, fmt.Errorf("encryption key not set")
	}

	return mtcrypto.Decrypt(block, cipherText)
}

// SetKey replaces the encryption key with raw key bytes.
func (c *DefaultCrypto) SetKey(key []byte) error {
	block, err := mtcrypto.InitKey(key)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.block = block
	c.mu.Unlock()
	return nil
}

// SetKeyFromString replaces the encryption key by decoding an obfuscated
// key string.
func (c *DefaultCrypto) SetKeyFromString(key string) error {
	block, err := mtcrypto.InitKeyFromString(key)
	if err != nil {
		return err
	}

	c.mu.Lock()
	c.block = block
	c.mu.Unlock()
	return nil
}
