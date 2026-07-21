<div align="center">

<h1>Phanes DNA</h1>

<p><strong>Phanes DNA — Resident Software Architect & Context Governance Platform for AI Coding Agents.</strong></p>

<p>
<a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License: MIT"></a>
<img src="https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white" alt="Go 1.25+">
<img src="https://img.shields.io/badge/MCP-stdio-purple" alt="MCP Server">
<img src="https://img.shields.io/badge/platform-macOS%20%7C%20Linux%20%7C%20Windows-lightgrey" alt="Platform">
</p>

</div>

---

## What It Does

**Phanes DNA** is a zero-dependency, resident-software-architect platform that discovers, indexes, and compresses architectural rules and AST syntax from multi-stack repositories, serving them seamlessly to AI coding assistants via the Model Context Protocol (MCP).

**The Problem**: AI coding agents lose architectural context across sessions, violating layer boundaries (e.g., calling raw DB queries inside a UI controller or bypassing domain services).

**The Solution**: Phanes DNA acts as an in-process architect. It parses multi-stack codebases, enforces layer rules via SQLite vector storage, compresses context with **Caveman Filtering** (saving ~40–65% of tokens), and gates code via Git Hooks and GitHub Actions.

---

## Supported Stacks & Key Features

### Supported Language Stacks

| Language | AST Parser Engine | Features |
| --- | --- | --- |
| **Java** | `go-tree-sitter` (Java grammar) | Full AST tree, Spring Boot annotations, Layer classification |
| **Go** | Native `go/parser` | Structs, interfaces, functions, methods, zero-dependency |
| **Python** | RegEx & AST Scanner | Classes, functions, layer heuristics (`.py`) |
| **TypeScript / JS** | RegEx & AST Scanner | ES6 classes, functions, routes (`.ts`, `.tsx`, `.js`, `.jsx`) |

---

## Core Capabilities

1. **MCP Stdio Server (`phanes-dna serve`)**:
   - Exposes `get_project_dna` and `review_architecture` tools natively to Claude Code, Cursor, Antigravity, OpenCode, and Codex.
2. **Caveman Token Compression Engine**:
   - Strips courtesies, hedging, connectors, and unneeded articles from architectural context before sending to LLMs, reducing prompt token costs by up to 65%.
3. **Local Cache & Sync Engine (`.dna` bundles)**:
   - Export project rules and architecture state into an ultralight, gzip-compressed `.dna` bundle (~150 bytes) to check into Git so the entire team shares identical architectural boundaries.
4. **Git Hooks Integration**:
   - Installs pre-commit & pre-push hooks (`phanes-dna hooks install`) to block architectural violations in milliseconds before code touches the repo.
5. **GitHub Action CI Bot**:
   - Executes `phanes-dna review --ci` during Pull Requests to annotate layer violations as native GitHub Action inline errors.
6. **One-Command Setup (`phanes-dna setup`)**:
   - Automatically registers `phanes-dna` stdio server in Antigravity, OpenCode, Claude Desktop, and Cursor configuration files.

---

## Quick Start

### Install (Recommended)

**macOS / Linux**

```bash
curl -fsSL https://raw.githubusercontent.com/arley/phanes-dna/main/scripts/install.sh | bash
```

**Windows (PowerShell)**

```powershell
irm https://raw.githubusercontent.com/arley/phanes-dna/main/scripts/install.ps1 | iex
```

### Build & Install from Source

Requires **Go 1.25+**:

```bash
git clone https://github.com/arley/phanes-dna.git
cd phanes-dna
go build -o phanes-dna ./cmd/phanes-dna
```

---

## Interactive Terminal UI

Simply run `phanes-dna` without arguments to open the interactive CLI menu:

```bash
phanes-dna
```

---

## CLI Command Reference

| Command | Description |
| --- | --- |
| `phanes-dna` | Launch interactive terminal UI menu |
| `phanes-dna analyze [path]` | Scan and index multi-stack source files into local SQLite store |
| `phanes-dna review [--strict] [--ci]` | Audit architecture compliance & layer rules |
| `phanes-dna serve` | Launch stdio MCP server for AI coding agents |
| `phanes-dna export [out.dna]` | Export project rules to an ultralight `.dna` bundle |
| `phanes-dna import <file.dna>` | Import `.dna` sync bundle into local store |
| `phanes-dna setup [agent]` | Auto-configure MCP stdio entry for AI assistants (`claude`, `cursor`, `antigravity`, `opencode`, `all`) |
| `phanes-dna hooks install` | Install Git pre-commit & pre-push hooks |
| `phanes-dna doctor` | Execute environment health & ecosystem diagnostics |
| `phanes-dna version` | Display binary version and active AI provider |

---

## MCP Integration Setup

Run the automatic installer:

```bash
phanes-dna setup
```

Or manually configure your agent's MCP JSON file:

```json
{
  "mcpServers": {
    "phanes-dna": {
      "command": "phanes-dna",
      "args": ["serve"]
    }
  }
}
```

---

## License

Distributed under the [MIT License](LICENSE).
