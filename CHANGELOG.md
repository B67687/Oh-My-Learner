# Changelog

All notable changes to this project will be documented in this file.

## 1.0.0 (2026-07-12)

Initial public release. Built by following the Development Protocol's PREP PHASE.

### Added
- SM-2 spaced repetition scheduler with interleaving
- Template-based problem generation (4 types: standard, code-trace, debug-find, explain-why)
- CLI with six commands: add, review, status, map, config
- Subject pack system with TOML-format templates
- Interleaving across subjects in review sessions
- Subject dependency visualization (`learn map`)
- SQLite storage (pure Go, no CGO)
- All 13 tests passing across 3 packages
