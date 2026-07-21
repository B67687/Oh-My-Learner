# Architecture

Oh-My-Learner is a Go CLI study tool built around spaced repetition. It
generates practice problems from parameterized templates or AI-synthesized
cards, then schedules them with SM-2 and research-backed interleaving.

## Macro Architecture

```
oh-my-learner/
├── main.go              # Entry point. Calls cmd.Execute().
├── cmd/                 # CLI layer: Cobra commands (10 files)
│   ├── root.go          # Root command, subcommand registration
│   ├── add.go           # Install subject packs (TOML or --ai)
│   ├── review.go        # Review session with interleaving
│   ├── explore.go       # Topic map with card counts and due status
│   ├── report.go        # Streak, retention, daily activity
│   ├── hook.go          # Shell/tmux prompt hooks for due counts
│   ├── map.go           # Prerequisite dependency graph
│   ├── status.go        # Due counts per subject
│   ├── config.go        # View or initialize config.toml
│   └── helpers.go       # Shared stdin reader, config path helpers
├── core/                # Business logic package
│   ├── core.go          # Types: CardState, Template, KnowledgeType, Scheduler interface
│   ├── scheduler.go     # SM-2 algorithm with Scheduler interface
│   ├── scheduler_test.go
│   ├── storage.go       # SQLite persistence (modernc.org/sqlite, no CGO)
│   ├── storage_test.go
│   ├── templater.go     # TOML subject pack loading and template rendering
│   ├── templater_test.go
│   └── core_test.go
├── agent/               # AI card generation
│   ├── agent.go         # DeepSeek API client (net/http only)
│   └── generator_test.go
├── subjects/            # TOML subject packs
│   ├── operating-systems.toml
│   └── algorithms.toml
└── docs/                # Documentation
```

### Package Map

```
main.go
  └── cmd/          (Cobra commands, CLI parsing, I/O)
        └── core/   (Scheduler, Storage, templater, types)
              └── agent/  (DeepSeek API client, response parsing)
```

Three layers. `cmd/` calls `core/`. `core/` is independent of `cmd/`.
`agent/` is independent of both -- `cmd/add.go` wires them together.

## Data Flow

### TOML Pack Path

`learn add operating-systems`

1. `cmd/add.go` calls `findSubjectPack("operating-systems")` which searches
   `subjects/` relative to CWD or the binary's directory.
2. `os.ReadFile` loads the TOML bytes.
3. `core.LoadSubjectPack` parses TOML into `[]core.Template`. Each template
   carries an ID, type (standard/code-trace/debug-find/explain-why), knowledge
   type (declarative/procedural), question/answer text, and optional variables.
4. `core.SubjectPackMeta` extracts the subject name and prerequisites.
5. `store.UpsertSubject` inserts or updates the subject row.
6. `store.SetPrerequisites` replaces the prerequisite list.
7. For each template, `store.InsertCard` creates a `CardState` with SM-2
   defaults (EF=2.5, interval=0, due=now) and persists the full template data
   to the `cards` table.

### AI Generation Path

`learn add operating-systems --ai`

1. `cmd/add.go` creates `agent.NewDeepSeekAgent()` using `OML_DEEPSEEK_KEY`.
2. `agent.GenerateCards(topic)` builds a system prompt describing the four card
   types and knowledge types. Sends it to `api.deepseek.com/v1/chat/completions`
   via stdlib `net/http`.
3. The response is parsed by `parseCardsResponse`, which strips markdown code
   fences (if the AI adds them despite instructions), unmarshals JSON, and
   validates each card's type, knowledge type, and non-empty content.
4. Back in `cmd/add.go`, each `AIGeneratedCard` is mapped to a `core.Template`
   with an `ai-{subject}-{uuid}` ID. Variables are nil (single-use cards).
5. `store.InsertCard` persists each one. Subject is upserted with the topic
   name.

### Review Path

`learn review`

1. `cmd/review.go` opens storage and calls `store.DueCards(time.Now())` which
   queries all cards where `next_review_at <= now`, ordered by `RANDOM()`.
2. If the count exceeds `daily_review_limit` (default 50), a random subset is
   selected. This prevents pile-up after absence.
3. Each card is loaded with its template via `store.GetCardWithTemplate` (called
   once during loading, stored in struct — avoids a duplicate DB query per card).
   Cards are split by knowledge type:
   - Procedural: shuffled across subjects (interleaved).
   - Declarative: subjects shuffled, cards within each subject kept together
     (blocked by subject).
4. Session order: procedural first, then declarative.
5. Per-card loop:
   a. `core.RenderTemplate` substitutes random variable bindings into the
   question and answer templates using Go's `text/template`.
   b. Question is displayed. Type-specific formatting applies (code blocks for
   code-trace and debug-find, "Explain why:" prefix for explain-why).
   c. User presses Enter to reveal the answer.
   d. If mode is "normal" (not "speed"), a self-explanation prompt is shown:
   "Why is this correct?" The user can press Enter to skip.
   e. User rates recall quality 0-5.
   f. `(&SM2Scheduler{}).ReviewCard` computes:
   | Input | Effect |
   | quality >= 3 (passing) | repetition++, interval grows by EF |
   | quality 0-2 (failing) | repetition=0, interval=1 day, EF=max(1.3, EF-0.2) |
   g. `store.UpdateCardState` persists the scheduling fields.
   h. `store.InsertReview` records the review event for analytics.
6. After the session, streak is updated:
   `store.UpdateStreak(today)`
   - days since last review <= 1: increment streak
   - days since last review <= 3: forgive (preserve streak)
   - days since last review > 3: reset streak to 1
     Then `store.LogDailyActivity` records reviewed/recalled for analytics.
7. Session summary shows recall ratio and current streak.

## Key Design Decisions

### 1. Go + Cobra

CLI-native, single binary, no GUI complexity. Cobra provides flags, help text,
and subcommand routing with minimal boilerplate. The entire CLI layer is 10
files with no business logic -- each command delegates to `core/` and renders
results.

### 2. SM-2 with Scheduler Interface

The SM-2 algorithm is the default scheduler. It's behind a `Scheduler` interface
with a single method: `ReviewCard(card, quality) CardState`. This makes FSRS
swap-ready: implement the interface, wire it in, done. Tests cover the SM-2
edge cases: perfect recall, second review (6-day jump), third review (EF-based
interval), failed review (reset to 1 day), and the 1.3 EF floor.

### 3. SQLite with modernc.org/sqlite

No CGO. The SQLite driver is a pure Go port. This means the binary compiles
on any Go-supported platform without a C compiler, and there are no native
library dependencies. Schema is auto-migrated on first use with `CREATE TABLE
IF NOT EXISTS`. WAL mode is enabled for concurrent readers.

### 4. Selective Interleaving

Procedural cards (how-to, code tracing, debugging) are shuffled across
subjects. Declarative cards (facts, definitions) are blocked by subject. This
follows a research-backed distinction: interleaving helps procedural
discrimination but harms declarative recall by increasing interference.

### 5. Self-Explanation Prompt

After each answer reveal, the tool prompts the user to explain why the answer
is correct. Self-explanation is one of the highest-effect learning strategies
(Chi et al., 1994). Users can skip it with `--mode speed` for faster sessions.
The response text is prompted but currently discarded (not stored in the DB).
This is a known gap: see SPEC-as-built.md discrepancy #6.

### 6. Backlog Forgiveness

After a multi-day absence, the scheduler caps due cards to
`daily_review_limit` (configurable in `~/.config/oh-my-learner/config.toml`,
default 50). Without this, returning after a week would dump hundreds of cards
at once. The random subset is representative across subjects.

### 7. Shared Stdin Reader

`helpers.go` has a package-level `bufio.Reader` singleton. Earlier versions
created a new reader per `readLine()` call, which caused EOF errors on piped
input. Each new reader consumed up to 4096 bytes from stdin on construction,
draining the pipe for subsequent calls. The singleton fix shares one buffer.

### 8. AI Integration with Stdlib

The `agent/` package uses `net/http` directly. No OpenAI SDK or third-party
API client. The API surface is one struct (`DeepSeekAgent`), one method
(`GenerateCards`), and one response parser. The system prompt and validation
logic live together in one file. If the provider changes, only the endpoint
URL and response shape need updating.

## Database Schema

```
subjects:         id TEXT PK, name TEXT
cards:            id TEXT PK, subject_id, template_type, knowledge_type,
                  template_question, template_answer, variables (JSON),
                  easiness_factor, interval_days, repetition,
                  next_review_at, created_at
reviews:          id TEXT PK, card_id, quality, reviewed_at
subject_prerequisites: (subject_id, prerequisite_id) PK
streak:           id INTEGER PK CHECK(id=1), current_streak, longest_streak, last_review_date
daily_log:        date TEXT PK, cards_reviewed, cards_recalled
```

Indexes: `cards(subject_id)`, `cards(next_review_at)`, `reviews(card_id)`, `subject_prerequisites(subject_id)`

## Module Responsibilities

### cmd/ -- Thin CLI Layer

- Parse flags and arguments via Cobra.
- Open storage, call core functions, display results.
- No business logic. No scheduling math. No database queries outside
  `core.Storage`.
- `helpers.go` handles shared concerns: DB path resolution, subject pack
  lookup, stdin reading, and config loading.

### core/ -- All Business Logic

- **core.go**: Type definitions (`CardState`, `Template`, `KnowledgeType`,
  `ReviewQuality`, `TemplateType`, `Scheduler` interface). Zero dependencies
  beyond `time`.
- **scheduler.go**: SM-2 implementation. Pure math, no I/O.
- **storage.go**: SQLite persistence. Schema migration, CRUD for subjects,
  cards, reviews, streaks, daily logs. All methods on the `Storage` struct.
- **templater.go**: TOML parsing via `go-toml/v2`, template rendering via
  Go's `text/template` with `missingkey=error` to catch unbound variables.

### agent/ -- AI Card Generation

- Single responsibility: call DeepSeek API, return structured cards.
- One file, one exported type (`DeepSeekAgent`), one exported method
  (`GenerateCards`).
- Response parsing tolerates markdown code fences that some models add
  despite JSON-only instructions.
- Validation rejects cards with empty questions/answers, unknown types, or
  unknown knowledge types (defaults invalid types to "declarative").
- No dependency on `core/`. Returns `[]AIGeneratedCard` -- `cmd/add.go`
  maps them to `core.Template`.

## Quality Guarantees

- **26 tests** across all packages: 4 in `agent/`, 16 in `core/` (scheduler,
  storage, templater, types), plus supporting helpers.
- **No panic/unwrap pattern** -- all errors propagate via Go's `error`
  interface. Every `InsertCard`, `UpdateCardState`, and `RenderTemplate`
  call returns an error.
- **Template validation at load time** -- `LoadSubjectPack` rejects invalid
  TOML. `RenderTemplate` fails on unbound variables or empty variable pools.
- **AI response validation** -- `parseCardsResponse` filters out malformed
  cards and errors if no valid cards remain.
- **SQLite WAL mode** -- enables concurrent reads without blocking on writes.
- **Single-writer constraint** -- `db.SetMaxOpenConns(1)` prevents SQLite
  locking issues from concurrent writes.
- **Build gates**: `go build ./...`, `go test ./... -count=1`, `go vet ./...`
  all pass with no CGO required.
