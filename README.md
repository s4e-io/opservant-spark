# Opservant Spark

<p align="center">
  <img src=".github/assets/name_and_logo.png" alt="Opservant Spark" width="460"/>
</p>

<p align="center">
  <a href="https://github.com/s4e-io/opservant-spark/releases"><img src="https://img.shields.io/github/v/release/s4e-io/opservant-spark?label=release&color=blue" alt="Release"></a>
  <img src="https://img.shields.io/badge/go-%3E%3D1.26-00ADD8?logo=go&logoColor=white" alt="Go version">
  <a href="LICENSE"><img src="https://img.shields.io/github/license/s4e-io/opservant-spark" alt="License"></a>
  <a href="https://github.com/s4e-io/opservant-spark/actions/workflows/ci.yml"><img src="https://github.com/s4e-io/opservant-spark/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
</p>

<p align="center">
  <b>Open source security agent that runs infrastructure playbooks anywhere.</b><br>
  A single binary. No runtime dependencies. Linux, macOS, and Windows.
</p>

<p align="center">
  <a href="#quick-start">Quick Start</a> •
  <a href="#how-it-works">How It Works</a> •
  <a href="#cacao-aligned-execution">CACAO Framework</a> •
  <a href="#playbook-format">Playbook Format</a> •
  <a href="#configuration">Configuration</a> •
  <a href="#contributing">Contributing</a>
</p>

---

## What is Spark?

Spark is a lightweight CLI agent that loads JSON security playbooks from disk and executes them on the target machine. It handles platform detection, privilege checks, human approval gates, variable injection protection, timeouts, and automatic rollback — so playbook authors can focus on *what* to do, not *how* to do it safely.

```
 ┌─────────────┐      ┌─────────────┐      ┌──────────────┐
 │  Playbook    │ ───▶ │  Processor  │ ───▶ │   Executor   │
 │  (JSON)      │      │  Pipeline   │      │  (CACAO)   │
 └─────────────┘      └─────────────┘      └──────────────┘
                       Ingest → Parse       Per-action safety
                       Validate → Enrich    checks & execution
```

Use it standalone or pair it with [**opservant-playbooks**](https://github.com/s4e-io/opservant-playbooks) — a community-maintained library of 100+ ready-to-use security playbooks.

## Quick Start

### Install from source

```bash
git clone https://github.com/s4e-io/opservant-spark.git
cd opservant-spark
make spark
```

Or without Make:

```bash
go build -o spark ./cmd/opservant-spark
```

### Configure

```bash
cp config.example.yaml config.yaml
```

```yaml
agent:
  name: "my-spark"
  uuid: "550e8400-e29b-41d4-a716-446655440000"  # any valid UUID

logging:
  level: "info"       # trace | debug | info | warn | error
  log_to_file: false
  log_dir: "./logs"
```

Generate a UUID with `uuidgen` (macOS/Linux) or `[guid]::NewGuid()` (PowerShell).

### Run

```bash
# Single playbook
./spark playbook --config config.yaml ssh-hardening.json

# All playbooks in a directory
./spark playbook --config config.yaml --dir ./playbooks
```

## How It Works

Spark processes every playbook through a **two-stage pipeline**:

### Stage 1: Processor Pipeline

The processor transforms a raw JSON file into a validated, enriched playbook:

| Step | What it does |
|------|-------------|
| **Ingest** | Reads the file, enforces a 10 MB size limit |
| **Parse** | Decodes JSON into the internal playbook model |
| **Validate** | Checks required fields, slug rules, risk levels, OS values |
| **Enrich** | Inherits playbook-level variables into tasks and actions |

### Stage 2: Execution Engine

The executor runs tasks sequentially, respecting dependency ordering (`depends_on`), with per-task and per-playbook timeouts. Each action passes through Spark's **CACAO-aligned** safety pipeline before the command is executed.

## CACAO-Aligned Execution

Spark's playbook model is inspired by the [OASIS CACAO](https://www.oasis-open.org/standard/cacao-security-playbooks-v2-0/) standard — **Collaborative Automated Course of Action Operations** — the industry standard for defining, sharing, and executing security playbooks in a structured, machine-readable format.

Every action passes through a multi-gate safety pipeline:

```
  ╭──────────────────────────────────────────────╮
  │          CACAO Execution Pipeline             │
  │                                               │
  │  1. Platform Gate    OS compatibility check   │
  │  2. Approval Gate    Human-in-the-loop        │
  │  3. Privilege Gate   Admin/root validation    │
  │  4. Assembly         Variable resolution      │
  │                      + injection protection   │
  │  5. Execution        Timeout-bound command    │
  │  6. Capture          Stdout/stderr + logging  │
  │  7. Rollback         Auto-revert on failure   │
  ╰──────────────────────────────────────────────╯
```

**Platform Gate** — Checks `supported_os` against the current platform. Skips gracefully if the action is not meant for this OS.

**Approval Gate** — If `approval_required: true`, Spark pauses and asks for interactive confirmation before proceeding. This maps to CACAO's human-in-the-loop decision points.

**Privilege Gate** — If `requires_admin: true`, verifies the agent is running with root/administrator privileges.

**Assembly** — Resolves `${var}` and `{{var}}` placeholders from playbook variables. Blocks shell injection sequences (`;`, `&&`, `` ` ``, `$(`, etc.) on both Unix and Windows.

**Execution** — Runs the command with enforced action-level, task-level, and playbook-level deadlines. The shortest remaining deadline always wins.

**Capture** — Captures command stdout/stderr with structured logging, categories, and execution summaries.

**Rollback** — When `auto_revert_on_failure` is enabled and a task fails, Spark executes `rollback_command` on completed actions in reverse order — implementing CACAO's course-of-action reversal pattern.

## Cross-Platform Build

| Command | Output | Platform |
|---------|--------|----------|
| `make spark` | `spark` | Current OS/arch |
| `make spark-linux-amd64` | `spark-linux-amd64` | Linux x86_64 |
| `make spark-linux-arm64` | `spark-linux-arm64` | Linux ARM64 |
| `make spark-darwin-amd64` | `spark-darwin-amd64` | macOS Intel |
| `make spark-darwin-arm64` | `spark-darwin-arm64` | macOS Apple Silicon |
| `make spark-windows-amd64` | `spark-windows-amd64.exe` | Windows x86_64 |
| `make spark-windows-arm64` | `spark-windows-arm64.exe` | Windows ARM64 |
| `make spark-all` | All of the above | All platforms |
| `make test` | — | Run all tests |
| `make vet` | — | Run Go vet |
| `make clean` | — | Remove built binaries |

Override the version tag:

```bash
make spark-all VERSION=1.2.0
```

## Playbook Format

Playbooks are self-contained JSON files. Each defines tasks and actions to execute:

```json
{
  "slug": "check-open-ports",
  "name": "Check Open Ports",
  "description": "Lists all listening ports on the host.",
  "risk_level": "low",
  "risk_score": 2,
  "timeout_seconds": 60,
  "supported_os": ["linux", "macos"],
  "target_tags": ["network", "ports"],
  "tasks": [
    {
      "slug": "scan-ports",
      "name": "Scan listening ports",
      "actions": [
        {
          "slug": "netstat-scan",
          "name": "Run netstat",
          "command": "netstat -tuln"
        }
      ]
    }
  ]
}
```

For 100+ ready-to-use playbooks covering hardening, information gathering, configuration, maintenance, and reboot across Linux, macOS, and Windows — see [**opservant-playbooks**](https://github.com/s4e-io/opservant-playbooks).

## Configuration

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `agent.name` | string | — | Human-readable name for this instance |
| `agent.uuid` | string | — | Unique identifier (valid UUID required) |
| `logging.level` | string | `info` | `trace` · `debug` · `info` · `warn` · `error` |
| `logging.log_to_file` | bool | `false` | Write logs to files in `log_dir` |
| `logging.log_dir` | string | `./logs` | Directory for `spark.log` and `execution.log` |

## Project Structure

```
opservant-spark/
├── cmd/opservant-spark/    # CLI entry point and commands
├── internal/
│   ├── agent/              # CACAO-aligned execution engine
│   ├── config/             # YAML configuration (Viper)
│   ├── initializer/        # Ordered boot sequence with cleanup
│   ├── logger/             # Structured logging with categories
│   ├── models/             # Playbook, Task, Action data models
│   ├── processor/          # Ingest → Parse → Validate → Enrich
│   └── system/             # OS-level privilege checks
├── examples/               # Sample playbooks for demo and testing
├── config.example.yaml
├── Makefile
└── go.mod
```

## Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for details.

```bash
# Development setup
git clone https://github.com/s4e-io/opservant-spark.git
cd opservant-spark
go mod download

# Run tests
go test ./...

# Run linter
go vet ./...
```

> **Note:** Playbook contributions belong in [opservant-playbooks](https://github.com/s4e-io/opservant-playbooks), not this repository.

## Getting Help

- [GitHub Issues](https://github.com/s4e-io/opservant-spark/issues) — bug reports and feature requests
- [Opservant Playbooks](https://github.com/s4e-io/opservant-playbooks) — ready-to-use security playbook library
- [Opservant](https://opservant.org/) — the full Opservant ecosystem

## License

Opservant Spark is licensed under the [GNU Affero General Public License v3.0](LICENSE).

---

<p align="center">
  Built by <a href="https://s4e.io">s4e.io</a>
</p>
