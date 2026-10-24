package manifest

import (
	"strings"
	"testing"
)

func TestChunkAndShardID_Validate(t *testing.T) {
	validHex64 := strings.Repeat("a", 64)

	tests := []struct {
		name    string
		input   string
		wantErr bool
		errText string // опционально проверим текст ошибки
	}{
		{
			name:    "valid sha256 hex",
			input:   validHex64,
			wantErr: false,
		},
		{
			name:    "too short",
			input:   "abc",
			wantErr: true,
			errText: "length 3, want 64",
		},
		{
			name:    "too long",
			input:   validHex64 + "a",
			wantErr: true,
			errText: "length 65, want 64",
		},
		{
			name:    "not hex characters",
			input:   strings.Repeat("g", 64),
			wantErr: true,
			errText: "not hex",
		},
		{
			name:    "valid hex but uppercase",
			input:   strings.Repeat("A", 64),
			wantErr: true,
			errText: "not lowercase, want lowercase",
		},
		{
			name:    "mixed case hex",
			input:   strings.Repeat("a", 63) + "A",
			wantErr: true,
			errText: "not lowercase, want lowercase",
		},
	}

	for _, tt := range tests {
		// Тестируем ChunkID
		t.Run("ChunkID_"+tt.name, func(t *testing.T) {
			c := ChunkID(tt.input)
			err := c.Validate()

			if (err != nil) != tt.wantErr {
				t.Fatalf("ChunkID.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.errText != "" && !strings.Contains(err.Error(), tt.errText) {
				t.Errorf("ChunkID.Validate() error = %v, context should contain %q", err, tt.errText)
			}
		})

		// Тестируем ShardID (поведение идентично, так как под капотом validateHex64)
		t.Run("ShardID_"+tt.name, func(t *testing.T) {
			s := ShardID(tt.input)
			err := s.Validate()

			if (err != nil) != tt.wantErr {
				t.Fatalf("ShardID.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && tt.errText != "" && !strings.Contains(err.Error(), tt.errText) {
				t.Errorf("ShardID.Validate() error = %v, context should contain %q", err, tt.errText)
			}
		})
	}
}

func TestCipherSpec_Validate(t *testing.T) {
	validKey := make(HexBytes, 32)
	validNonce := make(HexBytes, 12)

	tests := []struct {
		name    string
		spec    CipherSpec
		wantErr bool
	}{
		{
			name: "valid",
			spec: CipherSpec{
				Algorithm: "AES-256-GCM",
				KeyID:     validKey,
				Nonce:     validNonce,
			},
			wantErr: false,
		},
		{
			name: "wrong algorithm",
			spec: CipherSpec{
				Algorithm: "chacha20-poly1305",
				KeyID:     validKey,
				Nonce:     validNonce,
			},
			wantErr: true,
		},
		{
			name: "empty algorithm",
			spec: CipherSpec{
				Algorithm: "",
				KeyID:     validKey,
				Nonce:     validNonce,
			},
			wantErr: true,
		},
		{
			name: "short KeyID",
			spec: CipherSpec{
				Algorithm: "AES-256-GCM",
				KeyID:     validKey[1:],
				Nonce:     validNonce,
			},
			wantErr: true,
		},
		{
			name: "long KeyID",
			spec: CipherSpec{
				Algorithm: "AES-256-GCM",
				KeyID:     append(validKey, 0x00),
				Nonce:     validNonce,
			},
			wantErr: true,
		},
		{
			name: "nil KeyID",
			spec: CipherSpec{
				Algorithm: "AES-256-GCM",
				KeyID:     nil,
				Nonce:     validNonce,
			},
			wantErr: true,
		},
		{
			name: "short Nonce",
			spec: CipherSpec{
				Algorithm: "AES-256-GCM",
				KeyID:     validKey,
				Nonce:     validNonce[1:],
			},
			wantErr: true,
		},
		{
			name: "long Nonce",
			spec: CipherSpec{
				Algorithm: "AES-256-GCM",
				KeyID:     validKey,
				Nonce:     append(validNonce, 0x00),
			},
			wantErr: true,
		},
		{
			name: "nil Nonce",
			spec: CipherSpec{
				Algorithm: "AES-256-GCM",
				KeyID:     validKey,
				Nonce:     nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCodecSpec_Validate(t *testing.T) {

	tests := []struct {
		name    string
		spec    CodecSpec
		wantErr bool
	}{
		{
			name: "valid_empty",
			spec: CodecSpec{
				Algorithm:      "reed-solomon",
				DataShards:     4,
				ParityShards:   3,
				BlockSize:      4096,
				OriginalSize:   0,
				CiphertextSize: 16,
				EncodedSize:    paddedSize(16, 4, 4096),
			},
			wantErr: false,
		},
		{
			name: "valid_1mib",
			spec: CodecSpec{
				Algorithm:      "reed-solomon",
				DataShards:     4,
				ParityShards:   3,
				BlockSize:      4096,
				OriginalSize:   1048576,
				CiphertextSize: 1048592,
				EncodedSize:    1064960,
			},
			wantErr: false,
		},
		{
			name: "valid_1mib_plus_1",
			spec: CodecSpec{
				Algorithm:      "reed-solomon",
				DataShards:     4,
				ParityShards:   3,
				BlockSize:      4096,
				OriginalSize:   1048576,
				CiphertextSize: 1048592,
				EncodedSize:    1064960,
			},
			wantErr: false,
		},
		{
			name: "wrong algorithm",
			spec: CodecSpec{
				Algorithm:      "rs",
				DataShards:     4,
				ParityShards:   3,
				BlockSize:      4096,
				OriginalSize:   1048576,
				CiphertextSize: 1048592,
				EncodedSize:    1064960,
			},
			wantErr: true,
		},
		{
			name: "data_shards_zero",
			spec: CodecSpec{
				Algorithm:      "reed-solomon",
				DataShards:     4,
				ParityShards:   3,
				BlockSize:      4096,
				OriginalSize:   1048576,
				CiphertextSize: 1048592,
				EncodedSize:    1064960,
			},
			wantErr: true,
		},
		{
			name: "parity_shards_zero",
			spec: CodecSpec{
				Algorithm:      "reed-solomon",
				DataShards:     4,
				ParityShards:   3,
				BlockSize:      4096,
				OriginalSize:   1048576,
				CiphertextSize: 1048592,
				EncodedSize:    1064960,
			},
			wantErr: true,
		},
		{
			name: "too_many_shards",
			spec: CodecSpec{
				Algorithm:      "reed-solomon",
				DataShards:     4,
				ParityShards:   3,
				BlockSize:      4096,
				OriginalSize:   1048576,
				CiphertextSize: 1048592,
				EncodedSize:    1064960,
			},
			wantErr: true,
		},
		{
			name: "wrong_block_size",
			spec: CodecSpec{
				Algorithm:      "reed-solomon",
				DataShards:     4,
				ParityShards:   3,
				BlockSize:      4096,
				OriginalSize:   1048576,
				CiphertextSize: 1048592,
				EncodedSize:    1064960,
			},
			wantErr: true,
		},
		{
			name: "negative_original",
			spec: CodecSpec{
				Algorithm:      "reed-solomon",
				DataShards:     4,
				ParityShards:   3,
				BlockSize:      4096,
				OriginalSize:   1048576,
				CiphertextSize: 1048592,
				EncodedSize:    1064960,
			},
			wantErr: true,
		},
		{
			name: "original_too_large",
			spec: CodecSpec{
				Algorithm:      "reed-solomon",
				DataShards:     4,
				ParityShards:   3,
				BlockSize:      4096,
				OriginalSize:   1048576,
				CiphertextSize: 1048592,
				EncodedSize:    1064960,
			},
			wantErr: true,
		},
		{
			name: "ciphertext_wrong_plus",
			spec: CodecSpec{
				Algorithm:      "reed-solomon",
				DataShards:     4,
				ParityShards:   3,
				BlockSize:      4096,
				OriginalSize:   1048576,
				CiphertextSize: 1048592,
				EncodedSize:    1064960,
			},
			wantErr: true,
		},
		{
			name: "ciphertext_wrong_minus",
			spec: CodecSpec{
				Algorithm:      "reed-solomon",
				DataShards:     4,
				ParityShards:   3,
				BlockSize:      4096,
				OriginalSize:   1048576,
				CiphertextSize: 1048592,
				EncodedSize:    1064960,
			},
			wantErr: true,
		},
		{
			name: "encoded_less_than_ciphertext",
			spec: CodecSpec{
				Algorithm:      "reed-solomon",
				DataShards:     4,
				ParityShards:   3,
				BlockSize:      4096,
				OriginalSize:   1048576,
				CiphertextSize: 1048592,
				EncodedSize:    1064960,
			},
			wantErr: true,
		},
		{
			name: "encoded_not_aligned",
			spec: CodecSpec{
				Algorithm:      "reed-solomon",
				DataShards:     4,
				ParityShards:   3,
				BlockSize:      4096,
				OriginalSize:   1048576,
				CiphertextSize: 1048592,
				EncodedSize:    1064960,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.spec.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// paddedSize returns ceil(ciphertext / (D*BS)) * (D*BS).
// Independent implementation of spec §5.5 — not shared with production code,
// so a bug in Validate's divisibility check can't hide here.
func paddedSize(ciphertext int64, dataShards int, blockSize int64) int64 {
	unit := int64(dataShards) * blockSize
	if r := ciphertext % unit; r != 0 {
		return ciphertext + (unit - r)
	}
	return ciphertext
}
