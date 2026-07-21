# Phanes DNA — Architecture & Progress

> **Plataforma de Contexto Inteligente y Gobernanza para Agentes de Desarrollo**
> _Resident Software Architect — descubre, indexa y comprime reglas arquitectónicas para asistentes AI vía MCP._

---

## Table of Contents

1. [Project Status](#1-project-status)
2. [Tech Stack](#2-tech-stack)
3. [Architecture Overview](#3-architecture-overview)
4. [Package Map](#4-package-map)
5. [SQLite Schema](#5-sqlite-schema)
6. [Key Design Decisions](#6-key-design-decisions)
7. [Data Flow](#7-data-flow)
8. [Implemented Code](#8-implemented-code)
9. [Roadmap](#9-roadmap)
10. [SDD Artifacts](#10-sdd-artifacts)

---

## 1. Project Status

| Area | Status |
|------|--------|
| **Planning** | ✅ Complete (Proposal → Specs → Design → Tasks) |
| **PR #1 — Foundation** | ✅ **Done** — go.mod, types, Provider interface, Caveman |
| **PR #2 — Store + Analyzer** | ✅ **Done** — SQLite schema, vector search, Tree-sitter AST |
| **PR #3 — AI + MCP + CLI** | ✅ **Done** — Ollama, Gemini, Anthropic, MCP server & CLI |
| **PR #4 — Tests + RED** | ✅ **Done** — Analyzer, MCP, Benchmarks & Security RED tests |
| **Tests passing** | `go vet ./...` ✅ — `go build ./...` ✅ — `go test ./...` ✅ |

**PR strategy**: Stacked-to-main (4 PRs, each merges to main)

---

## 2. Tech Stack

| Layer | Technology | Rationale |
|-------|-----------|-----------|
| **Language** | Go 1.25.5 | Zero-dependency binary, native concurrency |
| **SQLite Driver** | `modernc.org/sqlite` (pure Go) | Zero CGO — fully static binary |
| **Vector Search** | In-memory brute-force cosine | Avoids CGO from sqlite-vec; fine for <50K chunks |
| **AST Parser** | `github.com/smacker/go-tree-sitter` + Java grammar | Native Go bindings, no subprocess |
| **MCP Server** | `github.com/mark3labs/mcp-go` | stdio transport (MVP) / SSE (post-MVP) |
| **AI Providers** | Ollama (local) + Gemini / Anthropic (cloud) | Hybrid — local offline, cloud when available |
| **Compression** | Caveman (custom regex pipeline) | ~40-65% token reduction |

---

## 3. Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                       phanes-dna binary                          │
│                                                                 │
│  ┌──────────────┐   ┌──────────────────┐   ┌─────────────────┐  │
│  │    CLI        │   │   MCP Server      │   │   Providers     │  │
│  │  (cobra)      │   │   (mcp-go stdio)  │   │  Ollama/Gemini  │  │
│  │               │   │                   │   │  /Anthropic     │  │
│  │ analyze       │   │ get_project_dna   │   │ Analyze()       │  │
│  │ serve         │   │ review_arch       │   │ Embed()         │  │
│  │ review        │   │                   │   │                 │  │
│  │ export/import │   │                   │   │                 │  │
│  │ setup         │   │                   │   │                 │  │
│  └───────┬───────┘   └────────┬──────────┘   └────────┬────────┘  │
│          │                    │                        │          │
│          ▼                    ▼                        ▼          │
│  ┌─────────────────────────────────────────────────────────────┐  │
│  │                    Core Engine                               │  │
│  │                                                              │  │
│  │  ┌────────────┐  ┌───────────┐  ┌───────────────────────┐   │  │
│  │  │  Analyzer   │  │  Store     │  │  Caveman              │   │  │
│  │  │ (Tree-sitter│  │ (SQLite +  │  │  Compression          │   │  │
│  │  │  AST walk)  │  │  Vector)   │  │  (4-stage filter)     │   │  │
│  │  └────────────┘  └───────────┘  └───────────────────────┘   │  │
│  └─────────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

---

## 4. Package Map

```
github.com/arley/phanes-dna/
│
├── cmd/phanes-dna/              # CLI entry point (TODO — PR #3)
│   └── main.go
│
├── internal/
│   ├── dna/                     # Core domain model + Caveman
│   │   ├── model.go             # Project, SourceFile, ASTNode, Dependency, etc.
│   │   ├── compress.go          # Caveman 4-stage compression pipeline
│   │   └── compress_test.go     # Tests & benchmarks
│   │
│   ├── store/                   # SQLite persistence + vector search
│   │   ├── sqlite.go            # CRUD: projects, source_files, ast_nodes, etc.
│   │   ├── vector.go            # Brute-force cosine similarity search
│   │   └── migrations.go        # Versioned schema migrations
│   │
│   ├── analyzer/                # Multi-Stack AST analyzers
│   │   ├── analyzer.go          # Analyzer interface & Registry
│   │   ├── java.go              # Tree-sitter Java AST walker
│   │   ├── golang.go            # Native Go AST parser (go/parser)
│   │   └── generic.go           # Python, TypeScript, and JavaScript parser
│   │
│   ├── sync/                    # Local Cache & Sync Engine
│   │   ├── bundle.go            # Export/Import gzip compressed .dna bundles
│   │   └── bundle_test.go       # Sync engine unit tests
│   │
│   ├── ai/                      # Provider abstraction & auto-detection
│   │   ├── provider.go          # Provider interface
│   │   ├── factory.go           # Auto-detection cascade
│   │   ├── ollama.go            # Ollama provider
│   │   ├── gemini.go            # Gemini provider
│   │   └── anthropic.go         # Anthropic provider
│   │
│   └── mcp/                     # MCP stdio server
│       ├── server.go            # MCP server setup
│       └── tools.go             # get_project_dna & review_architecture tools
│
├── go.mod
└── go.sum
```

---

## 5. SQLite Schema

7 tables created by migration 001:

```
schema_migrations(version, applied_at)

projects(id PK, name, root_path, analyzed_at)
source_files(id PK, project_id FK, rel_path, content_hash, package_path, layer)
ast_nodes(id PK, file_id FK, parent_id FK, node_type, name, annotations_json, start_line)
dependencies(id PK, src_node_id FK, tgt_node_id FK, dep_type)
layer_rules(id PK, src_layer, tgt_layer, allowed, severity)
chunks(id PK, file_id FK, chunk_idx, content, embedding BLOB)
violations(id PK, file, line, from_class, to_class, severity, rule_ref, recommendation)
```

Indexes on: `source_files(project_id)`, `ast_nodes(file_id, parent_id)`, `dependencies(src, tgt)`, `chunks(file_id)`, `violations(file)`.

---

## 6. Key Design Decisions

### D1 — Zero CGO, fully static binary
- **Decision**: Use `modernc.org/sqlite` (pure Go) instead of `mattn/go-sqlite3` (CGO)
- **Impact**: Binary is fully static, no libsqlite3 dependency, trivial distribution
- **Tradeoff**: Slightly slower SQLite performance, but negligible for MVP scale

### D2 — Brute-force vector search over indexed approach
- **Decision**: Store embeddings as BLOBs in SQLite, brute-force cosine similarity in Go
- **Impact**: Simple architecture, no extra dependencies
- **Tradeoff**: O(n×d) per query — acceptable for <50K chunks. Upgrade path: `cgo_vec` build tag swaps in sqlite-vec

### D3 — Tree-sitter via Go bindings (no subprocess)
- **Decision**: Use `smacker/go-tree-sitter` directly in-process
- **Impact**: No subprocess management, no serialization overhead
- **Tradeoff**: Links the native tree-sitter library at compile time

### D4 — MCP stdio transport for MVP
- **Decision**: Start with stdio transport; SSE deferred to post-MVP
- **Impact**: Simplest integration with code assistants (Claude Code, Cline, etc.)
- **Tradeoff**: No persistent daemon mode yet

### D5 — Provider auto-detection at startup
- **Decision**: Probe Ollama first (localhost:11434), fall through to Gemini/Anthropic keys
- **Impact**: Zero-config for local users, transparent fallback
- **Tradeoff**: Adds ~500ms startup latency for health checks

### D6 — Caveman position in pipeline
- **Decision**: Between vector retrieval and LLM context assembly
- **Impact**: Reduces token costs by 40-65% before injection
- **Tradeoff**: Global config for MVP (per-project post-MVP)

---

## 7. Data Flow

### Analysis flow

```
analyze <path>
    │
    ▼
Tree-sitter parses .java files
    │
    ├─► Class/annotation extraction ──► layer classification (controller/service/repository)
    ├─► Import extraction ────────────► dependency graph edges
    └─► Method/field extraction ──────► AST node tree
    │
    ▼
Store in SQLite:
    projects ──► source_files ──► ast_nodes ──────────► dependencies
                              └─► chunks ──► embedding via AI provider
                                                        │
                                                        ▼
                                              layer_rules inferred
```

### Query flow (MCP)

```
AI Assistant calls get_project_dna("How is DI wired?")
    │
    ▼
Embed query via active provider
    │
    ▼
Cosine similarity search (top-K chunks)
    │
    ▼
Caveman compression (compress retrieved context)
    │
    ▼
Return structured context to AI assistant
```

### Review flow

```
AI Assistant calls review_architecture(diff / paths)
    │
    ▼
Load dependencies + layer_rules from SQLite
    │
    ▼
Walk import graph ──► flag violations
    │
    ▼
Return violations[] with severity + recommendation
```

---

## 8. Implemented Code

### Files created (PR #1 + PR #2)

| File | Lines | Purpose |
|------|-------|---------|
| `go.mod` | 19 | Module def + deps (tree-sitter, modernc/sqlite, mcp-go, Google AI) |
| `internal/dna/model.go` | 60 | Core domain types (7 types) |
| `internal/dna/compress.go` | 93 | Caveman 4-stage compression |
| `internal/dna/compress_test.go` | ~80 | 6 tests: I/O pairs, conditionals, edge cases |
| `internal/ai/provider.go` | 32 | Provider interface + request/response types |
| `internal/store/sqlite.go` | 381 | Full CRUD for all 7 tables + helper funcs |
| `internal/store/vector.go` | 83 | Brute-force cosine similarity search |
| `internal/store/migrations.go` | 135 | Migration engine + 001 schema (7 tables, 7 indexes) |
| `internal/store/sqlite_test.go` | 118 | 3 tests: project round-trip, listing, files |
| `internal/analyzer/java.go` | 294 | Tree-sitter Java AST walker + layer classifier |
| **Total** | **~1295** | |

### Test coverage

```
go test ./...  →  PASS
  internal/dna/compress_test.go   — 6 tests
  internal/store/sqlite_test.go   — 3 tests
```

---

## 9. Roadmap

### PR #1 — Foundation ✅
- [x] `go.mod` + dependencies
- [x] Domain model types (`internal/dna/model.go`)
- [x] Provider interface (`internal/ai/provider.go`)
- [x] Caveman compression (`internal/dna/compress.go`)

### PR #2 — Store & Analyzer ✅
- [x] SQLite persistence: schema, migrations, CRUD (`internal/store/`)
- [x] Vector similarity search (`internal/store/vector.go`)
- [x] Tree-sitter Java AST analyzer (`internal/analyzer/java.go`)

### PR #3 — AI, MCP & CLI ✅
- [x] Ollama provider impl (`internal/ai/ollama.go`)
- [x] Gemini provider impl (`internal/ai/gemini.go`)
- [x] Anthropic provider impl (`internal/ai/anthropic.go`)
- [x] MCP stdio server (`internal/mcp/server.go`)
- [x] MCP tool handlers (`internal/mcp/tools.go`)
- [x] CLI entry point — analyze / serve / review (`cmd/phanes-dna/main.go`)

### PR #4 — Tests & RED ✅
- [x] Java analyzer golden file tests (`internal/analyzer/java_test.go`)
- [x] Caveman ≥40% reduction benchmark (`internal/dna/compress_bench_test.go`)
- [x] MCP handler JSON-RPC tests (`internal/mcp/tools_test.go`)
- [x] RED: comment sanitization & path traversal security tests (`internal/analyzer/security_test.go`)

### Features v0.3.0 ✅
- [x] Multi-Stack AST Analyzers (Java, Go, Python, TypeScript/JS)
- [x] Local Cache & Sync Engine (`.dna` gzip bundle export/import)
- [x] Git Hooks Integration (`internal/githooks/` — pre-commit / pre-push)
- [x] GitHub Action / CI Bot (`.github/workflows/phanes-dna-review.yml` + `--ci` format)
- [x] Automated MCP Setup Installer (`internal/setup/`)

### Post-MVP Roadmap
- [ ] Web GUI (Next.js dashboard)
- [ ] Session Interception (Copilot/ChatGPT Web tokens)
- [ ] Per-project Caveman config
- [ ] SSE MCP transport (daemon mode)

---

## 10. SDD Artifacts

All planning artifacts stored in Engram (`topic_key` format):

| Artifact | Topic Key | Status |
|----------|-----------|--------|
| Project init | `sdd-init/Codigo-Facilito` | ✅ |
| Proposal | `sdd/phanes-dna/proposal` | ✅ |
| Specs | `sdd/phanes-dna/spec` | ✅ |
| Design | `sdd/phanes-dna/design` | ✅ |
| Tasks | `sdd/phanes-dna/tasks` | ✅ |
| Apply progress | `sdd/phanes-dna/apply-progress` | ✅ (PR #1 + #2) |

---

> **Phanes DNA** — El arquitecto residente que tu equipo de IA necesita.
>
> _"El mejor código es el que nunca se escribió porque el contexto ya estaba claro."_
