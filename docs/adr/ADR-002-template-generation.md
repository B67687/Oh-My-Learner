# ADR-002: Template-Based Generation (No Hand-Written Cards)

**STATUS:** Accepted

## Context
Traditional flashcard apps require users to write cards, which is the hardest part — so students don't do it. The tool needed to generate practice problems automatically.

## Decision
Template-based generation using parameterized TOML templates + Go's text/template engine. Template authors write question/answer templates with `{{ variable }}` placeholders. The system randomly selects variable values at render time, creating unique problems each session.

## Consequences
**Positive:** Zero card-writing burden for the end user.
**Positive:** Subject packs are simple TOML files — easy to create and share.
**Negative:** Template-based generation cannot produce every possible question type. Complex problem types require custom template logic in code.
