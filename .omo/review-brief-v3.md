# Oh-My-Learner v3 — Review Brief for Independent Reviewer

## Instructions

Open a **NEW** AI session. Load ONLY the files listed below. Do NOT read any prior context or conversation history.

Run the **24-item fixed checklist** from REVIEW.md in the Development Protocol. Produce a findings document with PASS/FAIL per item.

## Files to load

**Project: Oh-My-Learner v3** at `/home/nami/projects/dev/Oh-My-Learner/`
**Protocol: Development Protocol** at `/home/nami/projects/dev/Development-Protocol/`

Specifically load these files for the review:

1. `REVIEW.md` — the checklist
2. `RULES.md` — to verify phase/route compliance
3. The full Oh-My-Learner project directory

## What the project claims

Oh-My-Learner v3 is a Go CLI study tool with:

- SM-2 (default) and FSRS-5 (opt-in) spaced repetition schedulers
- AI card generation via DeepSeek (`agent/` package)
- 6 subject packs covering NTU CS Y2S1 (88 templates)
- Self-explanation storage (new in v3)
- 49 passing tests
- Selective interleaving (procedural interleaved, declarative blocked)
- Backlog forgiveness and streak tracking
- Shell/tmux hooks for adherence

## Reviewer rules

1. **No prior context** — you have never seen this project before
2. **Fixed checklist only** — run ALL 24 checks from REVIEW.md verbatim
3. **Output-only** — read and report, do NOT edit any files
4. **Blind to intent** — compare spec vs code vs explainer. Do not accept "but the intent was..."
5. **Findings format** — Summary, FAIL items (CRITICAL/MAJOR/MINOR), PASS items, Reviewer Declaration

## Output

Save findings to `/home/nami/projects/dev/Oh-My-Learner/review-findings-v3.md`
