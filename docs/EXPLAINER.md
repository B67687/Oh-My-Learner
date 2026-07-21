# EXPLAINER.md — Oh-My-Learner v3 (AI Teaching Agent)

> Produced during POLISH phase, consumed before SPEC SYNC.
> Bridges the gap between "the AI built it" and "you understand what it does and why."
> One explanation per project. Updated as the project evolves.

## For the Reader

You built Oh-My-Learner v1 as a CLI that generates practice problems from templates and schedules them with spaced repetition. v2 turned it into an **AI teaching agent**: the AI generates cards on demand from any CS topic. v3 adds an FSRS-5 scheduler option alongside the default SM-2, stores self-explain responses in the database, and ships 6 curated subject packs with 88 templates.

This document explains how v3 works at every level — from the big picture down to individual files — so you can understand your own project without reading the code.

---

## 1. Macro Architecture

Oh-My-Learner v3 is a CLI that turns university CS topics into a daily study practice. You type `learn add virtual-memory` and the AI generates 5-10 practice questions from 4 different formats (definition, code trace, bug hunt, explanation). It then schedules these questions across time using spaced repetition — SM-2 by default, or FSRS-5 if you opt in — so you see each one right before you'd forget it. Every day you run `learn review` to see what's due, answer the questions, type a self-explain response that gets saved to the database, and the schedule adapts based on whether you got them right or wrong.

The system has 3 layers, each depending only on the one below:

```
cmd/ (CLI commands)  -->  agent/ (AI teaching agent)  -->  core/ (engine)
```

- **core/** is the engine: it stores cards in SQLite, runs the scheduling math (SM-2 by default, or FSRS-5), and renders templates. It has no idea AI exists. SM-2 was built in v1. FSRS-5 and the self-explain response column are new in v3.
- **agent/** is the new AI layer. It takes a topic name, calls DeepSeek V4 Flash (free API) to generate cards in the 4 template formats, validates the AI's output, and saves valid cards into core/ for scheduling.
- **cmd/** is the user interface — the commands you type. It calls agent/ when you add a topic, and calls core/ directly when you review (because review must work offline, no AI needed).

The key design principle: **AI is only used for card generation.** Reviewing, scheduling, and storage are entirely AI-free. You can study offline after cards are generated.

---

## 2. Data Flow Walk

Let's trace what happens when you run `learn add virtual-memory`:

1. **cmd/add.go** receives `"virtual-memory"` as the topic name. It creates an instance of the AI agent and calls `agent.GenerateCards("virtual-memory")`.

2. **agent/agent.go** constructs a prompt for DeepSeek V4 Flash. The prompt tells the AI:
   - You are a CS tutor preparing practice questions for a university student
   - Generate questions about virtual memory (paging, page tables, TLB, page faults, etc.)
   - Output exactly 5-10 cards in JSON format
   - Use the 4 template types: standard Q&A, code trace, debug-find, explain-why
   - Each card must fill in the template variables (question, answer, code snippet, etc.)

3. **agent/agent.go** sends the prompt to the DeepSeek API via HTTPS. It waits for the JSON response. If the API fails (no internet, server down), it returns a clear error: "AI service unavailable. Try again when connected."

4. **agent/agent.go** receives the JSON and parses it into Card structs. It validates each card:
   - Does it have a non-empty question?
   - Is the template type one of the 4 known types?
   - Are all required variables present for that template type?

   Invalid cards are rejected and logged. Valid cards are returned to cmd/add.go.

5. **cmd/add.go** receives the validated cards and calls `storage.InsertCards()` in core/ to save them to the SQLite database. Each card gets an initial scheduling state (SM-2 defaults: interval=0, ease=2.5, repetitions=0, due=now; FSRS-5 defaults are set on first review).

6. The user sees: "Generated 8 cards for virtual-memory. Added to study queue. 3 cards due today."

Next, when you run `learn review` the next day:

7. **cmd/review.go** calls `storage.DueCards()` which queries SQLite for all cards whose next review date is today or past. Cards from ALL topics are mixed together (interleaving).

8. For each due card, it calls `storage.GetCardWithTemplate()` which reads the card data and its template definition. It renders the template by substituting variables (the question text, code snippets, etc.).
9. The CLI shows each card one at a time. You answer, then rate your recall quality (0-5) and type a self-explain response. Both the rating and the response are saved to the reviews table (the `self_explain_response` column was added by migration in v3). The rating feeds into the scheduling algorithm.
10. **core/scheduler.go** updates the card's scheduling parameters using the `Scheduler` interface. The default is SM-2 (adjusts ease factor, interval, repetitions). If you opted into FSRS-5, it uses `FSRSScheduler` instead (adjusts stability, difficulty, and interval). Either way, a perfect recall extends the interval and a failed recall shortens it. The updated card is saved back to SQLite.
    The key insight: steps 7-10 work **completely offline**. No AI needed. This is intentional — you generate cards when you have internet, review them anywhere.

---

## 3. Module Breakdown

### core/ — The Engine (existing from v1)

| Module       | Responsibility                            | Public API                                                                    | Key Types                                                      |
| ------------ | ----------------------------------------- | ----------------------------------------------------------------------------- | -------------------------------------------------------------- |
| core.go      | Defines all data types                    | Types: CardState, Template, SubjectMeta, ReviewQuality                        | CardState (SM-2 + FSRS-5 fields), Template (name, type, body)  |
| scheduler.go | Scheduler interface + SM-2/FSRS-5 impl    | `ReviewCard(card, quality) -> CardState`; `FSRSScheduler` struct              | Scheduler interface, SM2Scheduler, FSRSScheduler, Rating (0-5) |
| storage.go   | SQLite persistence                        | `InsertCard, GetCard, DueCards, UpdateCardState, InsertReview, RemoveSubject` | SQL queries, modernc.org/sqlite driver                         |
| templater.go | Renders template variables into card text | `RenderTemplate(t Template, vars map) -> string`                              | Go text/template with variable substitution                    |

**What core/ hides**: It hides the SQLite schema (what tables exist, what indexes are used), the scheduling math (SM-2 or FSRS-5), and the template rendering internals. Other modules just call functions and get results.

### agent/ — The AI Teaching Agent

| Module            | Responsibility                                                         | Public API                                                    | Key Types                              |
| ----------------- | ---------------------------------------------------------------------- | ------------------------------------------------------------- | -------------------------------------- |
| agent.go          | AI card generation + prompt construction + API call + response parsing | `GenerateCards(topic string) -> []Card`                       | Agent interface, DeepSeekAgent struct  |
| generator.go      | Topic-to-cards pipeline                                                | Constructs prompt, sends to API, parses JSON, validates cards | Prompt template, Card validation logic |
| generator_test.go | Tests with recorded AI responses                                       | Tests parsing of success/partial/empty/garbage responses      | JSON fixtures in testdata/             |

**What agent/ hides**: It hides the exact prompt sent to the AI, the API endpoint and authentication, the JSON parsing logic, and the card validation rules. Other modules just call `GenerateCards("topic")` and get cards back.

### cmd/ — CLI Commands

| Command                 | File       | What it does                                                                         |
| ----------------------- | ---------- | ------------------------------------------------------------------------------------ |
| `learn add <topic>`     | add.go     | Calls agent.GenerateCards, then storage.InsertCards                                  |
| `learn review`          | review.go  | Queries due cards, shows them one at a time, collects ratings, updates schedule      |
| `learn status`          | status.go  | Shows due counts per topic, streak, progress                                         |
| `learn explore <topic>` | explore.go | Browses AI-generated topic map, shows prerequisite chains                            |
| `learn map <topic>`     | map.go     | Visualizes topic dependencies with prerequisite edges                                |
| `learn report`          | report.go  | Shows weekly retention, streak, daily activity; --verbose for self-explain responses |
| `learn hook`            | hook.go    | Installs shell/tmux hooks for review reminders                                       |
| `learn config`          | config.go  | Views/edits settings                                                                 |

### Architecture Diagram

```
+------------------+     +------------------+     +------------------+
|    cmd/ (CLI)   | --> |   agent/ (AI)    | --> |   core/ (Engine) |
|                  |     |                  |     |                  |
|  learn add       |     |  GenerateCards() |     |  SM-2 / FSRS-5   |
|  learn review    |     |  DeepSeek API    |     |  SQLite Storage  |
|  learn status    |     |  JSON parsing    |     |  Template Engine |
|  learn explore   |     |  Card validation |     |                  |
|  learn map       |     |                  |     |                  |
|  learn report    |     |                  |     |                  |
|  learn hook      |     |                  |     |                  |
|  learn config    |     |                  |     |                  |
+------------------+     +------------------+     +------------------+
         |                                              ^
         +----------------------------------------------+
                    (review path: no AI, direct to core)

---

## 4. Key Decisions

### Decision 1: Extend Go, don't rewrite in Python

**What happened:** v1 was already a working Go CLI with 2,000 lines of tested code. Python has a richer AI ecosystem, but rewriting would mean starting from scratch — all the SM-2 math, SQLite storage, and templates would need to be re-tested.
**What we chose:** Keep Go, add the AI agent as a new `agent/` package.
**Tradeoff:** Go is more verbose for constructing AI prompts and parsing JSON responses. But we keep the single-binary distribution (no Python runtime needed), and the review loop stays fast and offline.
**Why it matters to you:** You don't need to install Python or manage dependencies. The tool is still one binary.

### Decision 2: AI native knowledge over web search

**What happened:** We could either use the AI's existing knowledge to generate cards, or have it search the web for each topic. Web search is more accurate (especially for niche topics) but takes longer and costs money per search.
**What we chose:** Use DeepSeek V4 Flash's native knowledge first. Web search is an optional enhancement for later.
**Tradeoff:** For rapidly-changing topics (e.g., new JavaScript frameworks), the AI might produce outdated info. But for core CS subjects (OS, data structures, algorithms, networking), the material has been stable for decades and the AI knows it well.
**Why it matters to you:** Zero cost per card generation. And the AI can generate cards for a topic in under a minute instead of 5-10 minutes with search.

### Decision 3: AI fills existing 4 template types

**What happened:** v1 had 4 template types (standard Q&A, code trace, debug-find, explain-why) that were handwritten in TOML files. We could invent new types for AI generation, or reuse the existing ones.
**What we chose:** The AI fills the same 4 template types dynamically. The template engine and review loop in core/ work exactly as before.
**Tradeoff:** Some topics don't fit all 4 types. For example, a high-level concept might not have a code snippet to trace. The AI skips template types that don't make sense.
**Why it matters to you:** The review experience is identical whether cards were written by hand (v1) or AI-generated (v2). You don't need to learn a new UI.

### Decision 4: Free tier DeepSeek over paid OpenAI/Claude

**What happened:** OpenAI and Claude are more reliable but cost money per API call. DeepSeek V4 Flash has a free tier and comparable quality for structured card generation.
**What we chose:** DeepSeek API as the primary AI backend.
**Tradeoff:** DeepSeek free tier may have rate limits, downtime, or be discontinued. If it goes away, we fall back to running a local model (llama.cpp) which is slower but free and offline.
**Why it matters to you:** Zero operating cost. You can generate unlimited cards without paying API fees.

---

## 5. Quality Guarantees

### Tests

- **core/ (v1 + v3):** ~90% line coverage. SM-2 algorithm has 8 test cases covering perfect recall, failed review, ease factor adjustments, interval growth. FSRS-5 has dedicated tests covering stability/difficulty initialization, grade mapping, and edge cases. Storage tests cover the `self_explain_response` column, streak tracking, and weekly retention queries. Templates have tests for all 4 types.
- **agent/:** Tests use recorded AI responses (JSON files in agent/testdata/) — no real API calls during testing. Covers: successful parsing, partial responses, empty responses, garbage JSON, missing fields. This ensures the parsing and validation logic is robust even when the AI behaves unpredictably.
- **cmd/:** Integration tests run each command with a fake agent that returns canned responses. Tests that `learn add` correctly calls the agent and stores valid cards.
- **49 tests total** across all packages (up from 26 in v2, 38 before FSRS was added).

### Invariants

- **Every card has a non-empty question.** Validated before insertion.
- **Template type is always one of the 4 known types.** Unknown types are rejected.
- **Scheduling intervals grow monotonically** for successfully recalled cards (SM-2: interval_0 < interval_1 < interval_2; FSRS-5: stability increases with each successful recall).
- **No card appears twice in the same review session.** The session builder deduplicates by card ID.
- **Agent never touches core/ internals.** It only produces Card structs via the public API.

### Safety guarantees

- **Error handling is explicit.** No panics. All functions return errors that propagate to the CLI as user-facing messages.
- **Review works offline.** A network outage during `learn add` never blocks `learn review`. Generation and consumption are completely separate paths.
- **AI failures are isolated.** If the AI returns bad data, that card is rejected but valid cards in the same batch are saved. One bad card doesn't ruin the whole topic.
- **Go is a compiled language.** Type errors, nil pointer dereferences, and many other bug classes are caught at compile time, not at runtime.

### Automated checks

- **CI pipeline** (runs on every pull request):
  - `go build ./...` — does the code compile?
  - `go test ./... -count=1` — do all tests pass?
  - `go vet ./...` — does the static analyzer find issues?
- **Pre-commit** (expected): `go fmt` for consistent formatting.

### Limits (honest)

- **agent/ tests use recorded responses.** If the real DeepSeek API changes its response format, tests won't catch it until a manual integration test is run.
- **No integration test for the real AI API.** This is intentional (it would cost money and depend on internet), but it means the first time the real API is called might uncover issues.
- **core/ tests don't cover concurrent access.** SQLite with modernc.org has limited concurrency support. The tool is single-user, so this is acceptable.
- **No end-to-end test** that runs the full `learn add -> learn review -> learn status` flow automatically. This is verified manually.

- **Self-explain responses are now stored** (was a known gap in v2). The `self_explain_response` column on the `reviews` table captures reflective explanations after each review, shown via `learn report --verbose`. Future work could analyze these for metacognitive insight.
- **FSRS-5 uses default parameters** (Jarrett Ye's published w-vector). These have not been calibrated against this dataset. A future tuning pass could optimize them for the card pool.

---

## Mandatory Check

After reading this explanation, can you answer these questions in your own words?

1. **What does this project do, and what are its 3 main pieces?**
   Oh-My-Learner v3 is a CLI study tool where you give a CS topic and the AI generates practice questions, then schedules them with spaced repetition. The 3 layers are: CLI commands (cmd/), AI teaching agent (agent/), and engine (core/).

2. **What happens from start to finish when you run `learn add paging`?**
   cmd/add.go calls agent.GenerateCards which prompts DeepSeek to create 5-10 cards about paging in 4 formats. The response is parsed and validated. Valid cards are stored in SQLite with SM-2 or FSRS-5 scheduling. You see "Generated X cards."

3. **Which module has the most complexity, and what does it hide?**
   agent/generator.hides the exact AI prompt, API calls, JSON parsing, and card validation rules. Everything outside agent/ just calls GenerateCards("topic") and gets cards back.

4. **What was the hardest design decision, and why?**
   Whether to rewrite in Python (better AI tooling) or extend Go (keep existing code). We chose Go extension to keep the single-binary distribution and offline review capability.

5. **What would break first if something went wrong, and how would you know?**
   If DeepSeek API goes down, `learn add` would fail with "AI service unavailable." But `learn review` would still work because it uses cached cards from SQLite without any AI calls. If something breaks in review, you'd see a storage error message with details.

---

## Pipeline Integration

This EXPLAINER.md was produced during the Oh-My-Learner v3 SPECIFICATION phase. It describes the _planned_ architecture. After the v3 features were built (FSRS-5, self-explain storage, subject packs), this document was updated during POLISH to match what was actually built.

The full pipeline for v3:

```

RAW INTENT -> AMBITION -> SPECIFICATION -> EXPLAINER (this doc)
-> V3 EXECUTOR -> POLISH -> EXPLAINER update -> SPEC SYNC -> REVIEW -> SHIP

```

```
