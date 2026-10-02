# hellnet-api-template

> Opinionated, production-ready GitHub template for Go HTTP APIs.
> Built on [fast-platform](https://github.com/guilhermelinosp/fast-platform) (`platform` and `env`) and
> [hellnet-lib-telemetry](https://github.com/guilhermelinosp/hellnet-lib-telemetry) — you only write business logic.

Three non-negotiable statements about this codebase:

```text
fast-platform/platform provides config, router, middleware, server and the error envelope (Gin).
fast-platform/env loads the dev .env and reads typed environment variables.
hellnet-lib-telemetry is the standard observability layer.
```

---

[![pipeline](https://github.com/guilhermelinosp/hellnet-api-template/actions/workflows/pipeline.yml/badge.svg)](https://github.com/guilhermelinosp/hellnet-api-template/actions/workflows/pipeline.yml)
[![pr-check](https://github.com/guilhermelinosp/hellnet-api-template/actions/workflows/pr-check.yml/badge.svg)](https://github.com/guilhermelinosp/hellnet-api-template/actions/workflows/pr-check.yml)
[![CodeQL](https://github.com/guilhermelinosp/hellnet-api-template/actions/workflows/codeql.yml/badge.svg)](https://github.com/guilhermelinosp/hellnet-api-template/actions/workflows/codeql.yml)

## What you get for free

All infrastructure comes from two libraries — nothing is reimplemented here:

| Capability | Source |
|---|---|
| Env-first config (`HELLNET_*`, `.env` in dev) | [fast-platform/platform](https://github.com/guilhermelinosp/fast-platform) |
| HTTP router (Gin), validation, error envelope | fast-platform/platform |
| Structured logging (`slog`, JSON, trace-correlated) | [hellnet-lib-telemetry](https://github.com/guilhermelinosp/hellnet-lib-telemetry) |
| Distributed tracing + metrics (exported over OTLP) | hellnet-lib-telemetry |
| `/live` `/ready` `/health` platform probes | hellnet-lib-telemetry |
| Graceful shutdown with correct telemetry flush order | fast-platform/platform |
| Secure timeouts, request-id, security headers, CORS | fast-platform/platform |
| CI: lint/CodeQL/dependency-review/govulncheck | `.github/workflows` |
| Release: semver tag → GH release → GoReleaser → image | org reusable workflows + GoReleaser |
| Container (distroless, non-root, reproducible) | `Containerfile` |

Your 20% is `internal/hello/` — the reference business module demonstrating how to write domain logic on top of the library contracts.

## Initialize from this template

After **Use this template**, clone the new repository and run:

```bash
scripts/init-from-template.sh <repo-name> [service-name]   # renames the module, imports and cmd/ (services)
scripts/setup-repo.sh                                      # repo settings, "main" ruleset and CI variable
```

Then create the `HELLNET_ACTIONS_PRIVATE_KEY` secret (the script prints the exact command) and make sure the
`hellnet-actions` GitHub App is installed on the repository.

## Quick start

```bash
# 1. Initialise the repository first (see above); optionally replace internal/hello with your own module,
#    keeping the same shape: Handler (gin) + Service.

# 2. Run:
go run ./cmd/api/

# 3. Verify:
curl -s localhost:8080/live
curl -s localhost:8080/ready
curl -s localhost:8080/health
curl -s 'localhost:8080/api/v1/hello?name=you'
```

## Configuration

| Variable | Purpose | Default |
|---|---|---|
| `HELLNET_SERVICE` | service name (also the telemetry service name) | `hellnet-api-template` |
| `HELLNET_PORT` | listen port | `8080` |
| `HELLNET_ENVIRONMENT` | `Development` (gin debug) or any other value (release) | `Development` |
| `HELLNET_TELEMETRY_ENDPOINT` | OTLP collector URL (no-op without it) | *empty* |
| `SHUTDOWN_TIMEOUT`, `READ_TIMEOUT`, `WRITE_TIMEOUT`, `IDLE_TIMEOUT`, `READ_HEADER_TIMEOUT` | server timeouts (Go durations) | `10s`, `15s`, `30s`, `120s`, `10s` |
| `BODY_LIMIT`, `CORS_ALLOWED_ORIGINS`, `TRUSTED_PROXIES` | request body size (bytes), comma-separated origins and proxies | `1048576`, *none*, *none* |

## Architecture

```
cmd/api/main.go       ← wiring only (ctx → config → telemetry → router → routes → run)
internal/hello/        ← your domain (gin handler + service)
```

**`cmd/api/main.go` is intentionally tiny:**

1. Create the process context (`platform.Context()`: signal-aware, loads the dev `.env`)
2. Boot telemetry (`HELLNET_TELEMETRY_*` envs — runs in no-op mode without `ENDPOINT`)
3. Build the router (`platform.NewRouter`: gin + middleware) and mount the probes and business routes
4. Serve until SIGINT/SIGTERM (`platform.Run`: graceful shutdown), then flush telemetry

## Development

```bash
go test -race ./...
go vet ./...
golangci-lint run ./...
```

Install the git hooks once with `lefthook install`: they run formatting, vet, tests (with and without `-race`), build, `go mod tidy`, lint, `govulncheck` and a secrets scan. Commits follow [Conventional Commits](https://www.conventionalcommits.org/).

## CI/CD

| Workflow | Trigger | What it does |
|---|---|---|
| `pr-check` | pull request | shellcheck, merge strategy and Conventional Commits (`merge-check`), Gitleaks, labels and the Go quality gate (module integrity, vet, race tests with coverage, lint, build, dependency review). `pr-gate` aggregates them and is the required check |
| `pipeline` | push to `main` (ignores `.github/**`) or manual | semver guard (blocks an automatic major), immutable tag + GitHub Release, container image |
| `codeql` | nightly or manual | static analysis (CodeQL) |
| `security` | nightly or manual | Gitleaks and Trivy scans |
| `auto-pr` | push to `feat/**` or `fix/**` | opens the pull request automatically |
| `dependabot-actions-auto-merge` | Dependabot pull requests | auto-merges GitHub Actions bumps |

The workflows call reusable workflows from [templates](https://github.com/guilhermelinosp/templates) at `@latest`. Releases need the `HELLNET_ACTIONS_PRIVATE_KEY` secret and the `HELLNET_ACTIONS_CLIENT_ID` variable (set them with `scripts/setup-repo.sh`).

## Versioning

Releases follow [Conventional Commits]. Hellnet libraries stay on **v1**;
this template ships as a normal application (`v1.x.x`).

## Contributing and license

See [CONTRIBUTING.md](CONTRIBUTING.md) and [SECURITY.md](SECURITY.md). Licensed under [Apache 2.0](LICENSE).

[Conventional Commits]: https://www.conventionalcommits.org/
