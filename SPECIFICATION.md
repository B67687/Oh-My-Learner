# SPECIFICATION: Oh-My-Learner v3

## 1. Overview

**X:** CS knowledge retention across semesters is poor, causing cumulative re-learning stress before exams.
**Solution:** CLI study tool with AI-generated flashcards, spaced repetition (SM-2 with FSRS path), selective interleaving, and adherence infrastructure.
**User:** NTU CS Y2 student building alongside semester studies.
**Appetite:** 3 weeks.

## 2. Architecture

```
oh-my-learner/
├── cmd/        # 10 CLI commands (thin layer)
├── core/       # Types, SM-2/FSRS scheduler, SQLite storage (5 files)
├── agent/      # DeepSeek AI card generation via net/http
└── subjects/   # TOML subject packs per course
```

**Key design decisions:**

- Go CLI binary (no CGO, no external deps beyond cobra/sqlite/toml)
- Scheduler interface: SM-2 now, FSRS swap-ready
- Knowledge-type-aware interleaving (procedural interleaved, declarative blocked)
- Backlog forgiveness, streak with 2-day grace, shell/tmux hooks

## 3. Intent

Retain CS knowledge between semesters without re-learning. Measured by:

1. Review streak >= 5 days during semester (>3/week)
2. Retention rate > 80% for cards > 30 days old
3. User reports reduced exam stress vs pre-tool baseline

## 4. File Tree

```
subjects/algorithms.toml
subjects/operating-systems.toml
subjects/automata.toml
subjects/software-engineering.toml
subjects/networks.toml
subjects/sth.toml
```

5 subject packs for Y2S1 (OS already exists, 4 new). Each: 10-15 templates, mixed knowledge types.

## 5. CI

- `go build ./...`
- `go test ./... -count=1`
- `go vet ./...`
- Go 1.22/1.23 matrix

## 6. UX

```
learn add algorithms        # Install TOML pack
learn review                # Daily review (interleaved)
learn status --count        # Due count for shell hooks
learn report                # Streak, retention, daily activity
learn explore               # Topic map with card counts
```

## 7. Non-Goals

- Not a web app, mobile app, or GUI
- Not a community platform for sharing packs
- Not Anki-compatible import/export
- Not a tutor or interactive learning system

## 8. Success Metrics

| Metric        | Target                                | When   |
| ------------- | ------------------------------------- | ------ |
| Subject packs | 5/5 Y2S1 courses covered              | Week 2 |
| Test count    | >= 40 (was 38)                        | Week 3 |
| Build         | `go build ./...` clean                | Always |
| FSRS          | Scheduler interface with FSRS backend | Week 1 |
| Self-explain  | Responses stored, reviewable          | Week 1 |

## 9. Timeline

| Week       | Deliverable                                                                            |
| ---------- | -------------------------------------------------------------------------------------- |
| **Week 1** | FSRS scheduler (go-fsrs or custom impl), self-explain storage, all existing tests pass |
| **Week 2** | 4 new subject packs (Automata, SE, Networks, STH), verify AI quality                   |
| **Week 3** | Polish, EXPLAINER, SPEC_SYNC, REVIEW, pack for distribution                            |

## 10. Dependencies

No new external deps beyond what v2 uses:

- `github.com/spf13/cobra`
- `modernc.org/sqlite`
- `github.com/pelletier/go-toml/v2`

FSRS: evaluate `go-fsrs` or implement from spec.

## 11. Design for Change Rules

| Rule                                         | Status                                                                          |
| -------------------------------------------- | ------------------------------------------------------------------------------- |
| Interface (no interface before 2nd consumer) | Scheduler interface exists (SM-2 only) — FSRS planned as 2nd, OK                |
| Test contract over implementation            | All tests behavioral — PASS                                                     |
| Module boundary                              | Package-level in Go — PASS                                                      |
| Size (250/40 LOC)                            | storage.go split to 4 files (max 194 LOC), review.go split (max 150 LOC) — PASS |
| Shippable per cycle                          | Each week ships working features — PASS                                         |
| Appetite before scope                        | 3 weeks, scope adjusted — PASS                                                  |
| AI code same checks                          | vendored code = hand-code — PASS                                                |
| Rule of three                                | No premature abstractions — PASS                                                |
| Core ≠ infrastructure                        | core/ imports sqlite via storage only — PASS                                    |
| Clean backlog                                | No perpetual backlog — PASS                                                     |

## 12. Documentation Strategy

- README.md — install, commands, usage flow (existing)
- docs/GLOSSARY.md — 16 learning science terms (existing)
- docs/ARCHITECTURE.md — macro-to-micro (existing)
- SPEC-as-built.md — spec fidelity tracking (existing)

## 13. AI Attribution

Cards generated by DeepSeek V4 Flash via `agent/` package. No third-party AI training on user data. Subject packs are TOML files — inspectable, editable, version-controllable.

## 14. Verification Checklist

- [ ] FSRS scheduler passes all SM-2 test cases (retention match)
- [ ] Self-explain responses stored and retrievable via `learn report`
- [ ] 5 subject packs installed and reviewable
- [ ] `go build ./... && go test ./... && go vet ./...` all clean
- [ ] Student can do one full day of reviews without errors
- [ ] EXPLAINER, SPEC_SYNC, REVIEW documents produced
