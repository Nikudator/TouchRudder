# 📘 TouchRudder — Техническая спецификация v0.1.3

**Статус:** draft для пилота  
**Дата:** 2026-10-25  
**Кодовое имя:** `touchrudder` (репозиторий, бинарник — `trd`)  
**Публичное имя:** TouchRudder  
**Языковая адаптация:** Архитектура Go-native типов  

---

## 📑 Содержание

1. [О документе](#1-о-документе)
2. [Легенда названия](#2-легенда-названия)
3. [Обзор архитектуры](#3-обзор-архитектуры)
4. [Границы модулей и Go-native типы](#4-границы-модулей-и-go-native-типы)
5. [Слой хранения](#5-слой-хранения)
6. [Слой proof-of-storage](#6-слой-proof-of-storage)
7. [Слой экономики (ledger)](#7-слой-экономики-ledger)
8. [Топология обязательств](#8-топология-обязательств)
9. [Gossip](#9-gossip)
10. [Манифест и правила канонизации JSON](#10-манифест-и-правила-канонизации-json)
11. [Протокол challenge/response](#11-протокол-challengeresponse)
12. [Rebalance](#12-rebalance)
13. [Конфигурация](#13-конфигурация)
14. [Стратегия тестирования Go-типов](#14-стратегия-тестирования-go-типов)
15. [Порядок сборки](#15-порядок-сборки)
16. [План пилота](#16-план-пилота)
17. [Открытые вопросы](#17-открытые-вопросы)
18. [Журнал изменений](#18-журнал-изменений)

---

## 1. О документе

### 1.1 Назначение

Этот документ фиксирует **минимально достаточную модель** децентрализованного резервного копирования TouchRudder, пригодную для пилота на 10 узлах. Всё, что не описано явно, считается **отложенным** и перечислено в разделе 17.

### 1.2 Scope v0.1.3

**В scope:**
- Erasure coding **4+3** с AES-256-GCM
- **Паддинг чанков с `ciphertext_size`** (файл любого размера, включая пустой)
- Proof-of-storage на Merkle-дереве
- Попарный взаимный кредит с leaky bucket
- Правило замкнутых циклов ≤ 4 с якорем
- Gossip рёбер на глубину 3
- Ручной `trd rebalance` (восстановление избыточности)
- **Каноническая форма манифеста по стандарту JCS (RFC 8785) + domain-separated Ed25519**
- **Явный тип `peer.PeerID`** (без прямого импорта libp2p в чистые бизнес-пакеты)
- **Go-native типы `types.Hash32` и `types.Nonce16`** на фиксированных массивах байт для исключения аллокаций в куче и автоматического Hex-маршалинга.

**Вне scope:**
- Восстановление `identity.key`
- Merkle-доказательства для gossip (on-demand)
- Мультипротокольная совместимость
- UI/TUI
- Дедупликация между владельцами
- Автоматический rebalance
- Бэкап `ledger` при потере диска

### 1.3 Конвенции

| Обозначение | Значение |
|---|---|
| **FROZEN** | Параметр попадает в манифест; изменение ломает совместимость |
| **Tunable** | Меняется без миграции, влияет на новые операции |
| **Operational** | Пути, адреса, логи; на данные не влияет |
| `RSData`, `RSParity` | Параметры Reed-Solomon (**4, 3**) |
| `L_cap` | Потолок кредитного лимита |
| `TTL` | Глубина gossip-распространения |
| `peer.PeerID` | Строковый тип из пакета `peer/` |
| `types.Hash32` | Фиксированный тип хэша `[32]byte` |
| `types.Nonce16` | Фиксированный тип nonce `[16]byte` |

### 1.4 Что изменилось от v0.1.2

Кратко (подробно — в разделе 18):
- Устранены динамические слайсы `[]byte` и голые строки `string` из криптографических идентификаторов. Введены Go-native типы `types.Hash32` и `types.Nonce16` на фиксированных массивах байт.
- Для хэшей реализованы интерфейсы `json.Marshaler` и `json.Unmarshaler` для автоматического маршалинга в Hex-строки.
- Правила канонизации переведены на промышленный стандарт **JCS (RFC 8785)**.
- Устранена неопределенность формата наносекунд для структуры `time.Time` при хэшировании.

---

## 2. Легенда названия

### 2.1 Версия 1 — одно предложение

> **TouchRudder — распределённое резервное копирование, в котором данные удерживают курс благодаря лёгкому касанию: избыточность и локальные проверки вместо доверия к одному облаку.**

Использование: `README.md` первая строка, GitHub repo description, meta-тег.

### 2.2 Версия 2 — один абзац

> Название **TouchRudder** построено на морской метафоре. **Rudder** — руль, который удерживает судно на курсе, когда волны и течения сбивают его с пути. В нашей системе эту роль играет кодирование стирания: данные переживают потерю узлов так же, как корабль переживает волны. **Touch** — лёгкое касание руля: вмешательство минимально, проверки хранения занимают микросекунды, решения о размещении принимаются локально, без глобального консенсуса.

Использование: `docs/index.md` секция «About», первый слайд презентации, FAQ.

### 2.3 Версия 3 — одна страница

> **Почему TouchRudder.**
>
> Традиционное облачное хранение — это один руль на всех: если дата-центр недоступен, курс теряется для всех клиентов сразу. Мы строим систему, где руль есть у каждого узла, а курс удерживается за счёт избыточности.
>
> **Rudder** в названии — это механизм, который возвращает систему на курс после отказа: кодирование стирания (Reed-Solomon 4+3), при котором потеря до трёх шардов не означает потерю данных.
>
> **Touch** — это принцип минимального вмешательства. Проверка хранения (proof-of-storage) стоит микросекунды и несколько бит трафика. Решение о том, можно ли принять обязательство от нового пира, принимается локально, за три хопа gossip, без обращения к глобальному арбитру.
>
> Мы выбрали это имя, потому что оно описывает **архитектурное свойство**, а не технологию. Технологии меняются (libp2p, Reed-Solomon, Merkle), свойство остаётся: данные должны переживать отказы при минимальном управляющем воздействии.

Использование: разговор с ИБ, закупкой, инвесторами.

### 2.4 Ответ на прямой вопрос о двусмысленности

> Название обыгрывает несколько смыслов, но для нас главное — архитектурный: руль, который держит курс данных при отказах. Именно об этом документация и код; остальное — на совести тех, кто ищет второе дно.

**Чего не делать:** не отрицать, не подмигивать, не извиняться. Тон — спокойный, разговор возвращается к архитектуре.

---

## 3. Обзор архитектуры

### 3.1 Три слоя

```
┌─────────────────────────────────────────────────────┐
│  Слой хранения                                      │
│  Reed-Solomon 4+3 + AES-256-GCM                     │
│  Данные переживают потерю 3 из 7 шардов             │
└─────────────────────────────────────────────────────┘
                          ▲
┌─────────────────────────────────────────────────────┐
│  Слой проверки                                      │
│  Merkle-дерево + challenge-response                 │
│  Микросекунды на чанк, биты трафика                 │
└─────────────────────────────────────────────────────┘
                          ▲
┌─────────────────────────────────────────────────────┐
│  Слой экономики                                     │
│  Попарный взаимный кредит + leaky bucket            │
│  Правило циклов ≤ 4 с якорем                        │
└─────────────────────────────────────────────────────┘
```

### 3.2 Ключевые свойства

1. **Всё проверяется локально** — нет глобального консенсуса, нет блокчейна, нет токена.
2. **Порог входа — обычный SSD** — не требуется GPU, не требуется покупка активов.
3. **Стимулы — взаимность** — «дать место, потом взять» вместо «купить кредит».
4. **Устойчивость к Sybil** — попарный баланс требует реального хранения.
5. **Устойчивость к 3 отказам** — при схеме 4+3 на 7 шардах.

### 3.3 Ключевые параметры

| Параметр | Значение v0.1.3 |
|---|---|
| Схема кодирования | **Reed-Solomon 4+3** |
| Шардов на чанк | **7** |
| Переживает потерь | **3** |
| Накладные расходы | **+75%** (с паддингом — до +79% для мелких файлов) |
| Размер чанка | 1 МиБ (максимум) |
| Размер блока Merkle | 4 КиБ |
| Индексов в challenge | 3 |
| Потолок цикла | 4 |
| Глубина gossip | 3 |
| Период PoS | 24 ч |
| Порог нормы | 95% |
| Порог испытательного срока | 80% |

---

## 4. Границы модулей и Go-native типы

### 4.0 Базовые типы данных (`types/`)

Для обеспечения строгой типобезопасности компилятора и максимальной производительности (полное исключение аллокаций в куче при работе с миллионами хэшей) вводится пакет `types/`. Все хэши и фиксированные байтовые структуры объявляются как массивы фиксированной длины.

```go
package types

import (
	"encoding/hex"
	"fmt"
)

// Hash32 — базовый фиксированный тип для SHA-256 хэшей.
type Hash32 [32]byte

func (h Hash32) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, hex.EncodeToString(h[:]))), nil
}

func (h *Hash32) UnmarshalJSON(data []byte) error {
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("invalid hex string format")
	}
	b, err := hex.DecodeString(string(data[1 : len(data)-1]))
	if err != nil {
		return err
	}
	if len(b) != 32 {
		return fmt.Errorf("invalid hash length: expected 32 bytes, got %d", len(b))
	}
	copy(h[:], b)
	return nil
}

func (h Hash32) String() string { return hex.EncodeToString(h[:]) }

// Строгие типы-обертки для предотвращения случайного смешивания идентификаторов в коде
type ChunkID Hash32
type ShardID Hash32

// Nonce16 — фиксированный криптографический nonce для слоя проверки (16 байт = 32 hex-символа)
type Nonce16 [16]byte

func (n Nonce16) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, hex.EncodeToString(n[:]))), nil
}

func (n *Nonce16) UnmarshalJSON(data []byte) error {
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("invalid hex string format")
	}
	b, err := hex.DecodeString(string(data[1 : len(data)-1]))
	if err != nil {
		return err
	}
	if len(b) != 16 {
		return fmt.Errorf("invalid nonce length: expected 16 bytes, got %d", len(b))
	}
	copy(n[:], b)
	return nil
}
```

### 4.1 Граф зависимостей

```
                    ┌──────────────┐
                    │   cmd/cli    │
                    └──────┬───────┘
                           │
                    ┌──────▼───────┐
                    │    node/     │  ← оркестратор
                    └──┬──┬──┬──┬──┘
                       │  │  │  │
        ┌──────────────┘  │  │  └───────────────┐
        │                 │  │                  │
   ┌────▼────┐      ┌─────▼──▼─────┐      ┌─────▼─────┐
   │ storage/│      │   proof/     │      │ topology/ │
   └────┬────┘      └──────┬───────┘      └─────┬─────┘
        │                  │                    │
   ┌────▼────┐             │              ┌─────▼─────┐
   │manifest/│             │              │  gossip/  │
   └────┬────┘             │              └─────┬─────┘
        │                  │                    │
        │           ┌──────▼───────┐      ┌─────▼─────┐
        │           │   ledger/    │      │transport/ │
        │           │ (pure state) │      │ (libp2p)  │
        │           └──────────────┘      └─────┬─────┘
        │                                        │
        │                                  ┌─────▼─────┐
        │                                  │ identity/ │
        │                                  └─────┬─────┘
        │                                        │
        └────────────┐                  ┌────────▼───────┐
                     └─────────────────►│     peer/      │
                                        │  (только типы) │
                                        └────────────────┘
```

### 4.2 Правила зависимостей (проверяются линтером)

| Правило | Проверка |
|---|---|
| `peer` — чистый тип | Не импортирует ничего, кроме stdlib |
| `storage`, `manifest`, `ledger` — чистые | Не импортируют `transport`, `gossip`, `node`, `identity`, `peer` (кроме `manifest`/`ledger`, которым разрешён `peer`) |
| `proof`, `topology` — через интерфейсы | Зависят только от чистых пакетов и собственных интерфейсов |
| `transport` — единственный, кто импортирует libp2p | Конвертация `peer.PeerID ↔ libp2p.PeerID` — в одном месте |
| `node` — единственный оркестратор | Знает про все остальные пакеты |
| Циклических импортов нет | По построению |

**Уточнение:** `manifest` и `ledger` импортируют `peer/`. Пакеты `storage` и `proof` не импортируют `peer/`, так как оперируют строго типами `types.ChunkID` и `types.ShardID`.

### 4.3 Контракты пакетов

#### `peer/` — идентификатор пира

```go
package peer

type PeerID string // Хранит строковое представление Base58/Base36 (совместимое с libp2p)

// Parse проверяет синтаксис: multibase-префикс + base58/base36.
// Семантическая проверка (валидный multihash) — на границе transport/.
func Parse(s string) (PeerID, error)

func (p PeerID) String() string
func (p PeerID) IsZero() bool
```

**Инварианты:**
- `PeerID` — opaque-строка, прошедшая `Parse` **один раз** при входе в систему.
- `transport/` конвертирует `libp2p.PeerID ↔ peer.PeerID` (`transport.PeerToLibp2p`, `transport.PeerFromLibp2p`).

#### `identity/`

```go
package identity

import "touchrudder/peer"

type Identity struct { /* ed25519 + libp2p host key */ }

func Load(path string) (*Identity, error)
func Generate(path string) (*Identity, error)

func (i *Identity) PeerID() peer.PeerID
func (i *Identity) Sign(msg []byte) []byte
func Verify(pub peer.PeerID, msg, sig []byte) bool
```

**Инвариант:** `identity.key` — единственный источник истины для `PeerID`.

#### `storage/`

```go
package storage

import "touchrudder/types"

type Config struct {
    RSData    int    // 4
    RSParity  int    // 3
    ChunkSize int64  // 1 МиБ (максимум)
}

type Shard struct {
    ID       types.ShardID  
    ChunkID  types.ChunkID  
    Index    int            // 0..6
    Data     []byte         // Динамический слайс для контента
    Checksum types.Hash32   
}

type Encoder interface {
    Encode(plain []byte, cfg Config) ([]Shard, error)
    Decode(shards []Shard, cfg Config) ([]byte, error)
}

type Encryptor interface {
    Encrypt(plain, key []byte) (cipher, nonce []byte, err error)
    Decrypt(cipher, nonce, key []byte) ([]byte, error)
}

type Chunker interface {
    Split(data []byte, size int64) ([]Chunk, error)
    Join(chunks []Chunk) ([]byte, error)
}
```

**Инвариант:** `Decode` с менее чем `RSData` (=4) шардами возвращает `ErrInsufficientShards`. Паддинг чанков выполняется на уровне `node`.

#### `manifest/`

```go
package manifest

import (
	"context"
	"time"
	"touchrudder/peer"
	"touchrudder/types"
)

type Manifest struct {
    Version    int              `json:"version"`
    Revision   int              `json:"revision"`
    ChunkID    types.ChunkID    `json:"chunk_id"`
    Owner      peer.PeerID      `json:"owner"`
    CreatedAt  time.Time        `json:"created_at"`
    UpdatedAt  time.Time        `json:"updated_at"`
    Cipher     CipherSpec       `json:"cipher"`
    Codec      CodecSpec        `json:"codec"`
    Shards     []ShardLocation  `json:"shards"`     // Ровно 7 элементов
    Signature  []byte           `json:"signature,omitempty"`
}

type ShardLocation struct {
    Index      int           `json:"index"`
    ShardID    types.ShardID `json:"shard_id"`
    PeerID     peer.PeerID   `json:"peer_id"`
    Size       int64         `json:"size"`
    MerkleRoot types.Hash32  `json:"merkle_root"`
    PlacedAt   time.Time     `json:"placed_at"`
}

type CipherSpec struct {
    Algorithm string        `json:"algorithm"` // "AES-256-GCM"
    KeyID     types.Hash32  `json:"key_id"`
    Nonce     []byte        `json:"nonce"`
}

type CodecSpec struct {
    Algorithm      string  `json:"algorithm"` // "reed-solomon"
    DataShards     int     `json:"data_shards"`
    ParityShards   int     `json:"parity_shards"`
    BlockSize      int64   `json:"block_size"`
    OriginalSize   int64   `json:"original_size"`
    CiphertextSize int64   `json:"ciphertext_size"`
    EncodedSize    int64   `json:"encoded_size"`
}

type Store interface {
    Put(m Manifest) error
    Get(id types.ChunkID) (Manifest, error)
    List() ([]types.ChunkID, error)
    Delete(id types.ChunkID) error
}
```

#### `proof/`

```go
package proof

import (
	"context"
	"touchrudder/storage"
	"touchrudder/types"
)

type Challenge struct {
    ShardID types.ShardID `json:"shard_id"`
    Indices []int         `json:"indices"`
    Nonce   types.Nonce16 `json:"nonce"`
}

type Response struct {
    ShardID types.ShardID `json:"shard_id"`
    Nonce   types.Nonce16 `json:"nonce"`
    Proofs  []MerkleProof `json:"proofs"`
}

type MerkleProof struct {
	Index   int            `json:"index"`
	Block   []byte         `json:"block"`
	Path    []ProofStep    `json:"path"`
}

type ProofStep struct {
	Sibling  types.Hash32 `json:"sibling"`
	Position string       `json:"position"` // "left" или "right"
}

type Challenger interface {
    Generate(shard storage.Shard, k int) (Challenge, error)
    Respond(shard storage.Shard, c Challenge) (Response, error)
    Verify(root types.Hash32, c Challenge, r Response) bool
}

type ShardReader interface {
    ReadShard(ctx context.Context, id types.ShardID) (storage.Shard, error)
}

type Result int
const (
    ResultOK Result = iota
    ResultSoftFailure
    ResultHardFailure
)
```

#### `ledger/`

```go
package ledger

import (
	"time"
	"touchrudder/peer"
)

type Size int64

type Balance struct {
    Credit Size
    Debit  Size
}

type State int
const (
    StateNormal State = iota
    StateProbation
    StateAnnulled
)

type PeerRecord struct {
    Balance    Balance
    Limit      Size
    State      State
    Window     [100]bool
    FailStreak int
}

type Ledger struct { /* map[peer.PeerID]*PeerRecord + config */ }

func (l *Ledger) Credit(p peer.PeerID, amount Size)
func (l *Ledger) Debit(p peer.PeerID, amount Size)
func (l *Ledger) RecordSuccess(p peer.PeerID, now time.Time)
func (l *Ledger) RecordFailure(p peer.PeerID, hard bool, now time.Time) Change
func (l *Ledger) Limit(p peer.PeerID) Size
func (l *Ledger) State(p peer.PeerID) State
func (l *Ledger) Snapshot() []byte
func Restore(data []byte) (*Ledger, error)

type Change struct {
    Peer         peer.PeerID
    OldState     State
    NewState     State
    OldLimit     Size
    NewLimit     Size
    EdgeAnnulled bool
}
```

#### `gossip/`

```go
package gossip

import (
	"context"
	"touchrudder/peer"
	"touchrudder/types"
)

type Edge struct {
    From  peer.PeerID   `json:"from"`
    To    peer.PeerID   `json:"to"`
    Shard types.ShardID `json:"shard"`
    Sig   []byte        `json:"sig"`
}

type Gossip interface {
    Broadcast(ctx context.Context, e Edge) error
    Outgoing(p peer.PeerID) []Edge
    KnownPath(from, to peer.PeerID, maxLen int) ([]Edge, bool)
    Subscribe(handler func(Edge))
}
```

#### `topology/`

```go
package topology

import (
	"context"
	"touchrudder/ledger"
	"touchrudder/peer"
)

type Verdict int
const (
    VerdictAccept Verdict = iota
    VerdictRejectCycle
    VerdictRejectUnanchored
    VerdictRejectAnnulled
)

type LedgerView interface {
	State(p peer.PeerID) ledger.State
}

type Checker struct {
    gossip gossip.Gossip
    ledger LedgerView
}

func (c *Checker) WouldAccept(ctx context.Context, p peer.PeerID) (Verdict, error)
func (c *Checker) AnnulEdge(ctx context.Context, p peer.PeerID) error
func (c *Checker) IsAnchored(ctx context.Context, cycle []peer.PeerID) (bool, error)
```

#### `transport/`

```go
package transport

import (
	"context"
	"touchrudder/peer"
)

type Message struct {
	Type    string `json:"type"`
	Payload []byte `json:"payload"`
}

type Transport interface {
    Send(ctx context.Context, to peer.PeerID, msg Message) error
    Receive(handler func(from peer.PeerID, msg Message))
    Publish(ctx context.Context, topic string, data []byte) error
    Subscribe(topic string, handler func(from peer.PeerID, data []byte))
}
```

#### `node/`

```go
package node

import (
	"context"
	"time"
	"touchrudder/types"
)

type Node struct { /* композиция всех слоев */ }

func (n *Node) Backup(ctx context.Context, path string) (types.ChunkID, error)
func (n *Node) Restore(ctx context.Context, id types.ChunkID, dest string) error
func (n *Node) Rebalance(ctx context.Context) (RebalanceReport, error)

type RebalanceReport struct {
    ChunksScanned    int
    ChunksHealthy    int
    ChunksRebalanced int
    ChunksLost       int
    Errors           []error
    Duration         time.Duration
}
```

---

## 5. Слой хранения

### 5.1 Параметры

| Параметр | Значение | Обоснование |
|---|---|---|
| Схема | **Reed-Solomon 4+3** | Переживает потерю **3 из 7** шардов |
| Накладные расходы | **+75%** | 10 ГБ данных → 17.5 ГБ на дисках |
| Размер чанка | **1 МиБ (максимум)** | Файл режется на чанки ≤ `chunk_size` |
| Шифрование | AES-256-GCM | Стандарт, аппаратное ускорение |
| Ключ | На файл, не на чанк | Упрощает управление ключами |
| Nonce | 12 байт, уникален на чанк | Безопасность блочного шифра |

### 5.2 Почему 4+3, а не 4+2

При схеме 4+2 отключение 3 узлов из 10 теряет данные с вероятностью ≈ 16.7% на чанк при случайном сбое или гарантированно при направленной атаке. Критерий «отключение 3 нод» невыполним. Схема 4+3 оставляет минимум 4 шарда из 7, данные полностью восстановимы. Накладные расходы составляют +75% диска.

### 5.3 Размещение

- Не более **1 шарда одного чанка на узел**.
- Все **7 шардов** — строго на разных пирах.
- Кандидаты выбираются через `topology.WouldAccept`.
- Если доступных кандидатов меньше 7 → возвращается `ErrNoPlacement`.

### 5.4 Инкрементальные бэкапы

Используется чанкование по границам файлов. Версионирование выполняется целыми чанками. Изменение одного байта приводит к пересчету всего чанка, но не затрагивает соседние чанки.

### 5.5 Edge-кейсы чанкования и процедура паддинга

Произвольный файл не кратен базовому блоку. Паддинг применяется к ciphertext после шифрования. К ciphertext (`plaintext + 16 байт GCM-тега`) дописываются нулевые байты до размера, кратного `data_shards × block_size`.

**Три размера в манифесте:**
- `codec.original_size`: чистый plaintext.
- `codec.ciphertext_size`: `original_size + 16` байт (длина до паддинга).
- `codec.encoded_size`: длина после паддинга (передаваемая в Reed-Solomon).

**Процедура упаковки чанка:**
```
1. plaintext  = байты файла (размер ≤ chunk_size)
2. ciphertext = AES-GCM-Encrypt(plaintext)  // длина = original_size + 16
3. padded     = ciphertext || нули до ceil(len / (4 × 4096)) × (4 × 4096)
4. shards     = RS-Encode(padded)            // 7 шардов одинакового размера
5. Сохранение размеров в manifest.Codec.*
```

**Процедура распаковки чанка:**
```
1. padded     = RS-Decode(≥ 4 доступных шардов) // получаем encoded_size байт
2. ciphertext = padded[:ciphertext_size]
3. plaintext  = AES-GCM-Decrypt(ciphertext)
4. Проверка: len(plaintext) == original_size
```

**Таблица примеров чанкования (`data_shards=4`, `block_size=4096`, `chunk_size=1 МиБ`):**

| Исходный файл | Plaintext | Ciphertext | Padded | Размер 1 шарда |
|---|---|---|---|---|
| Пустой файл | 0 Б | 16 Б | 16 384 Б | 4 096 Б (1 блок) |
| Файл 500 КиБ | 512 000 Б | 512 016 Б | 524 288 Б | 131 072 Б (32 блока) |
| Ровно 1 МиБ | 1 048 576 Б | 1 048 592 Б | 1 064 960 Б | 266 240 Б (65 блоков) |

---

## 6. Слой proof-of-storage

### 6.1 Схема: Merkle-дерево над шардом

Шард разбивается на фиксированные блоки. Пир доказывает владение блоками по случайно сгенерированным индексам, предоставляя пути верификации (Merkle Proofs).

### 6.2 Структура

- Блок шарда: **4 КиБ** (4096 байт).
- Лист дерева: `leaf_i = SHA256(0x00 ∥ block_i)` — domain separation.
- Внутренний узел: `parent = SHA256(0x01 ∥ left ∥ right)`.
- При нечетном числе листьев последний дублируется.
- Итоговый корень дерева `merkle_root` записывается в манифест чанка.

### 6.3 Классификация результатов проверки

| Ситуация | Результат проверки |
|---|---|
| Все предоставленные proofs валидны | `ResultOK` |
| Истек таймаут ожидания ответа | `ResultSoftFailure` |
| Предоставленный `nonce` не совпадает с запросом | `ResultHardFailure` |
| Вычисленный Merkle Path не сходится к корню | `ResultHardFailure` |
| Невалидная Ed25519 подпись ответа | `ResultHardFailure` |

---

## 7. Слой экономики (ledger)

### 7.1 Модель: попарный взаимный кредит

Каждый узел ведет локальное состояние по отношению к известным пирам. Глобальный токен или общий реестр (блокчейн) отсутствуют.
- **Credit**: объем данных, которые пиры хранят для меня.
- **Debit**: объем данных, которые я храню для пиров.

### 7.2 Рост кредитного лимита: leaky bucket

Динамический лимит доверия рассчитывается по формуле:
\[	ext{Limit}(t+1) = \min(L_{	ext{cap}}, \gamma \cdot 	ext{Limit}(t) + eta \cdot 	ext{success})\]

*Параметры пилота:*
- `L_cap`: 10 ГиБ (абсолютный потолок).
- `β`: 100 МиБ (прибавка к лимиту за каждую успешную проверку PoS).
- `γ`: 0.99 (коэффициент забывания/затухания).

### 7.3 Трёхуровневая эскалация нарушений

1. **Норма (Уровень 0)**: Доля успешных ответов пира ≥ 95%.
2. **Испытательный срок (Уровень 1)**: Доля успехов в диапазоне 80% ≤ success < 95% ИЛИ зафиксирован 1 `HardFailure`. Текущий лимит замораживается. Новые обязательства от пира не принимаются, старые продолжают удерживаться. Выход на уровень 0 происходит после 50 проверок при восстановлении метрики ≥ 95%.
3. **Аннулирование (Уровень 2)**: Доля успехов < 80% ИЛИ получено 2 `HardFailure` подряд. Ребро обязательств считается разорванным. Лимит пира сокращается на 50%.

### 7.4 Старт нового узла

Балансы инициализируются нулями. Для получения места в сети новый узел обязан сначала предоставить дисковое пространство под чужие шарды.

### 7.5 Связь с `rebalance`

Переход пира в состояние `StateAnnulled` служит триггером для выполнения ручной команды `trd rebalance` с целью восстановления избыточности.

### 7.6 Потеря диска и локального ledger

При полной утере диска узел теряет локальное хранилище балансов и закрытый ключ `identity.key` (формируется новый `PeerID`). Другие пиры фиксируют отсутствие ответов от старого `PeerID`, генерируют серию `HardFailure` и через стандартную процедуру эскалации аннулируют старые обязательства. Экономический баланс восстанавливаемого узла начинается с нуля.

---

## 8. Топология обязательств

### 8.1 Правило топологии

Замкнутые циклы взаимных обязательств пиров допускаются длиной **не более 4 хопов**. Каждое ребро обязательства должно подтверждаться актуальным proof-of-storage. Замкнутый цикл без наличия внешнего "якоря" не увеличивает доступный кредитный лимит узла.

### 8.2 Локальная проверка циклов

Перед приемом обязательства от узла `X`, локальный менеджер топологии строит граф и ищет направленный путь `me → X` на глубину ≤ 3 хопов по данным gossip. Если путь обнаружен — транзакция закроет допустимый цикл (≤ 4). Если путь не найден — потенциальный цикл будет длиннее 4, транзакция отклоняется.

### 8.3 Определение якоря

Цикл считается заземленным (имеет якорь), если хотя бы один из участников цикла имеет открытый положительный баланс с внешним узлом, не входящим в данный цикл.

---

## 9. Gossip

### 9.1 Формат обмена данными

Узлы распространяют по сети плоские списки пар `peer.PeerID` (рёбра графа обязательств). Информация о конкретных объемах балансов в gossip не передается ради минимизации накладных расходов.

### 9.2 Параметры gossip-протокола

- **TTL (Time to Live)**: 3 (фиксировано).
- **Интервал рассылки**: 1 час.
- **Максимальный размер сообщения**: 64 КиБ (до 256 рёбер).
- **Rate limit**: не более 10 gossip-пакетов в секунду от одного пира.

### 9.3 Связь с топологией

Параметр `gossip.ttl` жестко связан с правилом топологии: `gossip.ttl >= topology.max_cycle_length - 1`. Изменение параметров валидируется при старте ноды.

---

## 10. Манифест и правила канонизации JSON

### 10.1 Формат структуры манифеста (v1)

Манифест описывает метаданные отдельного чанка. Благодаря использованию Go-native типов `types.Hash32` и `types.ChunkID`, поля автоматически сериализуются встроенным маршалером в Hex-строки.

```json
{
  "version": 1,
  "revision": 0,
  "chunk_id": "4a8a65529f798ecc0b4f8d757d544096d2b519098716bba84b901b0f55cf55cf",
  "owner": "12D3KooWLvUp4g...",
  "created_at": "2026-10-25T12:00:00Z",
  "updated_at": "2026-10-25T12:00:00Z",
  "cipher": {
    "algorithm": "AES-256-GCM",
    "key_id": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
    "nonce": "aabbccddeeff001122334455"
  },
  "codec": {
    "algorithm": "reed-solomon",
    "data_shards": 4,
    "parity_shards": 3,
    "block_size": 4096,
    "original_size": 1048576,
    "ciphertext_size": 1048592,
    "encoded_size": 1064960
  },
  "shards": [
    {
      "index": 0,
      "shard_id": "8fa1b490f230129c11823901bcf84321908741230abf891234bcde8912345678",
      "peer_id": "12D3KooWJz...",
      "size": 266240,
      "merkle_root": "1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b",
      "placed_at": "2026-10-25T12:00:00Z"
    }
  ],
  "signature": "3045022100a7b..."
}
```

### 10.2 Правила валидации манифеста

1. `version == 1`, `revision >= 0`.
2. Публичный ключ `owner` соответствует подписи `signature`.
3. Количество элементов в массиве `shards` строго равно `data_shards + parity_shards` (=7).
4. Индексы шардов покрывают диапазон `{0..6}` без пропусков и дубликатов.
5. Идентификатор шарда детерминирован: `shard_id = SHA256(chunk_id ∥ index_byte)`.
6. `ciphertext_size == original_size + 16`.
7. `encoded_size % (data_shards * block_size) == 0`.

### 10.6 Устранение недетерминизма хэширования в Go

**Проблема:** Встроенный JSON-маршалер Go (`encoding/json`) обходит внутренние ассоциативные массивы (`map`) в случайном порядке. Тип `time.Time` может сериализоваться с разной точностью наносекунд в зависимости от ОС. Это делает невозможной прямую проверку подписи Ed25519.

**Реализация канонизации (`manifest/canonical.go`):**
1. Для приведения к каноническому виду используется стандарт **JCS (RFC 8785)** посредством библиотеки `github.com/gowebpki/jcs`. Ключи сортируются лексикографически.
2. Все временные метки (`time.Time`) перед преобразованием переводятся в зону UTC и усекаются до секунд: `.UTC().Truncate(time.Second)`.
3. Поле `signature` принудительно исключается из сериализуемой структуры (`Signature = nil`).

**Domain separation подписи:**
```
sig = Ed25519.Sign(privkey, "touchrudder/manifest/v1
" || JCS_Transform(json_bytes))
```

---

## 11. Протокол challenge/response

### 11.1 Структура Challenge

Отправляется владельцем данных хранителю.
```json
{
  "type": "pos.challenge",
  "shard_id": "8fa1b490f230129c11823901bcf84321908741230abf891234bcde8912345678",
  "indices": [42, 1337, 2047],
  "nonce": "00112233445566778899aabbccddeeff",
  "issued_at": "2026-10-25T15:00:00Z",
  "signature": "..."
}
```
Строка `nonce` парсится непосредственно в Go-тип `types.Nonce16` (16 байт).

### 11.2 Структура Response

Ответ хранителя. Включает в себя оригинальный `nonce` и массив доказательств Merkle.
```json
{
  "type": "pos.response",
  "shard_id": "8fa1b490f230129c11823901bcf84321908741230abf891234bcde8912345678",
  "nonce": "00112233445566778899aabbccddeeff",
  "proofs": [
    {
      "index": 42,
      "block": "a1b2c3...",
      "path": [
        { "sibling": "1a2b...", "position": "left" }
      ]
    }
  ],
  "responded_at": "2026-10-25T15:00:02Z",
  "signature": "..."
}
```

### 11.3 Алгоритм верификации
```
leaf = SHA256(0x00 ∥ proof.block)
current_node = leaf
for step in proof.path:
    if step.position == "left":
        current_node = SHA256(0x01 ∥ step.sibling ∥ current_node)
    else:
        current_node = SHA256(0x01 ∥ current_node ∥ step.sibling)
assert current_node == manifest.shards[i].merkle_root
```

---

## 12. Rebalance

### 12.1 Назначение

Процедура ручного восстановления избыточности чанков (`trd rebalance`) в случае фиксации стабильного отказа (`StateAnnulled`) одного или нескольких пиров, удерживавших шарды.

### 12.2 Алгоритм восстановления
```
1. Сканирование доступных локально манифестов.
2. Для каждого чанка:
   a. Подсчет живых шардов (у которых статус пира == StateNormal).
   b. Если доступно < 4 шардов (RSData) -> Выход с ошибкой ErrDataLost.
   c. Если доступно от 4 до 6 шардов -> Запуск процедуры восстановления:
      i.   Загрузка уцелевших шардов.
      ii.  Декодирование Reed-Solomon -> Сборка padded-блока.
      iii. Усечение padded-блока до размера codec.ciphertext_size.
      iv.  Расшифровка AES-GCM (контроль целостности данных).
      v.   Повторное шифрование и генерация полного набора из 7 новых шардов.
      vi.  Поиск новых пиров через topology.WouldAccept для размещения недостающих шардов.
      vii. Запись нового манифеста с инкрементом поля revision.
```

---

## 13. Конфигурация

### 13.1 Формат конфигурационного файла: TOML

Файл содержит глобальные параметры ноды. Секреты (пароли ключей) в файл конфигурации не записываются и передаются исключительно через переменные окружения (`TR_IDENTITY_PASSWORD`, `TR_BACKUP_KEY`).

### 13.2 Структура `config.toml`

```toml
version = 1

[identity]
key_path = "/var/lib/touchrudder/identity.key"
key_file_mode = "0600"

[transport]
listen = ["/ip4/0.0.0.0/tcp/4001"]
dial_timeout = "10s"
request_timeout = "30s"
max_connections = 128

[storage]
data_dir = "/var/lib/touchrudder/shards"
manifest_dir = "/var/lib/touchrudder/manifests"
data_shards = 4       # FROZEN
parity_shards = 3     # FROZEN
block_size = 4096     # FROZEN
chunk_size = 1048576  # FROZEN (1 МиБ максимум)

[proof]
k = 3
timeout = "30s"
nonce_length = 16
check_interval = "24h"

[ledger]
limit_cap = "10GiB"
beta = "104857600"   # 100 MiB прибавка
gamma = 0.99
threshold_normal = 0.95
threshold_probation = 0.80

[topology]
max_cycle_length = 4
require_anchor = true
```

---

## 14. Стратегия тестирования Go-типов

Вся бизнес-логика пакетов `ledger`, `topology` и `storage` изолирована от сетевого рантайма транспорта через интерфейсы и покрывается unit/fuzz тестами.

### 14.1 Обязательные тестовые сценарии (Go Unit Tests)
- `Test_Hash32_JSON_Roundtrip`: проверка, что тип `types.Hash32` корректно обрабатывает корректные/некорректные Hex-строки при сериализации в JSON.
- `Test_Canonical_JCS_Compliance`: верификация подписи манифеста при случайном изменении порядка полей структур в рантайме Go (сверка с эталонным вектором из `testdata/canonical_vectors.json`).
- `Test_Padding_Edge_Cases`: Property-based тестирование кодека на массивах байт длиной: 0 байт, 500 КиБ, ровно 1 МиБ, 1 МиБ + 1 байт.

---

## 15. Порядок сборки

| Спринт | Модули | Выходной артефакт |
|---|---|---|
| Неделя 1 | `types/`, `peer/`, `storage/` + Паддинг | Валидный кодек 4+3, Hex-трансформеры |
| Неделя 2 | `manifest/` + `manifest/canonical.go` (JCS) | Детерминированный сериализатор манифестов |
| Неделя 3 | `ledger/` (State Machine) | Компонент расчёта лимитов без сетевых побочных эффектов |
| Неделя 4 | `proof/` (Merkle дерево и Challenger) | Генератор и валидатор Merkle Proofs |
| Неделя 5 | `transport/`, `gossip/`, `topology/` | Сетевой хост libp2p, валидатор циклов графа |
| Неделя 6 | `node/`, `cmd/cli/` | CLI-приложение `trd`, интеграционные тесты |

---

## 16. План пилота

Пилот разворачивается в закрытой сети на **10 реальных узлах** со временем проведения 6 недель.

### 16.2 Ключевые интеграционные сценарии пилота

- **Сценарий A8 (Тест пустого файла)**: Загрузка файла размером 0 байт. Ожидается создание манифеста с параметрами `ciphertext_size = 16`, `encoded_size = 16384` и успешное прохождение фазы `Restore`.
- **Сценарий B4 (Критерий устойчивости к 3 отказам)**: Принудительное отключение 3 узлов сети одновременно. Выполнение команды восстановления данных. Ожидается 100% успех скачивания и распаковки чанков из оставшихся 4 шардов.
- **Сценарий D7 (Тест иммунитета подписи)**: Искусственное изменение порядка следования полей в JSON-файле манифеста на стороне хранилища. Ожидается, что принимающий узел успешно валидирует подпись благодаря принудительной JCS-канонизации.

---

## 17. Открытые вопросы

| Задача | Срок реализации | Место в архитектуре |
|---|---|---|
| Механизм восстановления приватного ключа узла | v0.2 | `identity/recovery.go` |
| Автоматический фоновый запуск ребалансировки | v0.2 | `node/syncer.go` |
| Инструменты экспорта/импорта состояния ledger | v0.2 | `ledger/backup.go` |

---

## 18. Журнал изменений

### v0.1.3 (2026-10-25)
- **Исправление архитектуры типов:** Введены Go-native типы `types.Hash32` и `types.Nonce16` на базе фиксированных массивов байт вместо динамических слайсов `[]byte`, устраняющие лишние аллокаций в куче Go.
- **Интеграция с JSON:** Для хэш-типов добавлены встроенные реализации `MarshalJSON`/`UnmarshalJSON` для прозрачного перевода структур данных в Hex-строки.
- **Канонизация:** Старая кастомная процедура сортировки заменена на промышленный стандарт **JCS (RFC 8785)**.
- **Стабилизация времени:** Специфицировано принудительное округление и сброс наносекунд для `time.Time` перед расчётом хэша подписи манифеста.

---

## 💎 Итог

Спецификация версии **v0.1.3** полностью адаптирована к реализации на языке Go, ликвидирует рантайм-неопределенности JSON-маршалинга и готова к передаче в разработку.
