package peer

import (
	"errors"
	"strings"
)

// PeerID — идентификатор узла в сети TouchRudder.
// Формат уточняется при интеграции libp2p (неделя 4); сейчас — непустая строка.
type PeerID string

// ErrEmptyID возвращается Parse, если передана пустая строка.
// Используется с errors.Is; не заменять на inline errors.New —
// это сломает сравнение у вызывающего кода.
var ErrEmptyID = errors.New("peer: empty id")

func Parse(s string) (PeerID, error) {
	// Week 1: non-empty after trimming whitespace.
	// Goes beyond spec §4.3.0 (which says "non-empty"), tightening early
	// because an ID made only of whitespace is almost always a caller bug.
	// TODO(week4): replace with real multibase parsing.
	if strings.TrimSpace(s) == "" {
		return "", ErrEmptyID
	}
	return PeerID(s), nil
}

func (p PeerID) String() string { return string(p) }

func (p PeerID) IsZero() bool { return p == "" }
