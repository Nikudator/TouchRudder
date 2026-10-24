package manifest

import (
	"encoding/hex"
	"errors"
	"fmt"
)

const (
	BlockSize    int64 = 4096
	GCMTagSize   int64 = 16
	MaxChunkSize int64 = 8 * 1024 * 1024 // 8 МиБ
)

// ChunkID — идентификатор чанка: SHA256, 32 байта в hex (64 символа).
type ChunkID string

// Validate проверяет, что ChunkID — 64 hex-символа в нижнем регистре.
func (c ChunkID) Validate() error {
	if err := validateHex64(string(c)); err != nil {
		return fmt.Errorf("chunk_id: %w", err)
	}
	return nil
}

// ShardID — идентификатор шарда: SHA256, 32 байта в hex (64 символа).
type ShardID string

// Validate проверяет, что ShardID — 64 hex-символа в нижнем регистре.
func (s ShardID) Validate() error {
	if err := validateHex64(string(s)); err != nil {
		return fmt.Errorf("shard_id: %w", err)
	}
	return nil
}

func validateHex64(s string) error {
	if l := len(s); l != 64 {
		return fmt.Errorf("hex64: length %d, want 64", l)
	}
	decoded, err := hex.DecodeString(s)
	if err != nil {
		return fmt.Errorf("hex64: not hex: %w", err)
	}
	if hex.EncodeToString(decoded) != s {
		return errors.New("hex64: not lowercase, want lowercase")
	}
	return nil
}

// CipherSpec описывает шифрование чанка.
// Параметры FROZEN: попадают в манифест, используются при restore
// вместо значений из конфига (§13.7).
//
// Note on spec notation: §10.1 shows "<hex 32>" for key_id and "<hex 24>"
// for nonce, but the notation is ambiguous (bytes vs hex chars).
// Interpreted as bytes, consistent with §5.1:
//   - key_id = SHA256 output = 32 bytes
//   - nonce  = AES-GCM nonce  = 12 bytes (§5.1)
//
// Logged as spec ambiguity; see issue #<N>.
type CipherSpec struct {
	Algorithm string   `json:"algorithm"`
	KeyID     HexBytes `json:"key_id"`
	Nonce     HexBytes `json:"nonce"`
}

// Validate проверяет поля CipherSpec.
// Проверяется только длина: hex-валидность уже гарантирована
// HexBytes.UnmarshalJSON. Повторная проверка hex здесь — дублирование
// инварианта.
func (c CipherSpec) Validate() error {
	if c.Algorithm != "AES-256-GCM" {
		return fmt.Errorf("algorithm %q, want AES-256-GCM", c.Algorithm)
	}
	if len(c.KeyID) != 32 {
		return fmt.Errorf("key_id: length %d, want 32", len(c.KeyID))
	}
	if len(c.Nonce) != 12 {
		return fmt.Errorf("nonce: length %d, want 12", len(c.Nonce))
	}
	return nil
}

type CodecSpec struct {
	Algorithm      string `json:"algorithm"`     // "reed-solomon"
	DataShards     int    `json:"data_shards"`   // 4
	ParityShards   int    `json:"parity_shards"` // 3
	BlockSize      int64  `json:"block_size"`    // 4096
	OriginalSize   int64  `json:"original_size"`
	CiphertextSize int64  `json:"ciphertext_size"` // v0.1.2: длина ciphertext до паддинга
	EncodedSize    int64  `json:"encoded_size"`    // после паддинга
}

func (c CodecSpec) Validate() error {
	if c.Algorithm != "reed-solomon" {
		return fmt.Errorf("algorithm %q, want reed-solomon", c.Algorithm)
	}
	if c.DataShards <= 0 {
		return fmt.Errorf("data_shards must be greater than 0, got %d", c.DataShards)
	}
	if c.ParityShards <= 0 {
		return fmt.Errorf("parity_shards must be greater than 0, got %d", c.ParityShards)
	}
	if c.DataShards+c.ParityShards > 32 {
		return fmt.Errorf("data_shards + parity_shards = %d, exceeds max 32", c.DataShards+c.ParityShards)
	}
	if c.BlockSize != 4096 {
		return fmt.Errorf("block_size must be 4096, got %d", c.BlockSize)
	}
	if c.OriginalSize < 0 || c.OriginalSize > 8*1024*1024 {
		return fmt.Errorf("original_size must be greater than 0 but not must be greater than 8Mib, got %d", c.OriginalSize)
	}
	if c.CiphertextSize != c.OriginalSize+16 {
		return fmt.Errorf("ciphertext_size must be original_size + 16, got %d", c.CiphertextSize)
	}
	if c.EncodedSize < c.CiphertextSize {
		return fmt.Errorf("encoded_size must be >= ciphertext_size, got %d < %d", c.EncodedSize, c.CiphertextSize)
	}
	if c.EncodedSize%(int64(c.DataShards)*c.BlockSize) != 0 {
		return fmt.Errorf("encoded_size / (data_shards * block_size) must be zero, got %d", c.EncodedSize%(int64(c.DataShards)*c.BlockSize))
	}
	return nil
}
