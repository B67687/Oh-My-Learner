# Oh-My-Learner

A CLI study tool that generates practice problems from templates and schedules them with spaced repetition + interleaving.

## Quick Start

```bash
# Install a subject pack
learn add algorithms

# Review due cards
learn review

# Check your progress
learn status
```

## How It Works

Oh-My-Learner uses **active recall** and **spaced repetition** — the two study techniques with the highest evidence for long-term retention (Dunlosky et al., 2013). You don't write cards yourself. Subject packs provide templates that generate randomized practice problems, so you spend your energy answering, not authoring.

### Features

- **Template-based generation** — subject packs contain parameterized templates that create unique practice problems each session
- **SM-2 spaced repetition** — proven algorithm schedules reviews at optimal intervals
- **Interleaving** — sessions mix cards across subjects and template types, improving long-term retention (Bjork 1992, Rohrer 2012)
- **Multiple template types** — standard Q&A, code-trace (what does this output?), debug-find (what's the bug?), explain-why
- **Subject map** — `learn map` shows dependencies between topics

### Commands

| Command | Description |
|---------|-------------|
| `learn add <subject>` | Install a subject pack from `subjects/<subject>.toml` |
| `learn review` | Run an interleaved review session |
| `learn status` | Show due counts and progress per subject |
| `learn map [subject]` | Show subject dependency graph |
| `learn config` | View or edit settings |

### Subject Packs

Subject packs are TOML files in the `subjects/` directory. A minimal example:

```toml
name = "Algorithms"
prerequisites = []

[[templates]]
id = "complexity-sorting"
type = "standard"
question = "What is the worst-case time complexity of {{ algorithm }}?"
answer = "{{ algorithm }} has worst-case complexity of {{ complexity }}."

[templates.variables]
algorithm = ["Bubble Sort", "Merge Sort", "Quick Sort"]
complexity = ["O(n²)", "O(n log n)", "O(n²)"]
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

## Research

Built as a test of the [Development Protocol](https://github.com/B67687/Development-Protocol) — a document-driven framework for taking raw intention to finished product via AI agents. The full research backing is at `docs/research/learning-science.md`.

## License

MIT
