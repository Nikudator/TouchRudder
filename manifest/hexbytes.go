package manifest

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// HexBytes — байтовый срез, сериализуемый в JSON как hex-строка.
type HexBytes []byte

var nullJSON = []byte("null")

func (h HexBytes) MarshalJSON() ([]byte, error) {

	if h == nil {
		return []byte("null"), nil
	}
	s := hex.EncodeToString(h)
	return json.Marshal(s)

}

func (h *HexBytes) UnmarshalJSON(data []byte) error {

	if bytes.Equal(data, nullJSON) {
		*h = nil
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("hexbytes: not a string: %w", err)
	}
	decoded, err := hex.DecodeString(s)
	if err != nil {
		return fmt.Errorf("hexbytes: not valid hex: %w", err)
	}
	*h = decoded
	return nil

}
