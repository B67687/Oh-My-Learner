# Oh-My-Learner

[![CI](https://github.com/B67687/Oh-My-Learner/actions/workflows/ci.yml/badge.svg)](https://github.com/B67687/Oh-My-Learner/actions/workflows/ci.yml)

A CLI study tool that generates practice problems from templates (or
AI-synthesized cards), then schedules them with spaced repetition and
research-backed interleaving. Built in Go. No card authoring required.

## Quick Start

```bash
# Install a subject pack from its TOML definition
learn add operating-systems

# Or let AI generate cards on the fly
learn add operating-systems --ai

# One-time hook so your shell prompt shows due count
learn hook --shell zsh

# Do your daily reviews
learn review

# Check your streak and retention
learn report
```

## Installation

```bash
go install github.com/B67687/Oh-My-Learner@latest
```

Or build from source:

```bash
git clone https://github.com/B67687/Oh-My-Learner.git
cd Oh-My-Learner
go build -o learn ./main.go
```

## Core Concepts

### AI Card Generation

`learn add <subject> --ai` synthesizes flashcards using DeepSeek via the
built-in agent package. The AI generates cards in all four template formats
(standard, code-trace, debug-find, explain-why) and classifies each card's
knowledge type. This means you can study any topic on demand without waiting
for a curated subject pack.

### Template-Based Generation

Subject packs contain parameterized templates. Each session draws random
variable values so no two reviews feel identical. Four template types:

- **standard** — direct Q&A ("What is a syscall?")
- **code-trace** — "What does this code output?" (the student traces execution)
- **debug-find** — "What is the bug in this snippet?"
- **explain-why** — "Why does this design work / fail?"

### Spaced Repetition

Reviews are scheduled with SM-2, the algorithm made famous by SuperMemo.
The scheduler implements a `Scheduler` interface, so an FSRS-based scheduler
can be swapped in once its parameters are calibrated for the card pool.

### Selective Interleaving

Research shows interleaving improves long-term retention (Rohrer, 2012) but
not all material benefits equally. Oh-My-Learner applies a nuanced rule:

- **Procedural** cards (how-to, code, debugging) are interleaved across
  subjects, maximizing discrimination practice.
- **Declarative** cards (facts, definitions) are blocked by subject,
  reducing interference for pure memorization.

This follows the learning-science distinction between conceptual and
procedural knowledge (Bjork & Bjork, 1992; Soderstrom & Bjork, 2015).

### Self-Explanation Prompt

After each answer reveal, the tool prompts: "Explain why this answer is
correct in your own words." Self-explanation is one of the highest-effect
learning strategies (Chi et al., 1994). Skip it during speed reviews with
`--mode speed`.

### Backlog Forgiveness

After a multi-day absence, the scheduler caps due cards to
`daily_review_limit` instead of dumping the entire backlog on you. This
keeps review sessions manageable and prevents the discouraging wall of
hundreds of overdue cards.

### Streak Tracking

Daily review streaks are tracked with a 1-2 day forgiveness window.
A single missed day is forgiven (streak preserved but not incremented).
Missing 3+ consecutive days resets the streak to 1.

## Commands

| Command                        | Description                                          |
| ------------------------------ | ---------------------------------------------------- |
| `learn add <subject>`          | Install a subject pack from its TOML definition      |
| `learn add <subject> --ai`     | Generate cards for a subject via AI                  |
|                                | `learn review`                                       | Run review session (self-explain on; --mode speed to skip) |
| `learn explore`                | Topic map showing card counts and prerequisite links |
| `learn report`                 | Streak, weekly retention %, and 7-day activity log   |
| `learn hook --shell bash\|zsh` | Shell prompt integration showing due count           |
| `learn hook --tmux`            | Tmux status bar integration                          |
| `learn status`                 | Due counts per subject                               |
| `learn map [subject]`          | Dependency graph of subjects or a single subject     |
| `learn config`                 | View or edit settings                                |

## Subject Packs

A subject pack is a TOML file. Each template declares its knowledge type
so the scheduler knows whether to interleave or block it.

```toml
name = "Operating Systems"
prerequisites = []

[[templates]]
id = "context-switch"
type = "standard"
knowledge_type = "declarative"
question = "What happens during a context switch?"
answer = "The kernel saves the current process state (PCB)..."

[[templates]]
id = "race-condition"
type = "debug-find"
knowledge_type = "procedural"
question = "Find the bug in this concurrent counter increment."
answer = "The increment is not atomic..."
```

### knowledge_type

- **declarative** — facts, definitions, "what is" questions. Blocked by
  subject during review.
- **procedural** — how-to, code tracing, debugging, "why" questions.
  Interleaved across subjects.

## User Workflow

1. `learn add operating-systems` — install a curated subject pack.
2. `learn add operating-systems --ai` — or let AI generate cards for any
   topic immediately.
3. `learn hook --shell zsh` — one-time setup so your terminal prompt shows
   the number of cards due today.
4. `learn review` — run daily reviews. Cards self-organize: procedural cards
   mix across subjects, declarative cards stay grouped by subject.
5. `learn report` — check your streak, weekly retention, and recent activity.

## Research

The full learning-science backing is at `docs/research/learning-science.md`.
This project was built as a test of the [Development Protocol](https://github.com/B67687/Development-Protocol),
a document-driven framework for taking raw intention to finished product
via AI agents.

## License

MIT
