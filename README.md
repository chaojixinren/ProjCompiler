# ProjCompiler

[![Go Version](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-GPL%20v3-blue.svg)](LICENSE)

[中文说明](README.zh-CN.md)

## What is ProjCompiler

ProjCompiler is a Go CLI tool that scans local repositories, synthesizes compact project specifications, and compiles implementation prompts. It helps developers quickly generate structured project descriptions for collaboration with AI models.

## Core Features

- **Smart Repository Scanning** - Automatically identifies project structure, build manifests, docs, entrypoints, and representative code snippets
- **Spec Synthesis** - Derives compact project specs (goals, tech profile, modules, constraints) from raw facts
- **Prompt Compilation** - Applies `/init` pruning rules, keeping only information models might get wrong
- **Confidence Annotation** - Each fact carries a `FactSignal` (Observed/Inferred/Missing) for traceability
- **Interactive TUI** - Bubble Tea-based terminal interface with bilingual switch (Ctrl+L)
- **Graceful Degradation** - Falls back to deterministic synthesis when LLM is unavailable

## Architecture Overview

ProjCompiler uses a four-stage pipeline:

```
Scan → SpecBuilder → PromptCompiler → Exporter
```

| Data Structure | Source | Description |
|---------------|--------|-------------|
| `RepoFacts` | Scanner | Raw facts (file tree, docs, build info, entrypoints, snippets) |
| `ProjectSpec` | SpecBuilder | Compact spec (goals, tech profile, modules, constraints) |
| `PromptBundle` | PromptCompiler | Compiled prompt text |

Project layout:

```
cmd/projcompiler/     # Entry point
internal/
  app/                # App wiring, contracts
  config/             # Config loading
  scan/               # Repository scanning (orchestrator + 5 probe tools)
  spec/               # Shared data structures
  prompt/             # Prompt compilation
  adkflow/            # LLM spec synthesis
  tui/                # Bubble Tea terminal UI
```

## Screenshots

### TUI Interface

![TUI English](public/tui_EN.png)

### Configuration

![Config English](public/config_EN.png)

## Quick Start

### Build

```bash
go build -o projcompiler ./cmd/projcompiler
```

### Configure

Configuration can be set via `.env` file or directly in the TUI configuration page.

Copy `.env.example` to `.env`:

```bash
cp .env.example .env
```

Configuration options:

| Key | Description | Required |
|-----|-------------|----------|
| `PROJCOMPILER_BASEURL` | Model endpoint | Yes |
| `PROJCOMPILER_APIKEY` | API key | Yes |
| `PROJCOMPILER_MODEL` | Model name | Yes |
| `PROJCOMPILER_OUTPUT_LANGUAGE` | Output language | No |

### Run

```bash
./projcompiler
```

Enter a project path after launch. The app runs scan → spec synthesis → prompt compilation, then exports `prompt.txt`.

## Usage Examples

### Pack your project through a prompt

```bash
./projcompiler
> Enter path: /path/to/your/project
> Output: prompt.txt
```

Submit the generated `prompt.txt` to an AI model for implementation suggestions tailored to your project context.

## Contributing

Issues and Pull Requests are welcome.

1. Fork this repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Create a Pull Request

## License

[GPL-3.0](LICENSE)