# RULES.md — Oh-My-Learner Project Bootstrap Protocol

> Read this at the START of every AI session. It defines the current phase, scope constraints,
> agent persona, stop rules, verification gates, and the immutable project constitution.
> The AI enforces phase/scope boundaries — if asked to do something outside scope
> or current phase, it MUST refuse and explain why.
>
> **Current: PERFECT**

---

## Table of Contents

1. [Project Type Routing](#1-project-type-routing)
2. [Constitution (Immutable)](#2-constitution-immutable)
3. [Phase Definitions](#3-phase-definitions)
4. [V1 Scope & Scope Warps](#4-v1-scope--scope-warps)
5. [AI Persona & Constraints](#5-ai-persona--constraints)
6. [Stop Rules](#6-stop-rules)
7. [Verification Gates](#7-verification-gates)
8. [Test Philosophy](#8-test-philosophy)
9. [Evolution & Phase Exit](#9-evolution--phase-exit)
10. [Known Failure Patterns](#10-known-failure-patterns)
11. [Session Kickoff](#11-session-kickoff)

---

## 1. Project Type Routing

Oh-My-Learner is a STANDARD route project:

```
Project type: I know this domain well (can spec V1 upfront)
Route: STANDARD — Bootstrap → WORK → PERFECT → DISTRIBUTE
```

| Attribute     | Value                                                                   |
| ------------- | ----------------------------------------------------------------------- |
| Project name  | Oh-My-Learner                                                           |
| Type          | Go CLI study tool with AI card generation                               |
| Domain        | Spaced repetition, active recall, CLI tooling — well-understood         |
| Current phase | PERFECT (hardening, quality gates, docs — v2 just completed)            |
| Route         | STANDARD (WORK → PERFECT → DISTRIBUTE — no ITERATE, no DISCOVER needed) |
| Language      | Go 1.25 (pure Go, no CGO)                                               |

**No DISCOVER phase needed** — the spaced-repetition domain is well-understood
(SM-2, FSRS, interleaving research). V1 was already built and proven.

**No ITERATE phase needed** — UX was established in v1. v2 extends the same CLI
with AI generation. No UX loop required.

---

## 2. Constitution (Immutable)

> The Constitution is the project's immutable DNA. It is set once at bootstrap and
> governs ALL generation across ALL phases. The AI must reference the Constitution
> before every significant action. If a proposed action would violate the Constitution,
> the AI MUST refuse.

### Constitution

```
Oh-My-Learner:

1. No CGO — modernc.org/sqlite only (pure Go binary, cross-compile friendly).
2. No external API SDKs — stdlib net/http for AI integration. No OpenAI/Anthropic SDK.
3. SM-2 default, Scheduler interface for future FSRS swap — current implementation
   behind an interface so FSRS can be added without changing callers.
4. All errors propagate — no unwrap/panic pattern. Errors are returned through
   function signatures, caught at the CLI boundary, and displayed as user-facing
   messages. No silent recovery.
5. SQLite WAL mode with max 1 writer — single-user, offline-first, no concurrent
   write contention. modernc.org/sqlite with WAL journaling.
6. Inward dependencies — cmd/ → agent/ → core/. Core has zero knowledge of
   agent/ or cmd/. Agent never imports core/ internals, only its public types.
```

### How the Constitution Works

- **Set once** at bootstrap. Changing the Constitution is a project-wide decision.
- **AI reads it** before every significant action (same as stop rules).
- **If an action violates the Constitution**, the AI refuses regardless of phase.

---

## 3. Phase Definitions

### WORK (Completed)

**Purpose:** Build core features against fixed V1 scope.

**What was done:**

- v1: SM-2 scheduler, SQLite storage, template engine, 6 CLI commands, interleaving
- v2: AI agent (agent/ package), DeepSeek integration, selective interleaving by
  knowledge type, self-explanation prompt, streak tracking, backlog forgiveness,
  8 CLI commands (add --ai, review, status, explore, map, report, hook, config)

**Test rule:** Tests written before implementation (49 tests). Scheduler tests
pre-dated scheduler code. Agent tests use recorded JSON fixtures, not real API calls.

**Quality gate at exit:** `go build ./...` + `go test ./... -count=1` passing.
49 tests passing across 4 packages.

### PERFECT (Current — Active)

**Purpose:** Harden existing code. Enter only when WORK scope is complete.

**Allowed:**

- Fuzz testing, static analysis, audit, benchmarks
- Error handling polish, edge case hardening
- CI pipeline setup (`.github/workflows/ci.yml`)
- Constitution compliance audit
- Documentation hardening (SPEC-as-built.md, EXPLAINER.md sync)
- Review checklist execution (REVIEW.md)

**Not allowed:**

- New features or UX changes. PERFECT is for quality, not scope.
- Architectural changes, new Go dependencies
- AI prompt redesign or new template types

**Quality gates:**

- `go vet ./...` — zero warnings
- `go build ./...` — compiles clean
- `go test ./... -count=1` — all 49 tests pass
- Forbidden-patterns audit (no panics, no CGO imports, no SDK imports)
- Constitution compliance check
- SPEC SYNC: SPEC-as-built.md reflects actual codebase (19 discrepancies catalogued;
  resolve or accept each)

### DISTRIBUTE (Future)

**Purpose:** Package, document, publish.

**Allowed:** README updates, CHANGELOG, cross-compile builds, release tagging.

**Not allowed:** Any code changes.

**Quality gate:** `go build ./...` + `go test ./... -count=1` + spellcheck + format.

---

## 4. V1 Scope & Scope Warps

### IN SCOPE (must ship — all shipped)

- **SM-2 spaced repetition scheduler** with Scheduler interface for FSRS swap path
- **Template-based card generation** (4 types: standard, code-trace, debug-find, explain-why)
- **AI card generation** via DeepSeek V4 Flash (free-tier API, stdlib net/http only)
- **8 CLI commands:** `learn add`, `learn review`, `learn status`, `learn explore`,
  `learn map`, `learn report`, `learn hook`, `learn config`
- **Selective interleaving** — procedural cards interleaved, declarative cards blocked by subject
- **Self-explanation prompt** after each answer reveal (togglable via `--mode speed`)
- **Streak tracking** with 2-day forgiveness window
- **Backlog forgiveness** — daily review cap (default 50) prevents compound punishment
- **Adherence hooks** — shell prompt (bash/zsh) and tmux status bar integration
- **SQLite storage** via modernc.org/sqlite — pure Go, no CGO, WAL mode
- **Knowledge-type awareness** — declarative vs procedural classification on all cards
- **Subject dependency graph** with prerequisite chains and cycle detection

### OUT OF SCOPE (explicitly not for this project)

- **GUI/TUI** — CLI only. No bubbletea, no web UI, no desktop app
- **Web server** — no HTTP server, no REST API, no cloud back end
- **Collaborative features** — single user only. No shared decks, no community hub
- **Mobile apps** — terminal only
- **Anki import/export** — not needed when AI generates everything fresh
- **Non-CS subjects** — cs-first. Finance, medicine, history deferred to future scope
- **Points, badges, leaderboards** — streaks only (loss aversion: "don't break the chain")
- **Plugin/WASM system** — subjects are static template packs or AI-generated. No plugin API
- **Paid API tiers** — free stack only. DeepSeek free tier or local model fallback (llama.cpp)

### Scope Warps

No scope warps recorded. The project stayed within its v2 specification throughout
WORK phase. No scope-warp-log.md needed.

---

## 5. AI Persona & Constraints

**Role:** `Go CLI engineer for a spaced-repetition study tool with AI card generation`

**Autonomy:** `HIGH for feature work within spec | LOW at phase boundaries`

### Constraints (per-project)

- **Language / edition:** Go 1.25 — no CGO, no unsafe
- **Safety rules:** No panics. No unrecoverable errors. All errors propagate through
  function signatures. No `log.Fatal` outside of `main.go`.
- **Quality floor:**
  - `go vet` must pass with zero warnings
  - `go build ./...` must compile
  - `go test ./... -count=1` must pass
  - File size limit: 300 lines per file (agent/agent.go exception: 226 lines, under limit)
- **Dependency policy:**
  - No external API SDKs (OpenAI, Anthropic, etc.) — stdlib `net/http` only
  - No CGO-dependent libraries — `modernc.org/sqlite` is the only DB driver
  - New dependencies require stop-rule approval
- **Testing requirements:**
  - All agent tests use recorded JSON fixtures (`agent/testdata/`), never real API calls
  - Core tests must NOT import agent/ package
  - No test-only changes without corresponding code
- **Documentation requirements:**
  - Doc comments on all exported agent/ functions
  - AI prompt templates documented inline
  - SPECIFICATION.md always reflects the plan; SPEC-as-built.md reflects reality
  - CHANGELOG updated per release (Keep a Changelog format)
- **Tool-first rule:** Use `go fmt` for formatting, `go vet` for static analysis,
  compiler for type checking. Never hand-roll what a deterministic tool handles.
- **Error message style:** Include the values that caused the error, not just a message.
  Never include raw AI responses in user-facing errors (use `OML_DEBUG=true` for that).

### Decision Framework (inviolable priority order)

1. **Correctness** over speed — wrong output at any speed is useless. SM-2 math must be exact.
2. **Consistency** with existing patterns over novel approaches — the codebase is the
   source of truth. Follow v1 patterns in core/, cmd/ conventions.
3. **Simplicity** over complexity unless measured — SM-2 before FSRS, AI native before
   web search, offline-first before cloud sync.
4. **Explicit decisions** over implicit defaults — surface tradeoffs, don't hide them.
   Document architecture decisions in `docs/adr/`.
5. **Test evidence** over intuition — if a test doesn't prove it, it's not done.

---

## 6. Stop Rules

The AI MUST stop and ask before proceeding if ANY of these are true:

- [ ] Task touches **3+ files** in one change → ask for plan approval
- [ ] Task adds a **new Go dependency** → ask for permission
- [ ] Task **deletes or overwrites** existing code → confirm first
- [ ] Task is **outside current phase** → refuse, explain why (e.g., new feature
      requested during PERFECT phase)
- [ ] Task touches **OUT OF SCOPE** → refuse, explain why (e.g., "I can't build a
      web UI — this project is CLI-only per RULES.md section 4")
- [ ] Task would **change V1 scope** → refuse, require conscious scope warp
- [ ] Task violates the **Constitution** → refuse, cite which principle (e.g., "This
      would import an OpenAI SDK, violating Constitution principle 2")
- [ ] Task is **ambiguous** (multiple valid approaches with different trade-offs) →
      present options with tradeoff analysis
- [ ] Task exceeds **200 lines** of new code → propose plan first
- [ ] Task has **no test written first** (in WORK phase) → pause, write test first
      (not applicable in PERFECT — but still think about test coverage)

---

## 7. Verification Gates

| Phase          | Must pass before reporting done                                                                                                                                                              |
| -------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **WORK**       | `go build ./...` + `go test ./... -count=1` passes + tests written BEFORE code                                                                                                               |
| **PERFECT**    | `go vet ./...` zero warnings + `go build ./...` compiles + `go test ./... -count=1` all pass (26) + forbidden-pattern audit (no panic, no CGO, no SDK) + Constitution compliance + SPEC SYNC |
| **DISTRIBUTE** | Spellcheck + link check + format conformance + CHANGELOG updated + cross-compile check                                                                                                       |

### SPEC SYNC (Spec-to-Code Fidelity Gate)

SPEC-as-built.md catalogues 19 discrepancies between SPECIFICATION.md and the
actual codebase. Before entering DISTRIBUTE, each discrepancy must be resolved:

- **Deviations** accepted as design decisions (document in SPEC-as-built.md rationale)
- **Missing features** either implemented or formally deferred with documented reasoning
- **New features** added to SPECIFICATION.md as amendments

See `SPEC-as-built.md` for the full discrepancy catalogue and resolution status.

### PERFECT-Specific Gates

1. **No panic paths** — grep for `panic(` across all `.go` files. Zero allowed.
2. **No CGO references** — `grep 'import "C"'` across all `.go` files. Zero allowed.
3. **No SDK imports** — `github.com/sashabaranov/go-openai` or similar. Zero allowed.
4. **All errors checked** — no `_ = funcCall()` patterns that discard errors.
5. **`go vet` clean** — zero warnings.
6. **SPEC SYNC complete** — SPEC-as-built.md reviewed and current.

---

## 8. Test Philosophy

> Derived from TDD research: "Code without tests is not done. Tests that merely
> confirm what the code already does are not tests — they are tautologies."

### Current State

| Package | Files | Test files | Tests | Coverage focus                                                                                                 |
| ------- | ----- | ---------- | ----- | -------------------------------------------------------------------------------------------------------------- |
| core/   | 4     | 3          | ~18   | SM-2 scheduling (8), storage (insert, due, knowledge-type), templater (4 types), core types                    |
| agent/  | 1     | 1          | 6     | Response parsing with 6 JSON fixtures: valid, empty, garbage, invalid-types, markdown-wrapped, knowledge-types |
| cmd/    | 10    | 0          | 0     | **Known gap** — no CLI integration tests yet                                                                   |

**Total: 49 tests. No real AI API calls in any test.**

### The Rules

1. **Tests first.** In WORK phase, the test is written BEFORE the implementation.
   This was followed for v2 (agent tests written before agent.go was finalized).
2. **Tests verify, not confirm.** A test must FAIL on incorrect output and PASS on
   correct output. Agent tests use recorded fixtures that exercise both paths.
3. **One behavior per test.** Each test verifies exactly one behavior. Test names
   describe the expected outcome.
4. **Edge cases are explicit.** Agent tests cover empty responses, garbage JSON,
   missing fields. Scheduler tests cover EF floor, interval growth, failed reviews.
   Known gaps: storage (update non-existent card), backlog cap with zero cards.
5. **No test-only changes without corresponding code.** Every test has a passing
   implementation.
6. **Regression tests lock bugs.** When a bug is found, write a test that reproduces
   it FIRST. Then fix the code. The test stays as a regression guard.
7. **All agent tests use recorded fixtures.** No real AI API calls. Fixtures live
   in `agent/testdata/*.json`.
8. **Core tests must NOT import agent/.** The engine is AI-independent by design.

### Known Gaps (documented in review-findings.md)

- `cmd/` package has zero tests — no CLI integration tests (MINOR)
- Missing edge case tests: update non-existent card, daily limit with zero cards (MINOR)
- No end-to-end test for the full `learn add → learn review → learn status` flow
- Agent tests use recorded responses — won't catch API format changes automatically

---

## 9. Evolution & Phase Exit

### Phase Exit Checklist

Before transitioning from PERFECT to DISTRIBUTE, run this reflection:

```
Phase Exit: PERFECT → DISTRIBUTE

1. What did we learn in this phase?
   - Domain knowledge: [surprising discoveries about hardening a v2 AI tool]
   - Process: [what worked, what didn't about PERFECT phase rules]
   - Architecture: [decisions made that constrain DISTRIBUTE phase]

2. What should the NEXT phase (DISTRIBUTE) know?
   - Gotchas: [things to watch out for in packaging/release]
   - Open questions: [things still unresolved]
   - Priorities: [what matters most in DISTRIBUTE]

3. Protocol improvement?
   - Did any stop rule fire when it shouldn't have? [adjust rule]
   - Did any stop rule NOT fire when it should have? [tighten rule]
   - Did the phase boundaries hold? [if not, why?]

4. Constitution check?
   - Did any action violate the Constitution? [record and fix]
   - Does the Constitution need updating? [rare — think carefully]

5. Protocol self-audit?
   - Did the protocol's rules HELP in this phase? [which rules?]
   - Did any rule HURT? (slowed things down, blocked useful actions)
   - Was this the right ROUTE for the project type? (STANDARD — correct)
   - Was the timebox appropriate for this phase?
   - Would you use the same phase sequence again?
```

### Related Documents

| Document           | Purpose                                           | Link                                 |
| ------------------ | ------------------------------------------------- | ------------------------------------ |
| SPECIFICATION.md   | Locked plan (14 sections, 552 lines)              | `SPECIFICATION.md`                   |
| SPEC-as-built.md   | As-built discrepancy catalogue (19 items)         | `SPEC-as-built.md`                   |
| VISION.md          | Original ambition and scope contracts             | `VISION.md`                          |
| SPEC-plan.md       | Archived v1 plan                                  | `SPEC-plan.md`                       |
| REVIEW.md          | Meta-review checklist (per Development Protocol)  | `docs/REVIEW.md` (from Dev Protocol) |
| EXPLAINER.md       | Architecture explainer for non-coder verification | `docs/EXPLAINER.md`                  |
| CHANGELOG.md       | Release history (Keep a Changelog)                | `CHANGELOG.md`                       |
| docs/adr/          | Architecture Decision Records                     | `docs/adr/`                          |
| review-findings.md | Review findings (5 FAIL items)                    | `review-findings.md`                 |

---

## 10. Known Failure Patterns

### FP-CAT-1: Scope Expansion

| ID     | Pattern            | Description                                                                                                                      |
| ------ | ------------------ | -------------------------------------------------------------------------------------------------------------------------------- |
| FP-001 | Feature Creep      | AI adds "helpful" features not in scope — e.g., adding a web UI or cloud sync because nothing explicitly forbids them in context |
| FP-002 | Polish Trap        | Polishing before core works — triggering on cosmetic improvements during WORK phase                                              |
| FP-003 | Rabbit Hole        | Deep optimization of something that might be removed (e.g., optimizing SM-2 math further when it works correctly)                |
| FP-004 | Scope Warp Cascade | One expansion leads to another — e.g., adding non-CS subjects leads to requesting domain-specific template types                 |

### FP-CAT-2: Quality

| ID     | Pattern            | Description                                                                                                                         |
| ------ | ------------------ | ----------------------------------------------------------------------------------------------------------------------------------- |
| FP-010 | Tautological Tests | Tests that pass on first run and only confirm what code already does — agent testdata must include failure cases                    |
| FP-011 | Missing Edge Cases | Happy path works, edge cases crash silently — e.g., AI returns empty card list, or SQLite connection fails mid-session              |
| FP-012 | Security Blindness | AI generates code that leaks the API key in error messages or logs                                                                  |
| FP-013 | Dependency Bloat   | Adding a library instead of writing 5 lines of Go — e.g., importing an SDK instead of using stdlib net/http                         |
| FP-014 | Context Decay      | Later AI sessions contradict earlier decisions because context was lost — mitigated by RULES.md, SPECIFICATION.md, SPEC-as-built.md |

### FP-CAT-3: Process

| ID     | Pattern              | Description                                                                                                           |
| ------ | -------------------- | --------------------------------------------------------------------------------------------------------------------- |
| FP-020 | Phase Drift          | Working on DISTRIBUTE tasks (README polish, changelog) during PERFECT phase without realizing it                      |
| FP-021 | Silent Pivot         | Changing the AI backend from DeepSeek to another provider without documenting or approving the change                 |
| FP-022 | Assumption Hardening | Early assumptions become locked-in — e.g., assuming AI always returns valid JSON                                      |
| FP-023 | Review Debt          | AI generates more code than can be reviewed — cmd/ package (10 files, 0 tests) is an existing gap                     |
| FP-024 | Confident Wrongness  | Code compiles, runs, and is subtly incorrect — e.g., SM-2 interval math off by one, or streak forgiveness logic wrong |

### FP-CAT-4: Protocol Governance

| ID     | Pattern             | Description                                                                            |
| ------ | ------------------- | -------------------------------------------------------------------------------------- |
| FP-030 | Rule Rigidity       | Protocol rules that help general cases actively slow down specific project types       |
| FP-031 | Over-governance     | Spending more time managing the protocol than building the product                     |
| FP-032 | Self-Audit Skipping | Rushing phase exits without running the self-audit (Phase Exit Checklist in section 9) |
| FP-033 | Routing Error       | Choosing the wrong route at bootstrap — Oh-My-Learner chose STANDARD, which is correct |

### Using Failure Patterns

When the AI recognizes a failure pattern, it MUST:

1. Flag it: "Warning: this looks like FP-001 (Feature Creep)."
2. Explain why: "You asked for a new card template type, but this would modify the
   core template engine during PERFECT phase. New template types are not in scope."
3. Stop and ask: "Should I continue with this, or revert to the original scope?"

---

## 11. Session Kickoff

Every AI session starts with:

```
"Read RULES.md.
State current phase (PERFECT) and what that means I can/cannot do.
State V1 scope and what's out of scope.
State the Constitution principles (6 principles).
Check stop rules.
If blocked, refuse and explain. If clear, proceed."
```

Example session opening:

```
Current phase: PERFECT — hardening, quality gates, documentation.
I can: run audits, fix edge cases, improve error handling, set up CI,
       sync SPEC-as-built.md, run review checklist.
I cannot: add new features, change architecture, add dependencies.
V1 scope: All 8 CLI commands, SM-2 + AI gen, selective interleaving,
          streak/forgiveness/adherence features are in scope.
Out of scope: GUI, web server, collaboration, mobile, Anki import.
Constitution: No CGO, no SDKs, SM-2 with Scheduler interface, propagate
              all errors, SQLite WAL with max 1 writer, inward deps.
Stop rules: checked — all clear to proceed.
```

---

## Version

**Current: v1.0.0** — Initial project bootstrap for Oh-My-Learner v2 (July 2026).

Derived from Development Protocol RULES.md v2.2.0 template. Adapted for
Oh-My-Learner: a Go CLI spaced-repetition study tool with AI card generation.

---

## Origin

Generated July 2026 for Oh-My-Learner v2. Based on Development Protocol
specifications from Bus-Hop, Ithmb-Codec, and Oh-My-Learner v1.
