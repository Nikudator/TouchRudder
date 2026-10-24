package peer

import (
	"errors"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    PeerID
		wantErr error
	}{
		{name: "empty", input: "", want: "", wantErr: ErrEmptyID},
		{name: "valid_libp2p", input: "12D3KooWabc", want: "12D3KooWabc", wantErr: nil},
		{name: "valid_short", input: "xyz", want: "xyz", wantErr: nil},
		{name: "whitespace_only_rejected", input: " ", want: "", wantErr: ErrEmptyID},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Parse(%q) err = %v, want %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("Parse(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestPeerID_IsZero(t *testing.T) {
	if !PeerID("").IsZero() {
		t.Error("PeerID(\"\").IsZero() = false, want true")
	}
	if PeerID("abc").IsZero() {
		t.Error("PeerID(\"abc\").IsZero() = true, want false")
	}
}

func TestPeerID_String(t *testing.T) {
	if got := PeerID("").String(); got != "" {
		t.Errorf("PeerID(\"\").String() = %q, want \"\"", got)
	}
	// roundtrip: Parse(s).String() == s
	id, err := Parse("abc")
	if err != nil {
		t.Fatalf("Parse(\"abc\") unexpected error: %v", err)
	}
	if got := id.String(); got != "abc" {
		t.Errorf("Parse(\"abc\").String() = %q, want \"abc\"", got)
	}
}
