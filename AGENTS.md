# AGENTS.md — Phanes DNA Agentic Architecture & Developer Guide

Welcome, AI Coding Agent! This document explains the architecture, design principles, and available tooling for **Phanes DNA**.

---

## 🧬 Project Overview

**Phanes DNA** is a zero-dependency, resident-software-architect platform that discovers, indexes, and compresses architectural rules and AST syntax from multi-stack repositories, serving them seamlessly to AI coding assistants via the Model Context Protocol (MCP).

- **Language**: Go 1.25+ (pure Go, zero-CGO SQLite via `modernc.org/sqlite`).
- **Protocol**: MCP (Model Context Protocol stdio server using `mark3labs/mcp-go`).
- **Core Strategy**: Multi-stack AST parsing, Caveman token reduction, vector similarity search, Git hooks gating, GitHub Actions CI bot, and Dev Onboarding Mentor.

---

## 🏛️ Codebase Structure

```
├── cmd/phanes-dna/          # Main CLI entrypoint (analyze, review, onboard, commit, serve, setup, hooks, doctor)
├── internal/
│   ├── ai/                  # AI Provider abstraction (Ollama, Gemini, Anthropic, auto-detection)
│   ├── analyzer/            # Multi-stack AST parsers (Java tree-sitter, Go native AST, Python/TS generic)
│   ├── dna/                 # Core domain models, layer classification & Caveman token compressor
│   ├── doctor/              # Ecosystem health checks & diagnostics
│   ├── gitflow/             # Conventional commit generator based on branch & staged diff
│   ├── githooks/            # Pre-commit & pre-push POSIX Git hook installers
│   ├── mcp/                 # MCP stdio server handlers (get_project_dna, review_architecture, dev_onboarding)
│   ├── onboard/             # Dev Onboarding Agent mentor
│   ├── setup/               # Auto-configurator for Claude, Cursor, Antigravity, OpenCode
│   ├── store/               # SQLite database & vector similarity search engine
│   ├── sync/                # Gzip-compressed .dna bundle import/export engine
│   └── tui/                 # Interactive CLI terminal UI menu
├── scripts/                 # Single-command installers (install.sh, install.ps1)
├── README.md                # English documentation & Quick Start
└── README.es.md             # Spanish documentation & Quick Start
```

---

## 🛠️ MCP Stdio Tools Exposed to Agents

When `phanes-dna serve` is active, the following tools are available via stdio:

1. **`get_project_dna(query: string)`**:
   - Performs vector search across indexed code chunks and layer rules, returning Caveman-compressed context.
2. **`review_architecture()`**:
   - Audits project layer rules and returns a list of architectural violations.
3. **`dev_onboarding(topic: string)`**:
   - Generates a step-by-step developer onboarding guide based on codebase patterns and active layer conventions.

---

## 🧪 Verification & Development Commands

Always run full test verification before proposing changes:

```bash
# Run all unit tests across packages
go test -v ./...

# Build binary
go build -o phanes-dna ./cmd/phanes-dna

# Run health diagnostics
./phanes-dna doctor
```

---

## 📜 Commit Policy

Follow Conventional Commits format (`<type>(<scope>): <short description>`) in the language requested by the team or configured via `PHANES_LANGUAGE` / `--lang`.
