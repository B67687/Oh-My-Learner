# SPECIFICATION.md — The Plan IS the Spec (oh-my-learner)

---

## 0. Constitution

1. **Active recall is the core** — every feature must directly or indirectly increase the amount of active retrieval practice the user does. Passive consumption features are out of scope.
2. **No card-writing burden** — the user's energy goes to answering, not authoring. Template-based generation only.
3. **Desirable difficulty** — features that feel harder (interleaving, mixed subjects) but produce better retention are prioritized over features that feel easier (points, badges).
4. **Evidence-based** — every feature must be supported by at least one study from the learning science research (`docs/research/learning-science.md`).
5. **Offline-first** — no cloud, no accounts, no sync. The tool is a local CLI.

## 1. Overview

**Ambition:** A CLI that generates practice problems from templates and schedules them with spaced repetition + interleaving, helping the user learn more efficiently for university CS (and beyond).

**Success criteria:**
- WHEN a subject template pack is installed THEN the user can start reviewing immediately without writing any cards
- WHEN a review session runs THEN it includes cards from multiple subjects and template types (interleaving)
- WHEN the user has completed 100 reviews THEN retention is >=85% on cards due for review

**OUT OF SCOPE (V1):**
- Web UI or GUI (CLI only)
- Cloud sync or accounts (offline-only)
- Anki import/export
- AI/LLM generation (templates are deterministic)
- Points, badges, leaderboards

## 2. Architecture & Design Decisions

**Decision 1: Go language (unchanged)**
> In the context of an existing ~1,500-line Go codebase with working scheduler, storage, and templater,
> facing the tradeoff between Go (fast, single binary, simple) and Python (richer text ecosystem),
> we decided for **staying with Go**,
> to achieve zero rewrite cost and ship within the 3-week timebox,
> accepting that Go is more verbose for text manipulation tasks.

**Decision 2: SM-2 scheduler (unchanged)**
> In the context of a CLI starting from zero user data,
> facing the tradeoff between SM-2 (simple, proven, works from day one) and FSRS (state-of-the-art, needs 1000+ reviews to train),
> we decided for **SM-2** with a clear upgrade path to FSRS,
> to achieve immediate usable scheduling without a cold-start problem,
> accepting that users switching from Anki will feel the schedule is less precise.

**Decision 3: Add interleaving to the scheduler**
> In the context of strong evidence (Bjork 1992, Rohrer 2012) that interleaving improves long-term retention,
> facing the constraint that the existing SM-2 scheduler schedules cards independently,
> we decided for a **session-level interleaving layer** that draws from all due cards across subjects,
> to achieve the retention benefit without modifying the SM-2 algorithm,
> accepting that sessions will feel harder and require a minimum of 10-15 cards per session.

Alternatives considered:
- Per-subject sessions (no interleaving) — rejected because interleaving is the highest-evidence addition we can make

## 3. File Tree & Module Responsibilities

```
oh-my-learner/
├── main.go — Entry point, thin cobra wrapper
├── cmd/ — CLI commands (cobra)
│   ├── root.go — Root command, global flags
│   ├── add.go — `learn add <subject>` — install template packs
│   ├── review.go — `learn review` — run a review session (UPDATED: interleaving)
│   ├── status.go — `learn status` — show due counts, streak
│   ├── config.go — `learn config` — settings management
│   └── map.go — `learn map` — dependency visualization (NEW)
├── core/ — Library: scheduler, templater, storage
│   ├── core.go — Public API, types
│   ├── scheduler.go — SM-2 algorithm (UNCHANGED — interleaving lives at session level)
│   ├── scheduler_test.go — SM-2 tests (already comprehensive at 158 lines)
│   ├── storage.go — SQLite persistence
│   ├── templater.go — Template-based problem generation
│   └── templater_test.go — Template tests (201 lines)
├── subjects/ — Template packs (TOML)
└── docs/research/learning-science.md — Research backing (NEW, 505 lines)
```

## 4. CI, Tooling & Quality Gates

WHEN a pull request is opened
THEN CI SHALL run `go build ./...`
WHERE compilation fails
THEN CI SHALL fail with exit code 1

WHEN a pull request is opened
THEN CI SHALL run `go test ./... -count=1`
WHERE any test fails
THEN CI SHALL fail with the failing test output

WHEN a pull request is opened
THEN CI SHALL run `go vet ./...`
WHERE vet detects issues
THEN CI SHALL fail with the detected issues

Quality gates: compiles + all tests pass + go vet clean
Release: git tag v{semver} + optional binary build

## 5. Dependencies

| Package | Version | Purpose | Contract |
|---------|---------|---------|----------|
| `spf13/cobra` | latest | CLI framework | Standard cobra command interface |
| `modernc.org/sqlite` | latest | Pure Go SQLite (no CGO) | Standard SQL driver |
| `pelletier/go-toml/v2` | latest | Template pack parsing | Standard TOML marshal/unmarshal |

**No new runtime dependencies** for the V1 additions. Interleaving, new template types, and priority metadata are pure Go.

## 6. UX & Interface Contract

**Entry points:**
- `learn add <subject>` — install a template pack
- `learn review` — start an interleaved review session
- `learn status` — show due counts, streak, progress
- `learn map <subject>` — show topic dependency graph (NEW)
- `learn config` — view/edit settings

**User-facing behavior:**
WHEN a user runs `learn review`
THEN the system SHALL build a session from all due cards across ALL subjects
WHERE there is at least 1 due card
THEN the system SHALL present cards in random order (subject-mixed)
WHERE there are fewer than 10 due cards
THEN the system SHALL prompt "only N cards due. Add more subjects?"

**Error contract:**
| Condition | Error | Handling |
|-----------|-------|----------|
| No subjects installed | "No subjects found" | Show `learn add` help |
| No cards due | "Nothing due. Next review: {date}" | Show status instead |
| Corrupt SQLite DB | "Storage error: {details}" | Suggest restore from backup |
| Invalid template pack | "Invalid pack: {error}" | Show TOML parsing error with line |

**Interface contract:**
Module: scheduler
  Precondition: card exists in storage with valid SM-2 data
  Postcondition: SM-2 parameters updated, next review date calculated
  Invariant: interval_0 < interval_N for N in [0,∞)

Module: session builder (NEW — interleaving layer)
  Precondition: at least 1 card due across all subjects
  Postcondition: returned session has cards from >=2 subjects (if available)
  Invariant: no card appears twice in the same session

## 7. Timeline & Milestones

**Appetite:** 3 weeks before university starts

| Milestone | What ships | Checkpoint | Acceptance |
|-----------|------------|------------|------------|
| **M1 (by end week 1)** | Interleaving + session builder | `learn review` pulls from all subjects mixed | WHEN `learn review` runs with 2+ subjects THEN cards are drawn from both in random order |
| **M2 (by end week 2)** | New template types + priority metadata | Templates: code-trace, debug-find, explain-why | WHEN a template pack declares its type THEN the correct template engine renders it |
| **M3 (by end week 3)** | `learn map` + polish | Dependency graph for any installed subject | WHEN `learn map data-structures` runs THEN it shows topics and their prerequisite edges |

**Circuit breaker:** IF after M1 interleaving degrades SM-2 accuracy (more cards forgotten than expected) THEN the project SHALL reassess whether session-level interleaving conflicts with card-level scheduling.

---

## Verification Checklist

- [x] All sections filled
- [x] Out-of-scope list is non-empty
- [x] Each architecture decision includes a Y-Statement
- [x] Each CI gate is a concrete command
- [x] Timeline has a circuit breaker condition
- [x] Constitution has 5 principles (all research-backed)
- [x] Module contracts specified for scheduler AND session builder
- [x] Error contract covers 4 failure conditions
