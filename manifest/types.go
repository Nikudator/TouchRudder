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
