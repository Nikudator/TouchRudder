package manifest

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestHexBytes_MarshalJSON(t *testing.T) {
	tests := []struct {
		name  string
		input HexBytes
		want  string // ожидаемый JSON как строка
	}{
		{"nil", nil, "null"},
		{"empty", HexBytes{}, `""`},
		{"12 bytes", HexBytes{0x00, 0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb}, `"00112233445566778899aabb"`},
	}
	for _, tt := range tests {
		t.Run("HexBytes_"+tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.input)
			if err != nil {
				t.Fatalf("Marshal() unexpected error: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("Marshal() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestHexBytes_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string   // JSON
		want    HexBytes // ожидаемое значение
		wantErr bool
	}{
		{"null", "null", nil, false},
		{"empty string", `""`, HexBytes{}, false},
		{"valid hex", `"aabb"`, HexBytes{0xaa, 0xbb}, false},
		{"invalid hex", `"xyz"`, nil, true},
		{"odd length", `"abc"`, nil, true},
		{"not string", `123`, nil, true},
	}
	for _, tt := range tests {
		t.Run("HexBytes_"+tt.name, func(t *testing.T) {
			var got HexBytes
			err := json.Unmarshal([]byte(tt.input), &got)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Unmarshal() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if (got == nil) != (tt.want == nil) {
				t.Errorf("nil mismatch: got nil=%v, want nil=%v",
					got == nil, tt.want == nil)
			}
			if !bytes.Equal(got, tt.want) {
				t.Errorf("value = %x, want %x", got, tt.want)
			}
		})
	}
}

func TestHexBytes_StructIntegration(t *testing.T) {
	type wrapper struct {
		Nonce HexBytes `json:"nonce"`
	}

	// Marshal
	w := wrapper{Nonce: HexBytes{0xaa, 0xbb}}
	data, err := json.Marshal(w)
	if err != nil {
		t.Fatalf("Marshal() unexpected error: %v", err)
	}
	want := `{"nonce":"aabb"}`
	if string(data) != want {
		t.Fatalf("Marshal() = %s, want %s", data, want)
	}
	// ↑ если увидишь "qg==" в got — интерфейс не подхватился, receiver не тот

	// Unmarshal
	var gotW wrapper
	if err := json.Unmarshal(data, &gotW); err != nil {
		t.Fatalf("Unmarshal() unexpected error: %v", err)
	}
	if !bytes.Equal(gotW.Nonce, w.Nonce) {
		t.Errorf("Nonce = %x, want %x", gotW.Nonce, w.Nonce)
	}
}
