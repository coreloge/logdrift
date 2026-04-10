# logdrift

A lightweight CLI tool for tailing and diffing structured log streams across multiple services simultaneously.

---

## Installation

```bash
go install github.com/yourusername/logdrift@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/logdrift.git && cd logdrift && go build -o logdrift .
```

---

## Usage

Tail logs from multiple services and highlight structural differences in real time:

```bash
logdrift tail --services auth-service,payment-service --format json
```

Diff two log streams to surface field-level divergence:

```bash
logdrift diff --left auth-service --right user-service --field level,timestamp,error
```

Pipe existing log files for offline comparison:

```bash
logdrift diff --left ./logs/svc-a.log --right ./logs/svc-b.log --format json
```

### Flags

| Flag | Description |
|------|-------------|
| `--services` | Comma-separated list of services to tail |
| `--format` | Log format: `json`, `logfmt` (default: `json`) |
| `--field` | Fields to include in diff output |
| `--follow` | Keep streaming until interrupted (default: `true`) |
| `--since` | Show logs since duration e.g. `5m`, `1h` |

---

## Requirements

- Go 1.21+

---

## License

MIT © 2024 yourusername