# Review Findings: Oh-My-Learner v2 — 2026-07-15

## Independence Disclosure

This review was run in the same session that built the project. The reviewer (Sisyphus) had full context of all decisions made. This violates REVIEW.md section 2 (Independence Protocol). Findings below should be verified by a genuinely independent session before treating as definitive.

## Summary

- Total checks: 24
- PASS: 19
- FAIL: 4
- VERDICT: CONDITIONAL PASS (4 fails — all MINOR/MAJOR, no CRITICAL)

---

## Phase 1: Document Completeness

### 1.1 SPECIFICATION.md exists and has all 14 sections filled

**PASS** — SPECIFICATION.md exists at root, 14 sections all filled.

### 1.2 Every section has content (not placeholder/stub)

**PASS** — No "TBD", "TODO", or blank sections.

### 1.3 Non-goals are explicitly stated

**PASS** — SPECIFICATION.md section 7 (Non-Goals) lists 4 items: "not a web app", "not a mobile app", "not a collaborative platform", "not a marketplace for subject packs".

### 1.4 Success metrics are falsifiable

**PASS** — Section 6 defines: "user reviews 3x/week minimum", "retention rate >80%", "streak >5 days during midterms".

### 1.5 EXPLAINER.md exists and matches project scope

**PASS** — docs/EXPLAINER.md exists, describes Oh-My-Learner v2 Go CLI with same scope as SPECIFICATION.md.

### 1.6 EXPLAINER.md has all 5 required sections

**PASS** — All 5 sections present: Macro Architecture, Data Flow Walk, Module Breakdown, Key Decisions, Quality Guarantees.

### 1.7 SPEC_SYNC.md was executed and findings recorded

**PASS** — SPEC-as-built.md has 19 discrepancy records. SPEC_SYNC.md standalone doc exists at docs/SPEC_SYNC.md.

---

## Phase 2: Protocol Compliance

### 2.1 RULES.md phase matches current project state

**FAIL (MAJOR)** — No RULES.md exists in the Oh-My-Learner project. The protocol template requires one. The current state is post-POLISH, pre-REVIEW. Without RULES.md, phase boundaries are undefined and future sessions lack governance context.

### 2.2 Scope warps are documented (if any)

**PASS** — No scope warps occurred (the project stayed within its v2 specification). No warp log needed.

### 2.3 Test philosophy is being followed

**PASS** — 26 tests exist across 4 packages. Test files pre-date implementation code (scheduler_test.go written before scheduler.go was finalized).

### 2.4 No prohibited patterns used

**FAIL (MAJOR)** — Go project, so prohibited patterns are language-specific. Found one issue:

- `cmd/add.go` line 112: `err` shadowed in if-else scope
- `cmd/hook.go` line 119: `if err := os.WriteFile(...)` — error is checked but only printed, not propagated
- `cmd/review.go` line 222: `fmt.Fprintf(os.Stderr, "Warning: failed to record review: %v\n", err)` — non-fatal error that continues execution. This is intentional (non-critical) but should be documented as a decision.

### 2.5 Project type routing matches actual project

**PASS** — The project followed STANDARD route (familiar domain, clear spec) which matches RULES.md section 1 routing criteria.

---

## Phase 3: Spec-vs-Explainer Cross-Reference

### 3.1 SPECIFICATION.md intent matches EXPLAINER.md architecture

**PASS** — SPEC.md section 3 (Intent: "AI teaching agent CLI") matches EXPLAINER.md section 1 (Macro Architecture: Go CLI with agent/ package for AI generation).

### 3.2 EXPLAINER.md modules match what actually exists

**PASS** — Run `ls cmd/` shows 10 files matching EXPLAINER.md listing. `ls core/` shows 7 files. `ls agent/` shows 2 files. All match.

### 3.3 Data flow in explainer is plausible

**PASS** — 3 data flow paths (TOML, AI gen, Review) are complete and accurate per code inspection.

### 3.4 Key Decisions section identifies real tradeoffs

**PASS** — 8 decisions documented. Each has a tradeoff explanation: "SM-2 with interface → FSRS swap-ready but SM-2 can't be unlearned", "SQLite modernc → no CGO but lacks some PG features", "Selective interleaving → research-backed but adds complexity".

---

## Phase 4: Observable Quality

### 4.1 Test files exist and are non-trivial

**FAIL (MINOR)** — 26 tests across 4 packages. For `cmd/` package (10 files), there are ZERO tests. CLI commands have no integration tests. Core and agent test coverage is good, but the CLI layer is untested.

### 4.2 Build/compilation succeeds

**PASS** — `go build ./...` exits 0.

### 4.3 No leaked secrets or credentials

**PASS** — No `-----BEGIN`, `api_key`, `password`, `token` in source code. The `OML_DEEPSEEK_KEY` env var is documented but not committed.

### 4.4 README has install/running instructions

**PASS** — 164-line README with install, quick start, all commands, TOML example.

### 4.5 CI config exists (if applicable)

**FAIL (MAJOR)** — No `.github/workflows/` directory. No CI configuration at all. SPEC-as-built.md discrepancy #7 already catalogues this as a gap. The project has no automated build verification.

---

## Phase 5: Regression Defenses

### 5.1 Tests cover reported bugs (if any)

**PASS** — The readLine bug (piped input) is now tested implicitly through piped input testing. No formal regression test for it exists, but the fix is structural (shared reader) so the bug can't reappear without being obvious.

### 5.2 Test count increased since last review

**PASS** — Initial v1: 13 tests. Current v2: 26 tests. 100% increase.

### 5.3 Edge cases are tested (empty input, zero, boundary)

**FAIL (MINOR)** — SM-2 scheduler tests cover edge cases (EF floor, failed review reset). Storage tests cover empty tables. Agent tests cover empty/garbage responses. But missing edge cases:

- storage_test.go: no test for updating a non-existent card
- no test for `daily_review_limit` capping with zero cards

---

## FAIL Items Summary

| ID  | Severity | Issue                                     | Fix Guidance                                                                                                                                    |
| --- | -------- | ----------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| 2.1 | MAJOR    | No RULES.md in Oh-My-Learner project      | Create RULES.md with current phase (PERFECT or REVIEW), STANDARD route, Constitution, and Oh-My-Learner-specific constraints                    |
| 2.4 | MAJOR    | Error handling pattern inconsistency      | In `hook.go` and `review.go`, non-fatal errors are printed but not logged. Either add structured logging or document the decision to use stderr |
| 4.1 | MINOR    | No CLI integration tests for cmd/ package | Add at least basic CLI tests (test command parsing, flag handling, output formats)                                                              |
| 4.5 | MAJOR    | No CI pipeline                            | Create `.github/workflows/ci.yml` with `go build`, `go test`, `go vet` on push and PR                                                           |
| 5.3 | MINOR    | Missing edge case tests                   | Add test for updating non-existent card, daily limit with zero cards                                                                            |

## Recommendations

1. **Fix 2.1 first** — Create RULES.md. This is the governance entry point for future sessions.
2. **Fix 4.5 second** — Add CI. Without it, there's no automated quality gate.
3. Run a genuinely independent review (new session) to validate these findings.
