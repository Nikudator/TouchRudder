package manifest

import (
	"encoding/hex"
	"errors"
	"fmt"
)

// PeerID — идентификатор узла в сети TouchRudder.
// Формат уточняется при интеграции libp2p (неделя 4); сейчас — непустая строка.
type PeerID string

// Validate проверяет, что PeerID не пуст.
func (p PeerID) Validate() error {
	if p == "" {
		return errors.New("peer_id: empty")
	}
	return nil
}

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
