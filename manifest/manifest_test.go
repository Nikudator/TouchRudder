package manifest

import (
	"strings"
	"testing"
)

func TestPeerID_Validate(t *testing.T) {
	tests := []struct {
		name    string
		peerID  PeerID
		wantErr bool
	}{
		{
			name:    "valid peer id",
			peerID:  "some-valid-libp2p-id",
			wantErr: false,
		},
		{
			name:    "empty peer id",
			peerID:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.peerID.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("PeerID.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

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
