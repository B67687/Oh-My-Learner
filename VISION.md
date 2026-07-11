# oh-my-learner — VISION

**CLI**: `learn`  
**Repo**: `oh-my-learner`

## One Sentence

A CLI that generates practice problems from templates and schedules them with spaced repetition + interleaving. Subject-agnostic. No hand-written cards.

## Why This Exists

Every flashcard tool makes you write your own cards. That's the hard part — so students don't do it. `learn` generates problems from templates so you spend your energy *answering*, not *writing*. The FSRS spaced repetition scheduler ensures you see each problem at the optimal moment for retention. Interleaving (mixing subjects/problem types) is built into every session per robust learning science evidence.

Learning science (Dunlosky et al., 2013) shows only two study techniques have high utility: **practice testing / active recall** and **distributed practice / spaced repetition**. This tool does both, plus interleaving (Bjork 1992, Rohrer 2012).

## Fixed Goal (Frozen — Will Not Change)

### Scope (yes)

| What | How |
|---|---|
| What | How |
|---|---|
| Interface | **CLI first**. TUI comes only after CLI is solid (`bubbletea` consumes `core` package, no rewrite) |
| Scheduler | **SM-2** — proven algorithm, ~80 lines of math, no external dep needed |
| Interleaving | **Main feature** — sessions actively mix subjects and problem types per Bjork (1992), Rohrer (2012) |
| Problem generation | Template-based with `text/template` + pre-process — never hand-write a card |
| Storage | Local SQLite (`modernc.org/sqlite` — pure Go, no CGO). Full offline. |
| Subjects | Pluggable template packs (TOML) — local files |
| Gamification | Streaks only (loss aversion — "don't break the chain"). No points, no badges, no leaderboards |
| Config | `~/.config/oh-my-learner/config.toml` |
| Standards | Audit against standards framework (ADR, changelog, CI, commit conventions, etc.) |
| Retrospectives | Apply retrospective methodology after each milestone |

### Anti-Scope (won't ever do)

- ❌ No web UI, no GUI
- ❌ No Anki import/export
- ❌ No shared decks or community hub
- ❌ No cloud sync or accounts
- ❌ No mobile app
- ❌ No WASM/plugin system — subjects are static template packs
- ❌ No AI/LLM generation — templates are deterministic
- ❌ No points, badges, leaderboards, or competitive features

### Architecture

oh-my-learner/
├── core/        # Library: scheduler, templater, storage
├── cmd/         # Binary: `learn` command (cobra)
├── main.go      # Entry point
└── subjects/    # Template packs per subject
```

`core` is a reusable internal package. CLI is a thin cobra wrapper around it.

### Stack

| Component | Choice | Rationale |
|---|---|---|
| Language | **Go 1.24+** | Faster iteration, better agent debugging, pure Go libs (no CGO) |
| Scheduler | **SM-2 (own impl)** | Proven algorithm, ~80 lines of math, no external dep |
| Database | `modernc.org/sqlite` | Pure Go SQLite, no CGO, cross-compile friendly |
| CLI | `spf13/cobra` | Industry standard for subcommands, active 2026 |
| Templates | `text/template` (stdlib) + pre-process | Zero deps, random var selection done in Go before render |
| Config | `pelletier/go-toml/v2` | Fast, tree API, active 2026 |
| TUI (future) | `charmbracelet/bubbletea` | Modern Elm-arch TUI, growing ecosystem |
| Errors | Stdlib (`errors`, `fmt`) | Go 1.24 `errors.Join`, proper `%w` wrapping |

### Success Criterion

```
$ learn add algorithms
✓ Installed algorithms (12 templates)

$ learn status
  algorithms: 5 due today, 12 mastered, 31 total

$ learn review
  1/5 ── algorithms ──
  Q: What is O(n log n) worst-case for comparison-based sorting?
  [press Enter to reveal answer]
  A: Comparison sort lower bound is n log n via decision tree.
  Quality (0-5): 4
  Next review: in 3 days
```

## What This Is Not

This is not an Anki competitor. This is not a flashcard app. This is not a study social network. This is a focused CLI tool that generates practice problems and schedules them with interleaving. Nothing more.

## The Contract

This goal will not move. No scope creep. No feature bloat. The ithmb-codec lesson: a project that tries to be everything becomes nothing. This tool does one thing and does it well.
