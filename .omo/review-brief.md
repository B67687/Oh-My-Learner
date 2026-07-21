# Review Brief: Oh-My-Learner v2

**Instructions:** Open a NEW OpenCode session (not this one). Load ONLY this file + the project files mentioned below. Run REVIEW.md from the Development Protocol. Do NOT read any other context.

## What to load

1. **Oh-My-Learner project** at `/home/nami/projects/dev/Oh-My-Learner/`
2. **Development Protocol** — copy these files to the review session:
   - `REVIEW.md` (the checklist)
   - `RULES.md` (to check phase/route compliance)
3. **Read this review brief** (`.omo/review-brief.md`) — that's all the context you get.

## Project for review

Oh-My-Learner v2 — a Go CLI study tool with AI card generation.

- Built by: Orchestrator (Sisyphus) via OMO task agents
- Spec: `SPECIFICATION.md` (14 sections, 552 lines)
- Code: `SPEC-as-built.md` catalogues 19 discrepancies vs spec
- Explainer: `docs/EXPLAINER.md` (234 lines)
- Tests: 26 passing

## Reviewer rules (from REVIEW.md)

1. **No prior context** — you have NOT seen this project before. That's intentional.
2. **Fixed checklist only** — run the 24 items in REVIEW.md verbatim
3. **Output-only** — read and report. Do NOT edit any file.
4. **Blind to intent** — compare spec vs code vs explainer. Do not accept "but the intent was..."
5. **Produce findings** — CRITICAL/MAJOR/MINOR per severity levels in REVIEW.md

## Output format

Save findings to `review-findings.md` in the project root. Follow the REVIEW.md output template: Summary, FAIL items, PASS items, Warnings, Reviewer Declaration.
