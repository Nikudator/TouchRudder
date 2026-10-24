# 📘 TouchRudder — Техническая спецификация v0.1.2

**Статус:** draft для пилота
**Дата:** 2026-10-24
**Кодовое имя:** `touchrudder` (репозиторий, бинарник — `trd`)
**Публичное имя:** TouchRudder

> **Как использовать этот документ.** Сохраните его как `touchrudder-spec-v0.1.2.md`. Он самодостаточен: содержит архитектуру, модули, форматы, конфиг, план пилота. Всё, что не описано явно, — отложено и перечислено в разделе 17.

---

## 📑 Содержание

1. [О документе](#1-о-документе)
2. [Легенда названия](#2-легенда-названия)
3. [Обзор архитектуры](#3-обзор-архитектуры)
4. [Границы модулей](#4-границы-модулей)
5. [Слой хранения](#5-слой-хранения)
6. [Слой proof-of-storage](#6-слой-proof-of-storage)
7. [Слой экономики (ledger)](#7-слой-экономики-ledger)
8. [Топология обязательств](#8-топология-обязательств)
9. [Gossip](#9-gossip)
10. [Манифест](#10-манифест)
11. [Протокол challenge/response](#11-протокол-challengeresponse)
12. [Rebalance](#12-rebalance)
13. [Конфигурация](#13-конфигурация)
14. [Стратегия тестирования](#14-стратегия-тестирования)
15. [Порядок сборки](#15-порядок-сборки)
16. [План пилота](#16-план-пилота)
17. [Открытые вопросы](#17-открытые-вопросы)
18. [Журнал изменений](#18-журнал-изменений)

---

## 1. О документе

### 1.1 Назначение

Этот документ фиксирует **минимально достаточную модель** децентрализованного резервного копирования TouchRudder, пригодную для пилота на 10 узлах. Всё, что не описано явно, считается **отложенным** и перечислено в разделе 17.

### 1.2 Scope v0.1.2

**В scope:**
- Erasure coding **4+3** с AES-256-GCM
- **Паддинг чанков с `ciphertext_size`** (файл любого размера, включая пустой)
- Proof-of-storage на Merkle-дереве
- Попарный взаимный кредит с leaky bucket
- Правило замкнутых циклов ≤ 4 с якорем
- Gossip рёбер на глубину 3
- Ручной `trd rebalance` (восстановление избыточности)
- **Каноническая форма манифеста + domain-separated Ed25519**
- **Явный тип `peer.PeerID`** (без импорта libp2p в чистые пакеты)

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
| `peer.PeerID` | Строковый тип из пакета `peer/` (см. 4.3.0) |

### 1.4 Что изменилось от v0.1

Кратко (подробно — в разделе 18):

- `parity_shards`: **2 → 3**. Схема **4+2 → 4+3**. Критерий пилота «отключение 3 узлов» невыполним при 4+2.
- Добавлен **`trd rebalance`** — ручное восстановление избыточности.
- Добавлен раздел «План пилота» (16).
- Явно задокументировано ограничение: **`ledger` не переживает потерю диска** (7.6).

*(Дельта v0.1.1 → v0.1.2 — в разделе 18.)*

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

| Параметр | Значение v0.1.2 |
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

## 4. Границы модулей

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

**Уточнение:** `manifest` и `ledger` импортируют `peer/` (это чистый пакет без внешних зависимостей). `storage`, `proof` не импортируют `peer/`, потому что работают только с `ChunkID`/`ShardID`.

### 4.3 Контракты пакетов

#### `peer/` (4.3.0) — идентификатор пира

**Назначение.** Явный тип для `PeerID`, не зависящий от libp2p. Пакет `manifest/` (правило 4.2) не может импортировать `transport/`, а значит и `libp2p.PeerID`. Голый `string` — потеря типобезопасности: компилятор не отличит `PeerID` от `ChunkID`/`ShardID`.

```go
package peer

type PeerID string

// Parse проверяет синтаксис: multibase-префикс + base58/base36.
// Семантическая проверка (валидный multihash) — на границе transport/.
func Parse(s string) (PeerID, error)

func (p PeerID) String() string
func (p PeerID) IsZero() bool
```

**Инварианты:**
- `PeerID` — opaque-строка, прошедшая `Parse` **один раз** при входе в систему.
- Внутри системы `PeerID` не валидируется повторно.
- `transport/` конвертирует `libp2p.PeerID ↔ peer.PeerID` (`transport.PeerToLibp2p`, `transport.PeerFromLibp2p`).
- `peer/` не импортирует ничего, кроме `errors`, `fmt`, `strings`.

#### `identity/`

```go
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
type Config struct {
    RSData    int    // 4
    RSParity  int    // 3
    ChunkSize int64  // 1 МиБ (максимум)
}

type Shard struct {
    ID       ShardID
    ChunkID  ChunkID
    Index    int    // 0..6
    Data     []byte
    Checksum [32]byte
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

**Инвариант:** `Decode` с менее чем `RSData` (=4) шардами возвращает `ErrInsufficientShards`.

**Паддинг:** `Encode` принимает данные, уже дополненные до `data_shards × block_size` (см. 5.5). `Decode` возвращает padded-данные; усечение до `ciphertext_size` — на стороне `node`.

#### `manifest/`

```go
import "touchrudder/peer"

type Manifest struct {
    Version    int
    Revision   int
    ChunkID    ChunkID
    Owner      peer.PeerID
    CreatedAt  time.Time
    UpdatedAt  time.Time
    Cipher     CipherSpec
    Codec      CodecSpec
    Shards     []ShardLocation  // 7 элементов
    Signature  []byte
}

type ShardLocation struct {
    Index      int
    ShardID    ShardID
    PeerID     peer.PeerID
    Size       int64
    MerkleRoot [32]byte
    PlacedAt   time.Time
}

type CodecSpec struct {
    Algorithm      string  // "reed-solomon"
    DataShards     int     // 4
    ParityShards   int     // 3
    BlockSize      int64   // 4096
    OriginalSize   int64
    CiphertextSize int64   // v0.1.2: длина ciphertext до паддинга
    EncodedSize    int64   // после паддинга
}

type Store interface {
    Put(m Manifest) error
    Get(id ChunkID) (Manifest, error)
    List() ([]ChunkID, error)
    Delete(id ChunkID) error
}
```

**Инвариант:** манифест **не содержит** ключа шифрования в открытом виде — только `KeyID`.

**Канонизация:** подпись — над канонической формой без поля `signature` (см. 10.6).

#### `proof/`

```go
type Challenge struct {
    ShardID ShardID
    Indices []int
    Nonce   [16]byte
}

type Response struct {
    ShardID ShardID      // v0.1.2: было PeerID — исправлено (P0.1)
    Nonce   [16]byte
    Proofs  []MerkleProof
}

type Challenger interface {
    Generate(shard Shard, k int) (Challenge, error)
    Respond(shard Shard, c Challenge) (Response, error)
    Verify(root [32]byte, c Challenge, r Response) bool
}

type ShardReader interface {
    ReadShard(ctx context.Context, id ShardID) (Shard, error)
}

type Result int
const (
    ResultOK          Result = iota
    ResultSoftFailure
    ResultHardFailure
)
```

**Инвариант:** `Verify` детерминирован.

#### `ledger/`

```go
import "touchrudder/peer"

type Size int64

type Balance struct {
    Credit Size // мне должны
    Debit  Size // я должен
}

type State int
const (
    StateNormal    State = iota
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

**Инварианты:**
- `Limit(peer) ≤ L_cap` всегда
- Переходы состояний — только через `RecordSuccess`/`RecordFailure`
- `Snapshot`/`Restore` идемпотентны

#### `gossip/`

```go
import "touchrudder/peer"

type Edge struct {
    From  peer.PeerID
    To    peer.PeerID
    Shard ShardID
    Sig   []byte
}

type Gossip interface {
    Broadcast(ctx context.Context, e Edge) error
    Outgoing(p peer.PeerID) []Edge
    KnownPath(from, to peer.PeerID, maxLen int) ([]Edge, bool)
    Subscribe(handler func(Edge))
}
```

**Реализация v0.1.2:** полный список `PeerID` без балансов, broadcast через libp2p PubSub, TTL=3.

**Инвариант:** `KnownPath` возвращает только подтверждённые пути.

#### `topology/`

```go
import "touchrudder/peer"

type Verdict int
const (
    VerdictAccept         Verdict = iota
    VerdictRejectCycle
    VerdictRejectUnanchored
    VerdictRejectAnnulled
)

type Checker struct {
    gossip Gossip
    ledger LedgerView
}

func (c *Checker) WouldAccept(ctx context.Context, p peer.PeerID) (Verdict, error)
func (c *Checker) AnnulEdge(ctx context.Context, p peer.PeerID) error
func (c *Checker) IsAnchored(ctx context.Context, cycle []peer.PeerID) (bool, error)
```

**Алгоритм `WouldAccept`:**
1. `State(peer) == Annulled` → `RejectAnnulled`
2. `KnownPath(me, peer, 3)` → цикл ≤ 4. Проверить `IsAnchored`.
3. Нет пути → цикл ≥ 5 → `RejectCycle`

**Инвариант:** консервативен — при неполных данных возвращает `Reject`.

#### `transport/`

```go
import "touchrudder/peer"

type Transport interface {
    Send(ctx context.Context, to peer.PeerID, msg Message) error
    Receive(handler func(from peer.PeerID, msg Message))
    Publish(ctx context.Context, topic string, data []byte) error
    Subscribe(topic string, handler func(from peer.PeerID, data []byte))
}

// Единственное место в коде, где peer.PeerID конвертируется в libp2p.PeerID и обратно.
func PeerToLibp2p(p peer.PeerID) (libp2p.PeerID, error)
func PeerFromLibp2p(p libp2p.PeerID) peer.PeerID
```

**Зачем отдельный пакет:** изоляция libp2p; замена транспорта не затрагивает бизнес-логику.

#### `node/`

```go
type Node struct {
    id        *identity.Identity
    transport transport.Transport
    storage   storage.Encoder
    manifest  manifest.Store
    proof     *proof.Verifier
    ledger    *ledger.Ledger
    gossip    gossip.Gossip
    topology  *topology.Checker
    cfg       Config
}

func New(cfg Config) (*Node, error)
func (n *Node) Start(ctx context.Context) error
func (n *Node) Stop() error

// Сценарии
func (n *Node) Backup(ctx context.Context, path string) (ChunkID, error)
func (n *Node) Restore(ctx context.Context, id ChunkID, dest string) error
func (n *Node) Rebalance(ctx context.Context) (RebalanceReport, error)
```

**Инвариант:** `node` не содержит бизнес-логики — только композицию. Паддинг/усечение чанков — здесь, не в `storage`.

---

## 5. Слой хранения

### 5.1 Параметры

| Параметр | Значение | Обоснование |
|---|---|---|
| Схема | **Reed-Solomon 4+3** | Переживает потерю **3 из 7** шардов |
| Накладные расходы | **+75%** | 10 ГБ данных → 17.5 ГБ на дисках |
| Размер чанка | **1 МиБ (максимум)** | Файл режется на чанки ≤ `chunk_size`; последний — меньше |
| Шифрование | AES-256-GCM | Стандарт, аппаратное ускорение |
| Ключ | На файл, не на чанк | Упрощает управление |
| Nonce | 12 байт, уникален на чанк | — |
| Версионирование | По границам файлов | Новая версия хранит изменённые чанки |

**Уточнение v0.1.2.** `chunk_size` — **максимальный** размер чанка, а не фиксированный. Валидация конфига (`chunk_size % (data_shards × block_size) == 0`) относится к **верхней границе**, а не к каждому пользовательскому файлу. Как обрабатываются файлы меньше `chunk_size` и не кратные — см. 5.5.

### 5.2 Почему 4+3, а не 4+2

При 4+2 (6 шардов) отключение 3 узлов из 10 **теряет данные с вероятностью ≈ 16.7%** на чанк (случайная тройка) или **гарантированно** (намеренная тройка). Критерий пилота «отключение 3 нод» невыполним.

При 4+3 (7 шардов) отключение 3 узлов из 10 оставляет **минимум 4 шарда** из 7 — данные восстановимы **всегда**. Цена: +75% диска вместо +50%. Для пилота несущественно.

### 5.3 Размещение

**Правила:**
- Не более **1 шарда одного чанка на узел**
- Все **7 шардов** — на разных пирах
- Кандидаты выбираются через `topology.WouldAccept`
- Если кандидатов меньше 7 → `ErrNoPlacement`

**Открытый вопрос:** размещение «лишних» шардов при цикле из 4 узлов. Рекомендация — у пиров с положительным балансом, не создавая нового цикла.

### 5.4 Инкрементальные бэкапы

Кодирование стирания **плохо дружит с инкрементальными обновлениями**. Решение:
- Чанкование по границам файлов
- Версионирование целыми чанками
- Изменение одного байта → пересчёт всего чанка, но не соседних

### 5.5 Edge-кейсы чанкования

**Проблема.** Валидация конфига требует `chunk_size % (data_shards × block_size) == 0`, но произвольный файл этому не подчиняется. Нужна явная процедура.

**Правило паддинга.** После шифрования ciphertext (`plaintext + 16 байт GCM-тега`) дополняется нулями до `ceil(len / (data_shards × block_size)) × (data_shards × block_size)`.

**Три размера в манифесте:**

| Поле | Смысл |
|---|---|
| `codec.original_size` | plaintext (для GCM) |
| `codec.ciphertext_size` | длина ciphertext **до** паддинга = `original_size + 16` |
| `codec.encoded_size` | длина **после** паддинга (что уходит в RS) |

**Процедура упаковки чанка:**

```
1. plaintext  = chunk bytes (0 ≤ len ≤ chunk_size)
2. ciphertext = AES-GCM-Encrypt(plaintext)  // len + 16
3. padded     = ciphertext || zeros до ceil(len / (4 × 4096)) × (4 × 4096)
4. shards     = RS-Encode(padded)            // 7 × (encoded_size / 4)
5. manifest.Codec.{OriginalSize, CiphertextSize, EncodedSize} = ...
```

**Процедура распаковки:**

```
1. padded     = RS-Decode(≥ 4 shards)        // encoded_size байт
2. ciphertext = padded[:ciphertext_size]
3. plaintext  = AES-GCM-Decrypt(ciphertext)
4. verify len(plaintext) == original_size
```

**Примеры** (`data_shards=4`, `block_size=4096`, `chunk_size=1 МиБ`):

| Случай | Plaintext | Ciphertext | Padded | Шард | Блоков на шард |
|---|---|---|---|---|---|
| Пустой файл | 0 | 16 | 16 384 | 4 096 | 1 |
| Файл 500 КиБ | 512 000 | 512 016 | 524 288 | 131 072 | 32 |
| Ровно 1 МиБ | 1 048 576 | 1 048 592 | 1 064 960 | 266 240 | 65 |
| 1 МиБ + 1 байт | 1 048 577 | 2 чанка: (1 048 576 / 1 048 592 / 1 064 960) + (1 / 17 / 16 384) | — | — | 65 + 1 |

**Следствия:**
- Минимальный `encoded_size` = 16 384 (для пустого файла): накладные — **+1024×**.
- Для файлов ≥ `chunk_size` накладные от паддинга ≤ `4 × block_size / chunk_size` = 16 КиБ / 1 МиБ ≈ 1.6%.
- Число блоков Merkle на шард = `shard_size / block_size`, всегда целое (следует из паддинга).

---

## 6. Слой proof-of-storage

### 6.1 Схема: Merkle-дерево над шардом

**Причина выбора над SNIPS:** SNIPS требует специализированной криптографии и библиотек. Merkle-дерево реализуется в сотни строк, допускает **неограниченное** число проверок и даёт тот же эффект: пир не знает корень, не может подогнать ответ.

### 6.2 Структура

- Шард разбивается на блоки по **4 КиБ**
- `leaf_i = SHA256(0x00 ∥ block_i)` — domain separation
- `parent = SHA256(0x01 ∥ left ∥ right)`
- Нечётное число листьев → последний дублируется
- `merkle_root` хранится в манифесте

**Обоснование 4 КиБ:**

| Размер блока | Глубина дерева | Размер proof | Размер ответа |
|---|---|---|---|
| 512 Б | 9 | 288 Б | ~800 Б |
| **4 КБ** | **6** | **192 Б** | **~4.2 КБ** |
| 64 КБ | 2 | 64 Б | ~64 КБ |

### 6.3 Классификация результатов

| Ситуация | Результат |
|---|---|
| Все proof валидны | `ResultOK` |
| Timeout | `ResultSoftFailure` |
| `nonce` не совпадает | `ResultHardFailure` |
| Path не сходится к root | `ResultHardFailure` |
| Подпись невалидна | `ResultHardFailure` |
| Ответ позже timeout | `ResultSoftFailure` |

---

## 7. Слой экономики (ledger)

### 7.1 Модель: попарный взаимный кредит

- Каждый узел ведёт локальную таблицу `peer.PeerID → {Credit, Debit, Limit, State}`
- **Credit** — сколько мне должны (я хранил для них)
- **Debit** — сколько я должен (они хранили для меня)
- Нет глобального токена, нет блокчейна

### 7.2 Рост лимита: leaky bucket

```
L(t+1) = min(L_cap, γ·L(t) + β·success)
```

| Параметр | Значение (pilot) | Смысл |
|---|---|---|
| `L_cap` | 1 ГиБ | Абсолютный потолок |
| `β` | 100 МиБ | Прибавка за успех |
| `γ` | 0.95 | Коэффициент забывания |

**Свойства:** естественно насыщается, «забывает» старые успехи.

### 7.3 Трёхуровневая эскалация

| Уровень | Условие входа | Действие | Выход |
|---|---|---|---|
| **0 — норма** | Успешность ≥ 95% | — | — |
| **1 — испытательный срок** | 80% ≤ успешность < 95%, **или** 1 hard failure | Лимит заморожен, новые обязательства не принимаются, существующие сохраняются | 7 дней **или** 50 проверок; успешность ≥ 95% → уровень 0 |
| **2 — аннулирование** | Успешность < 80%, **или** 2 hard failure подряд, **или** повторный вход в уровень 1 в течение 30 дней | Ребро аннулируется, цикл аннулируется, лимит × 0.5 | Только через новый старт с нуля |

**Что важно не делать:**
- Не аннулировать данные при провале — удаление шардов отдельное решение владельца
- Не путать soft и hard failures

### 7.4 Старт нового узла

- Нулевой баланс у всех
- Чтобы получить место — **сначала дать своё**
- Нет «кредита доверия» авансом

### 7.5 Связь с `rebalance`

При переходе в `StateAnnulled` ребро аннулируется, шард считается недоступным. Это **триггер для `trd rebalance`** (раздел 12), а не автоматическое действие: rebalance запускается вручную.

### 7.6 Известное ограничение: `ledger` не переживает потерю диска

`ledger` — **локальное** состояние. При потере диска узел теряет:
- свои шарды (для владельцев — потеря),
- свой `ledger` (все балансы),
- `identity.key` (если не забэкаплен) → новый `PeerID`.

Другие узлы **помнят** свои записи о нём. Асимметрия в v0.1.2 **не решается автоматически**. Зафиксированное поведение:

> При потере `ledger` узел начинает с нулевого баланса. Другие узлы, обнаружившие, что пир не отвечает на PoS (hard failure), аннулируют записи о нём через стандартную эскалацию. Старые обязательства считаются «прощёнными» — это цена отсутствия глобального консенсуса в v0.1.2.

Это означает: восстановление узла с нуля проверяет **восстановление сети**, а не экономики. Экономика после такого восстановления начинается заново.

---

## 8. Топология обязательств

### 8.1 Зафиксированное правило

> **Замкнутые циклы обязательств допускаются только длины ≤ 4. Каждое ребро цикла должно иметь актуальный proof-of-storage. Замкнутый цикл без внешнего якоря не учитывается при расчёте кредитного лимита. Открытые цепочки длины ≤ 4 разрешены без дополнительных условий.**

### 8.2 Механика локальной проверки

Перед принятием входящего обязательства от `X`:
1. Ищем направленный путь `me → X` длины ≤ 3 (через gossip глубины 3)
2. Если найден → принятие замкнёт цикл ≤ 4
3. Если нет → принятие создаст цикл ≥ 5 → отказ

### 8.3 Определение якоря

Цикл **заземлён**, если хотя бы один его участник имеет ненулевой баланс с узлом **вне цикла**. Незаземлённый цикл не повышает кредитный лимит.

### 8.4 Обоснование потолка 4

| Аспект | Оценка |
|---|---|
| Локальная выявляемость | Gossip глубины 3 — дешёво |
| Sybil-барьер | Поднимает планку (но не решает сам по себе) |
| Влияние на малые сообщества | Циклы 5+ редки |
| Устойчивость к сговору | Замкнутые циклы уязвимы → нужен якорь |

---

## 9. Gossip

### 9.1 Формат сообщения (v0.1.2)

Полный список `peer.PeerID` **без балансов** — вариант, который позволяет проверять пути локально без round-trip.

**Почему не Merkle:** хэш не позволяет проверить наличие конкретного ребра.

**Почему не Bloom-фильтр:** ложноположительные срабатывания → узел примет обязательство, которое должен был отклонить.

**Когда пересмотреть:** `d > 50` или сотни узлов.

### 9.2 Параметры

| Параметр | Значение |
|---|---|
| TTL | 3 (фиксировано) |
| Интервал re-broadcast | 1 ч |
| Max edges в сообщении | 256 |
| Max размер сообщения | 64 КиБ |
| Rate limit на пира | 10 сообщений/с |

### 9.3 Связь с топологией

**Жёсткая связь:** `gossip.ttl ≥ topology.max_cycle_length - 1`. Валидируется при старте. Если TTL меньше — правило «≤ 4» перестанет работать.

---

## 10. Манифест

### 10.1 Формат (v1)

```json
{
  "version": 1,
  "revision": 0,
  "chunk_id": "<hex 32>",
  "owner": "<peer_id>",
  "created_at": "2026-10-24T12:00:00Z",
  "updated_at": "2026-10-24T12:00:00Z",

  "cipher": {
    "algorithm": "AES-256-GCM",
    "key_id": "<hex 32>",
    "nonce": "<hex 24>"
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
      "shard_id": "<hex 32>",
      "peer_id": "<peer_id>",
      "size": 266240,
      "merkle_root": "<hex 32>",
      "placed_at": "2026-10-24T12:00:00Z"
    }
    // ... всего 7 записей, index 0..6
  ],

  "signature": "<hex 128>"
}
```

**Поля `codec`:**
- `original_size` — plaintext (для GCM).
- `ciphertext_size` — **v0.1.2:** длина ciphertext до паддинга = `original_size + 16`.
- `encoded_size` — длина после паддинга (что уходит в RS).

### 10.2 Правила валидации

1. `version == 1`, `revision ≥ 0`
2. `owner == signer(signature)`
3. `len(shards) == data_shards + parity_shards` (=7)
4. Множество `indices` — ровно `{0..6}` без дубликатов
5. `sum(size) == encoded_size × (data_shards + parity_shards) / data_shards`
6. `shard_id == H(chunk_id ∥ index_byte)`
7. `peer_id` не повторяется в `shards[]`
8. `ciphertext_size == original_size + 16` **(v0.1.2)**
9. `encoded_size ≥ ciphertext_size` **(v0.1.2)**
10. `encoded_size % (data_shards × block_size) == 0` **(v0.1.2)**
11. `size == encoded_size / data_shards` для каждого шарда (следует из 5)
12. `key_id` проверяется владельцем при restore

### 10.3 Derivation shard_id

```
shard_id = SHA256(chunk_id ∥ index_byte)
```

**Важно:** `shard_id` — не checksum, а детерминированный идентификатор. Целостность проверяется через `merkle_root`.

### 10.4 Что **не** хранится в манифесте

| Не хранится | Почему |
|---|---|
| Ключ шифрования | Только `key_id`; ключ отдельно под паролем |
| Challenge-таблица | Заменена Merkle-корнем |
| Балансы | Это `ledger`, не часть контракта |
| Полные шарды | Манифест ~1 КБ; шарды у пиров |
| История ревизий | Только последняя в v0.1.2 |

### 10.5 Хранение шарда на пире

Пир получает:
```json
{
  "type": "store.shard",
  "shard_id": "<hex 32>",
  "block_size": 4096,
  "content": "<hex>",
  "signature": "<hex 128>"
}
```

Пир сохраняет:
- `content` → `shards/<shard_id>.bin`
- `shard_id`, `block_size` → локальный индекс

**Пир не получает `merkle_root`** — только владелец знает корень. Это ключевое свойство: пир не может «подогнать» ответ под известный корень.

### 10.6 Каноническая форма и подпись (v0.1.2)

**Проблема.** Порядок полей в JSON не гарантирован (RFC 8259 §4). Ed25519 над «JSON как есть» недетерминирован: та же структура после сериализации другим кодом даст другую подпись.

**Канонизация (для подписи):**

1. UTF-8, без BOM.
2. Ключи объектов сортируются побайтово по возрастанию на каждом уровне вложенности.
3. Без пробелов, переносов, отступов.
4. Только целые числа (float в манифесте отсутствует).
5. Строки экранируются по RFC 8259 §7 (минимально: `"`, `\`, `\u00XX` для `0x00–0x1F`).
6. Поле `signature` **исключается** из подписываемого контента.

**Domain separation:**

```
sig = Ed25519.Sign(privkey,
    "touchrudder/manifest/v1\n" || canonical_json_without_signature)
```

Префикс защищает от кросс-протокольного переиспользования подписи, если та же ключевая пара подписывает другие типы сообщений (`gossip.Edge`, `proof.Challenge` — у каждого свой префикс).

**Верификация:**
1. Удалить `signature` из манифеста.
2. Канонизировать.
3. `Ed25519.Verify(pubkey, "touchrudder/manifest/v1\n" || canonical, sig)`.

**Реализация:** `manifest/canonical.go` — функция `Canonicalize(m Manifest) ([]byte, error)`. Любая сериализация для подписи идёт **только через неё**, не через `encoding/json` напрямую.

---

## 11. Протокол challenge/response

### 11.1 Challenge (владелец → пир)

```json
{
  "type": "pos.challenge",
  "chunk_id": "<hex 32>",
  "shard_id": "<hex 32>",
  "indices": [42, 1337, 2047],
  "nonce": "<hex 32>",
  "issued_at": "2026-10-24T12:00:00Z",
  "signature": "<hex 128>"
}
```

| Поле | Смысл |
|---|---|
| `nonce` | **16 байт = 32 hex-символа** (v0.1.2: уточнено; раньше `<hex 32>` читалось как «32 байта»). Соответствует `Challenge.Nonce [16]byte`. |
| `signature` | Ed25519 над `"touchrudder/challenge/v1\n" \|\| canonical_json_without_signature` |

### 11.2 Response (пир → владелец)

```json
{
  "type": "pos.response",
  "shard_id": "<hex 32>",
  "nonce": "<hex 32>",
  "proofs": [
    {
      "index": 42,
      "block": "<hex>",
      "path": [
        { "sibling": "<hex 32>", "position": "left" },
        { "sibling": "<hex 32>", "position": "right" }
      ]
    }
  ],
  "responded_at": "2026-10-24T12:00:01Z",
  "signature": "<hex 128>"
}
```

**v0.1.2:** `shard_id` в `Response` — тип `ShardID` (было `PeerID`), см. 4.3.

### 11.3 Верификация

```
leaf = SHA256(0x00 ∥ block)
node = leaf
for step in path:
    if step.position == "left":
        node = SHA256(0x01 ∥ step.sibling ∥ node)
    else:
        node = SHA256(0x01 ∥ node ∥ step.sibling)
assert node == manifest.shards[i].merkle_root
```

Плюс проверки:
- `response.shard_id == challenge.shard_id`
- `response.nonce == challenge.nonce`
- Все `indices` покрыты ровно один раз
- `len(proofs) == k`
- Подпись `response.signature` валидна для `peer_id`
- `responded_at - issued_at ≤ timeout` (30 с)

### 11.4 Параметры

| Параметр | Значение | Обоснование |
|---|---|---|
| `k` | 3 | Вероятность пройти без данных при N=64 ≈ 2⁻¹⁹² |
| `block_size` | 4096 | Баланс глубины и размера ответа |
| `timeout` | 30 с | Достаточно для домашнего интернета |
| `max_response_size` | 64 КиБ | Защита от amplification |
| `nonce_length` | 16 байт (32 hex) | Anti-replay |

---

## 12. Rebalance

### 12.1 Назначение

Восстановление избыточности чанков после потери шардов (hard failure, уход пира, аннулирование ребра). **Ручной в v0.1.2**; автоматический — отложен.

### 12.2 Алгоритм

```
1. Для каждого манифеста:
   a. Собрать список доступных шардов (State != Annulled, PoS OK)
   b. Если len(available) >= data_shards + parity_shards → skip
   c. Если len(available) < data_shards → чанк потерян, пометить как ErrDataLost
   d. Иначе:
      i.   Скачать available шардов (минимум data_shards)
      ii.  Decode → padded → truncate до ciphertext_size → Decrypt → plain
      iii. Encrypt → padded → Encode → полный набор (7 шардов)
      iv.  Для недостающих шардов:
           - выбрать пира через topology.WouldAccept
           - отправить store.shard
           - записать в новый манифест
      v.   manifest.Put(revision+1, новая подпись через canonical)
2. Вернуть RebalanceReport
```

### 12.3 Что меняется в манифесте

- `revision` инкрементируется
- `updated_at` обновляется
- `shards[]` — новые `peer_id` и `placed_at` для восстановленных шардов
- `merkle_root` **не меняется** (шарды детерминированы при том же plain + key + nonce)
- `signature` пересчитывается через каноническую форму (10.6)

### 12.4 Интерфейс

```go
type RebalanceReport struct {
    ChunksScanned    int
    ChunksHealthy    int
    ChunksRebalanced int
    ChunksLost       int
    Errors           []error
    Duration         time.Duration
}

func (n *Node) Rebalance(ctx context.Context) (RebalanceReport, error)
```

CLI: `trd rebalance [--dry-run]`

`--dry-run` — только отчёт, без изменений.

### 12.5 Ограничения v0.1.2

- **Ручной запуск** — не автоматически.
- **Один чанк за раз** — без параллелизма между чанками.
- **Не решает потерю `ledger`** — только восстановление шардов.
- **Не восстанавливает `key_id`** — если ключ шифрования потерян, rebalance невозможен.

---

## 13. Конфигурация

### 13.1 Формат: TOML

**Принципы:**
- Все константы в одном месте
- Три класса: FROZEN / Tunable / Operational
- Секреты — только через env (`TR_*`)
- Один файл — один профиль (`pilot.toml`, `prod.toml`)
- Явное лучше неявного — всё валидируется

### 13.2 Полная структура

```toml
# ============================================================
# TouchRudder node configuration — v0.1.2
# ============================================================

version = 1

# ------------------------------------------------------------
# [identity] — Operational
# ------------------------------------------------------------
[identity]
key_path = "/var/lib/touchrudder/identity.key"
key_file_mode = "0600"

# ------------------------------------------------------------
# [transport] — Operational
# ------------------------------------------------------------
[transport]
listen = ["/ip4/0.0.0.0/tcp/4001", "/ip6/::/tcp/4001"]
announce = []
dial_timeout = "10s"
request_timeout = "30s"
max_connections = 128
enable_relay = true

# ------------------------------------------------------------
# [storage] — FROZEN + TUNABLE
# ------------------------------------------------------------
[storage]
data_dir = "/var/lib/touchrudder/shards"
manifest_dir = "/var/lib/touchrudder/manifests"

# FROZEN — попадают в манифест
data_shards = 4
parity_shards = 3
block_size = 4096
chunk_size = 1048576       # 1 МиБ — МАКСИМУМ, не фиксированное значение

# TUNABLE
max_shards_per_peer = 1
max_concurrent_writes = 8
stream_buffer = 65536

# ------------------------------------------------------------
# [proof] — TUNABLE
# ------------------------------------------------------------
[proof]
k = 3
timeout = "30s"
max_response_size = 65536
nonce_length = 16
check_interval = "24h"
concurrency = 4

# ------------------------------------------------------------
# [ledger] — TUNABLE
# ------------------------------------------------------------
[ledger]
limit_cap = "10GiB"
beta = "104857600"   # 100 МиБ
gamma = 0.99
success_window = 100
threshold_normal = 0.95
threshold_probation = 0.80
probation_duration = "168h"  # 7 дней
probation_checks = 50
repeat_window = "720h"       # 30 дней
hard_fail_streak = 2
annul_penalty = 0.5

# ------------------------------------------------------------
# [gossip] — TUNABLE
# ------------------------------------------------------------
[gossip]
ttl = 3
rebroadcast_interval = "1h"
max_edges_per_message = 256
max_message_size = 65536
rate_limit_per_peer = 10

# ------------------------------------------------------------
# [topology] — TUNABLE (менять с осторожностью)
# ------------------------------------------------------------
[topology]
max_cycle_length = 4
require_anchor = true
anchor_min_balance = "1048576"  # 1 МиБ

# ------------------------------------------------------------
# [node] — TUNABLE
# ------------------------------------------------------------
[node]
backup_paths = ["/home/user/Documents", "/home/user/Projects"]
backup_concurrency = 2
restore_concurrency = 6
backup_timeout = "1h"
restore_timeout = "30m"
manifest_scan_interval = "6h"

# ------------------------------------------------------------
# [logging] — Operational
# ------------------------------------------------------------
[logging]
level = "info"
format = "json"
output = "stderr"
max_size = "100MiB"
max_backups = 5

# ------------------------------------------------------------
# [metrics] — Operational
# ------------------------------------------------------------
[metrics]
enabled = false
listen = "127.0.0.1:9090"
path = "/metrics"
```

### 13.3 Секреты (только env)

| Переменная | Назначение | Обязательна |
|---|---|---|
| `TR_IDENTITY_PASSWORD` | Пароль для расшифровки `identity.key` | Если key зашифрован |
| `TR_BACKUP_KEY` | Мастер-ключ шифрования бэкапов | Для backup/restore |
| `TR_METRICS_TOKEN` | Bearer-токен для `/metrics` | Если metrics enabled |

### 13.4 Env-переопределения

Любое поле можно переопределить через `TR_<SECTION>_<FIELD>`:
- `TR_STORAGE_CHUNK_SIZE`
- `TR_LEDGER_LIMIT_CAP`
- `TR_PROOF_K`

**Исключение:** секреты — только через явные `TR_*_PASSWORD` и `TR_*_KEY`.

### 13.5 Валидация (при старте, до открытия портов)

```go
func (c *Config) Validate() error {
    if c.Version != 1 { return err("unsupported version") }

    // Storage FROZEN
    if c.Storage.DataShards < 1 { return err("data_shards ≥ 1") }
    if c.Storage.ParityShards < 1 { return err("parity_shards ≥ 1") }
    if c.Storage.DataShards + c.Storage.ParityShards > 32 { return err(...) }
    if c.Storage.BlockSize != 4096 { return err("block_size=4096") }
    if c.Storage.ChunkSize < 256*1024 || c.Storage.ChunkSize > 8*1024*1024 {
        return err("chunk_size 256 КиБ – 8 МиБ")
    }
    // v0.1.2: chunk_size — максимум, но верхняя граница всё равно
    // должна быть кратна data_shards × block_size, чтобы
    // накладные от паддинга были предсказуемы.
    if c.Storage.ChunkSize % (int64(c.Storage.DataShards) * c.Storage.BlockSize) != 0 {
        return err("chunk_size кратен data_shards × block_size")
    }

    // Proof
    if c.Proof.K < 3 { return err("k ≥ 3") }
    if c.Proof.NonceLength < 16 { return err("nonce_length ≥ 16") }
    if c.Proof.MaxResponseSize > 1<<20 { return err(...) }

    // Ledger
    if c.Ledger.Gamma <= 0 || c.Ledger.Gamma >= 1 {
        return err("gamma ∈ (0, 1)")
    }
    if c.Ledger.ThresholdProbation >= c.Ledger.ThresholdNormal {
        return err("threshold_probation < threshold_normal")
    }

    // Topology + Gossip — критическая связь
    if c.Topology.MaxCycleLength < 2 || c.Topology.MaxCycleLength > 8 {
        return err("max_cycle_length 2–8")
    }
    if c.Gossip.TTL < c.Topology.MaxCycleLength - 1 {
        return err("gossip.ttl ≥ max_cycle_length - 1")
    }

    // Identity
    if fi, err := os.Stat(c.Identity.KeyPath); err == nil {
        if fi.Mode().Perm() != 0600 {
            return err("identity.key permissions 0600")
        }
    }

    return nil
}
```

### 13.6 Профили

**`pilot.toml`** (10 узлов, дружеский пул):
```toml
[ledger]
limit_cap = "1GiB"
gamma = 0.95
success_window = 20

[proof]
check_interval = "6h"

[transport]
max_connections = 32
```

**`prod.toml`** (сотни узлов):
```toml
[ledger]
limit_cap = "100GiB"
gamma = 0.995
success_window = 200

[proof]
check_interval = "24h"

[transport]
max_connections = 256
```

### 13.7 Классификация параметров

| Параметр | Класс | Когда менять |
|---|---|---|
| `storage.data_shards`, `parity_shards` | FROZEN | Только на пустом узле |
| `storage.block_size` | FROZEN | Только на пустом узле |
| `storage.chunk_size` | FROZEN | Только на пустом узле |
| `proof.k` | FROZEN-soft | Старые challenge невалидны |
| `proof.timeout`, `check_interval` | Tunable | В любой момент |
| `ledger.*` | Tunable | В любой момент |
| `gossip.ttl` | Tunable | В любой момент |
| `topology.max_cycle_length` | Tunable-risky | Уменьшение → старые рёбра невалидны |
| `transport.*`, `logging.*`, `metrics.*` | Operational | В любой момент |

**Правило:** FROZEN-поля дублируются в манифесте. При restore используется значение **из манифеста**, а не из конфига.

### 13.8 Что НЕ должно быть в конфиге

| Не класть | Почему |
|---|---|
| Пароли, ключи | Только env |
| Пути к чужим данным | Конфиг — про **свой** узел |
| Список доверенных пиров | Это gossip-состояние |
| Балансы | Это live state `ledger` |
| Версию протокола libp2p | Фиксирована в коде |

---

## 14. Стратегия тестирования

| Модуль | Тип | Что проверяется |
|---|---|---|
| `peer` | unit | `Parse` принимает валидные multibase, отвергает мусор; `IsZero` |
| `storage` | property-based (rapid) | `Decode(Encode(x)[k:]) == x` для всех подмножеств; 4+3: любая четвёрка из 7 |
| **`storage` (5.5)** | **unit + табличные** | **Пустой файл, 500 КиБ, ровно 1 МиБ, 1 МиБ + 1 байт; `padded % (4×4096) == 0`; `ciphertext_size == original_size + 16`** |
| `manifest` | unit + fuzz | Round-trip сериализации; версии; 7 шардов; валидация полей `codec.*` |
| **`manifest/canonical`** | **unit + golden** | **Golden-файлы: одинаковый манифест → одинаковые байты; перестановка ключей не меняет результат; `signature` исключается** |
| `proof` | unit + табличные | Детерминизм `Verify`; все классы `Result`; `Response.ShardID` — `ShardID` |
| `ledger` | state machine testing | Эскалация, leaky bucket, `Snapshot`/`Restore` идемпотентны; миллионы сценариев |
| `gossip` | in-memory mock | `KnownPath` на глубине 3; TTL; отсутствие ложных путей |
| `topology` | unit + property | Правило ≤ 4; якорь; консерватизм при пустом gossip |
| `node` | integration (10 in-process) | Все сценарии из раздела 16 |
| `rebalance` | integration | Восстановление после потери 1–3 шардов; идемпотентность `--dry-run` |

**Ключевое:** `ledger` и `topology` тестируются **без сети и диска** — это возможно благодаря чистым интерфейсам.

### 14.1 Тестовые векторы

`testdata/manifest_vectors.json` — детерминированные входы/выходы для проверки совместимости любой реализации. Для v0.1.2 — 7 шардов, `encoded_size = 1064960` для 1 МиБ plaintext (учтён паддинг и GCM-тег).

Отдельный файл `testdata/canonical_vectors.json` — эталонные канонические байты для набора манифестов (включая граничные случаи: пустой файл, unicode-строки, 7 шардов с разными `peer_id`).

---

## 15. Порядок сборки

| Неделя | Что | Артефакт |
|---|---|---|
| 0 | Обновление спецификации до v0.1.2 | Этот документ |
| 1 | **`peer/`** + `storage` (4+3 + паддинг) + `manifest` (с `ciphertext_size`) | Работающий кодек, in-memory хранилище |
| 1 | **`manifest/canonical`** + golden-тесты | Канонизация |
| 2 | `ledger` (state machine + эскалация) | Самая сложная логика |
| 3 | `proof` (Challenger + Verifier) | С in-memory `ShardReader` |
| 4 | `identity` + `transport` (с конвертерами `peer.PeerID ↔ libp2p`) + `gossip` | Работающий хост |
| 5 | `topology` + `node` (Backup/Restore) | Базовые сценарии |
| 6 | `node.Rebalance` + `cmd/cli` | `trd rebalance`, CLI |
| 7 | Integration-тесты на 10 in-process узлах | Все сценарии A, B, C, D |
| 8 | Развёртывание пилота | 10 реальных узлов |

---

## 16. План пилота

### 16.1 Организация

| Параметр | Значение |
|---|---|
| Узлов | 10 |
| Участники | Друзья (закрытый пул) |
| Длительность | 6 недель |
| Владельцы данных | Каждый участник — и владелец, и хранитель |
| Рекомендуемый uptime | ≥ 12 ч/день |
| Уведомление о выходе | За неделю + `trd rebalance` |

**Юридическое:** в пилоте **не бэкапить персональные данные** — только код, фото без метаданных, документы без ПДн. Это снимает 90% вопросов по 152-ФЗ.

**Конфиденциальность:** друзья видят размер ваших бэкапов, частоту, IP-адреса. Содержимое — нет (шифрование). Проговорить с участниками **до** старта.

### 16.2 Сценарии

#### Группа A: базовая функциональность

| # | Сценарий | Ожидаемый результат |
|---|---|---|
| A1 | Backup файла 100 МБ | Манифест создан, 7 шардов на 7 разных узлах, `trd status` = healthy |
| A2 | Restore 100 МБ сразу после backup | SHA-256 совпадает |
| A3 | Backup 10 ГБ | Все чанки размещены, укладывается в `backup_timeout` |
| A4 | Restore 10 ГБ | SHA-256 совпадает; время измерено |
| A5 | Backup при 6 доступных узлах (4+3) | `ErrNoPlacement` — корректный отказ |
| A6 | Restore при 4 доступных шардах из 7 | Успех |
| A7 | Restore при 3 доступных шардах из 7 | `ErrDataLost` — корректный отказ |
| **A8** | **Backup пустого файла** | **Успех; `encoded_size = 16384`, `ciphertext_size = 16`** |
| **A9** | **Backup файла 1 МиБ + 1 байт** | **2 чанка; restore даёт SHA-256 исходного** |

#### Группа B: отказы и восстановление

| # | Сценарий | Ожидаемый результат |
|---|---|---|
| B1 | Отключить 1 узел, 2 интервала PoS | `ledger`: soft failures, `Probation`, status=degraded |
| B2 | Отключить 1 узел навсегда, `trd rebalance` | Избыточность 7/7 на новых пирах, `revision++` |
| B3 | Отключить 2 узла, `rebalance` | 7/7; restore работает |
| B4 | **Отключить 3 узла (4+3), restore** | **Успех** — восстановление из 4 шардов |
| B5 | Отключить 3 узла, `rebalance` | 7/7 |
| B6 | Восстановить узлы с `identity.key`, но пустым диском | Тот же `PeerID`, без шардов; PoS → hard failure → `rebalance` |
| B7 | Восстановить узел с нуля (диск + key) | Новый `PeerID`; `ledger` с нуля; `rebalance` компенсирует |

#### Группа C: экономика и топология

| # | Сценарий | Ожидаемый результат |
|---|---|---|
| C1 | Новый узел без кредита делает backup | `ErrInsufficientCredit` |
| C2 | Узел даёт место, лимит растёт | Насыщение у `L_cap` |
| C3 | Узел перестаёт отвечать на PoS | Эскалация → `Annulled` |
| C4 | Узел «мигает» (85% успешности) | `Probation`, заморозка, выход через 7 дней |
| C5 | Цикл из 4 узлов | Принят при наличии якоря |
| C6 | Цикл из 5 узлов | `RejectCycle` |
| C7 | Цикл из 4 без якоря | `RejectUnanchored` |

#### Группа D: атаки и злоупотребления

| # | Сценарий | Ожидаемый результат |
|---|---|---|
| D1 | Один участник с 3 фейковыми узлами | Попарные балансы не дают кредита у честных |
| D2 | Фейковый цикл из 4 узлов | Без якоря не учитывается; с якорем PoS требует хранения |
| D3 | Неверные блоки в ответ на PoS | `HardFailure` → эскалация |
| D4 | Replay старого challenge | Отклонён по `nonce` |
| D5 | MITM подменяет ответ | Отклонён по подписи |
| D6 | Спам gossip | Rate limit срабатывает, узел стабилен |
| **D7** | **Манифест с переставленными полями JSON** | **Подпись валидна после канонизации; golden-тест проходит** |

### 16.3 Метрики

#### Функциональные

| Метрика | Как измерять | Целевое |
|---|---|---|
| Backup throughput | ГБ/мин | ≥ 0.5 ГБ/мин |
| Restore throughput | ГБ/мин | ≥ 1 ГБ/мин |
| Целостность | SHA-256 | 100% |
| Успешность backup | Доля успешных | ≥ 95% |
| Успешность restore (≥4 шардов) | Доля успешных | 100% |

#### Отказоустойчивость

| Метрика | Как измерять | Целевое |
|---|---|---|
| Время обнаружения отказа | До первого soft failure | ≤ 6 ч + 30 с |
| Время восстановления | `rebalance` → 7/7 | ≤ 1 ч на 10 ГБ |
| MTTR | Среднее по B-сценариям | ≤ 2 ч |
| Потери при 3 отказах | Чанков с <4 шардами | **0** |

#### Экономические

| Метрика | Целевое |
|---|---|
| Ложные soft failures | ≤ 5% |
| Ложные эскалации | ≤ 1 за пилот |
| Насыщение лимита | ≤ 2 недели |

#### Топологические

| Метрика | Целевое |
|---|---|
| Принятые циклы ≤4 | ≥ 3 |
| Отклонённые циклы ≥5 | ≥ 1 |
| Ложные отклонения | **0** |
| Gossip-трафик | ≤ 10 КБ/узел/ч |

#### Операционные и человеческие

| Метрика | Как измерять |
|---|---|
| Uptime узлов | % времени онлайн |
| Дисковое использование | ГБ на узел |
| Сетевой трафик | ГБ/узел/неделя |
| NPS участников | Опрос в конце |
| Время до первого backup | От получения инструкции |

### 16.4 Критерии успеха

Все должны выполниться:

1. **B4** пройден: 3 отказа — данные восстановимы.
2. **B5** пройден: `rebalance` восстанавливает 7/7.
3. **B7** пройден: восстановление с нуля не ломает сеть.
4. **A3/A4** пройдены: 10 ГБ укладываются в таймауты.
5. **A8/A9** пройдены: пустой файл и файл 1 МиБ + 1 байт.
6. **C5/C6/C7** пройдены: правило циклов работает.
7. **D1/D2** пройдены: Sybil не даёт преимущества.
8. **D7** пройден: подпись устойчива к перестановке полей.
9. Целостность **100%** во всех успешных restore.
10. Не менее **8 из 10** узлов дожили до конца.

### 16.5 Критерии провала

Любой из:

1. Потеря данных при **2** отказах (баг в erasure coding или размещении).
2. `rebalance` не восстанавливает избыточность.
3. Ложные эскалации > 3 за пилот.
4. Ложные отклонения циклов > 0.
5. Готовность продолжить < 50%.
6. Восстановление с нуля ломает ledger у других (рассинхронизация).

### 16.6 Таймлайн

| Неделя | Фаза | Что делаем |
|---|---|---|
| 0 | Подготовка | Спецификация v0.1.2, реализация |
| 1 | Развёртывание | 10 узлов, `trd status` = сеть |
| 2 | Функциональность | A1–A9, throughput |
| 3 | Отказы | B1–B5, MTTR |
| 4 | Восстановление | B6–B7, документирование ограничений |
| 5 | Экономика и атаки | C1–C7, D1–D7 |
| 6 | Анализ | Свод метрик, go/no-go для v0.2 |

**Контрольные точки:**
- Конец недели 2: если A3/A4 не укладываются — пересмотр `chunk_size`/`concurrency`.
- Конец недели 3: если B4 провален — стоп, разбор erasure coding.
- Конец недели 4: если B7 ломает ledger — решение о бэкапе ledger в v0.2.
- Конец недели 6: go/no-go для v0.2.

---

## 17. Открытые вопросы

| Вопрос | Отложено до | Куда ляжет |
|---|---|---|
| Восстановление `identity.key` | v0.2+ | `identity/` + новый `recovery/` |
| Merkle + on-demand proof для gossip | Когда `d > 50` | `gossip/` |
| Мультипротокольная совместимость | v0.2+ | `transport/` (версия в handshake) |
| Placement-политика (выбор узлов) | После пилота | `node/placement.go` → `topology/placement` |
| UI/TUI | После пилота | `cmd/cli` |
| Дедупликация между владельцами | v0.2+ | `manifest/` + `storage/` |
| История ревизий манифеста | v0.2+ | `manifest/` (журнал) |
| Делегированная верификация | v0.2+ | Новый `trustees/` |
| Совместимость `block_size` | При первом изменении | `proof/` |
| Автоматический rebalance | v0.2+ | `node/` (фоновый loop) |
| Бэкап `ledger` | v0.2+ | `ledger/` (экспорт/импорт) |

---

## 18. Журнал изменений

### v0.1.2 (2026-10-24)

**Исправлено:**

- **P0.1.** `proof.Response.ShardID` имел тип `PeerID` — исправлено на `ShardID`. Баг не ловился компилятором, если оба типа — `type X string`, но `response.shard_id == challenge.shard_id` могло давать тихую ошибку в рантайме.
- **Nonce.** В 4.3 `Challenge.Nonce` — `[16]byte`, в 11.1 JSON — `<hex 32>`. Уточнено: **16 байт = 32 hex-символа**.

**Добавлено:**

- **`peer/` (4.3.0).** Явный тип `peer.PeerID` без импорта libp2p. Все контракты (`manifest`, `ledger`, `gossip`, `topology`, `transport`, `identity`) переведены на `peer.PeerID`. Конвертация `↔ libp2p.PeerID` — только в `transport/`.
- **5.5 «Edge-кейсы чанкования».** Правило паддинга, три размера в манифесте (`original_size` / `ciphertext_size` / `encoded_size`), процедуры упаковки/распаковки, таблица примеров (пустой файл, 500 КиБ, 1 МиБ, 1 МиБ + 1 байт).
- **10.6 «Каноническая форма и подпись».** Правила канонизации JSON, domain separation (`"touchrudder/manifest/v1\n"`), `manifest/canonical.go` как единственная точка сериализации для подписи.
- **Тестовые векторы:** `testdata/canonical_vectors.json` — эталонные канонические байты.
- **Сценарии пилота:** A8 (пустой файл), A9 (1 МиБ + 1 байт), D7 (перестановка полей JSON → подпись валидна).

**Обновлено:**

- **1.2 Scope** — добавлены паддинг, канонизация, `peer.PeerID`.
- **3.3** — «1 МиБ (максимум)», накладные «+75% (до +79% для мелких файлов)».
- **4.2** — `peer/` как чистый пакет; `manifest`/`ledger` могут импортировать `peer/`.
- **4.1** — ASCII-схема зависимостей.
- **5.1** — уточнение «`chunk_size` — максимум».
- **10.1** — добавлено `codec.ciphertext_size`; пример `encoded_size = 1064960` (учтён GCM-тег и паддинг; в v0.1.1 было ошибочно `1835008`).
- **10.2** — правила валидации 8–11 (`ciphertext_size`, `encoded_size`, кратность).
- **12.2** — алгоритм rebalance учитывает truncate до `ciphertext_size`.
- **12.3** — подпись пересчитывается через каноническую форму.
- **13.2** — комментарий к `chunk_size`: «МАКСИМУМ, не фиксированное значение».
- **13.5** — комментарий к валидации `chunk_size`.
- **14** — тесты `peer`, `storage` (5.5), `manifest/canonical`, `proof` (`Response.ShardID`).
- **15** — порядок сборки: `peer/` и `manifest/canonical` на неделе 1.
- **16.2** — сценарии A8, A9, D7.
- **16.4** — критерии успеха 5 и 8.

### v0.1.1 (2026-10-21)

**Изменено:**
- `storage.parity_shards`: **2 → 3**. Схема **4+2 → 4+3**. Шардов на чанк: **6 → 7**.
- Обоснование: критерий пилота «отключение 3 узлов» невыполним при 4+2 (вероятность потери ≈ 16.7% на чанк при случайной тройке; 100% при намеренной). При 4+3 потеря 3 из 10 оставляет минимум 4 шарда — данные восстановимы **всегда**.
- Накладные расходы: **+50% → +75%**.

**Добавлено:**
- Раздел 12 «Rebalance» — ручной `trd rebalance`.
- Раздел 16 «План пилота».
- Раздел 7.6 — ограничение: `ledger` не переживает потерю диска.

**Обновлено:**
- Манифест: 7 шардов.
- Конфиг: `parity_shards = 3`.
- Стратегия тестирования: `rebalance`.
- Порядок сборки.
- Открытые вопросы.

### v0.1 (2026-10-21)

Первая версия: схема 4+2, три слоя, 9 модулей, план пилота в зачаточном виде.

---

## 💎 Итог

Спецификация v0.1.2 **замкнута** и готова к реализации:

- **Название** — легенда варианта C, три версии, готовые ответы
- **Архитектура** — три слоя, всё проверяется локально
- **Модули** — **10 пакетов** (добавлен `peer/`), чистые `ledger`/`topology`/`storage`
- **Хранение** — **RS 4+3**, AES-256-GCM, 1 МиБ чанк (максимум), **явный паддинг**
- **Proof** — Merkle 4 КиБ, k=3, 30 с, `Response.ShardID` исправлен
- **Экономика** — попарный кредит, leaky bucket, трёхуровневая эскалация
- **Топология** — циклы ≤ 4 с якорем
- **Gossip** — полный список peer-ID, TTL=3
- **Манифест** — JSON v1, **канонизация + domain-separated Ed25519**, `ciphertext_size`
- **Rebalance** — ручной, через `trd rebalance`
- **Конфиг** — TOML v1, три класса параметров, секреты через env
- **План пилота** — 10 узлов, 6 недель, сценарии A–D (включая edge-кейсы), метрики, критерии

**Следующий шаг** — начало реализации по разделу 15: неделя 1 — `peer/` + `storage` (4+3 + паддинг) + `manifest` (с `ciphertext_size` и `canonical`).