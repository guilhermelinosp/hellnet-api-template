# hellnet-api-template

> Opinionated, production-ready GitHub template for Go HTTP APIs.
> Built entirely on [hellnet-lib-api](https://github.com/guilhermelinosp/hellnet-lib-api) — you only write business logic.

Three non-negotiable statements about this codebase:

```text
Gin is an implementation detail (hidden inside hellnet-lib-api).
hellnet-lib-api provides config, routing, middleware, server and platform probes.
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
| Env-first config (`HELLNET_*` + `APP_*` fallback, `.env` in dev) | [hellnet-lib-api/config](https://github.com/guilhermelinosp/hellnet-lib-api) |
| HTTP routing, adapters, validation, error envelope | [hellnet-lib-api/api](https://github.com/guilhermelinosp/hellnet-lib-api) |
| Structured logging (`slog`, JSON, trace-correlated) | [hellnet-lib-telemetry](https://github.com/guilhermelinosp/hellnet-lib-telemetry) |
| Distributed tracing + metrics (`/metrics` Prometheus) | hellnet-lib-telemetry |
| `/live` `/ready` `/health` platform probes | hellnet-lib-api/platform (via telemetry) |
| Graceful shutdown with correct telemetry flush order | hellnet-lib-api/platform |
| Secure timeouts, request-id, security headers, CORS | hellnet-lib-api/adapter |
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
#    keeping the same shape: Handler + Service using hellnet-lib-api/api contracts.

# 2. Run:
go run ./cmd/api/

# 3. Verify:
curl -s localhost:8080/live
curl -s localhost:8080/ready
curl -s localhost:8080/health
curl -s localhost:8080/metrics
curl -s 'localhost:8080/api/v1/hello?name=you'
```

## Configuration

| Variable | Purpose | Default |
|---|---|---|
| `HELLNET_APP_NAME` / `APP_NAME` | service name (also telemetry fallback) | `hellnet-api-template` |
| `HELLNET_APP_PORT` / `APP_PORT` | listen port | `8080` |
| `HELLNET_TELEMETRY_ENDPOINT` / `HELLNET_ENDPOINT` | OTLP collector URL (no-op without it) | *empty* |
| `HELLNET_TELEMETRY_SERVICE` / `HELLNET_SERVICE` | telemetry service name | app name |
| `HELLNET_ENVIRONMENT` / `APP_ENV` | `development` / `production` | `development` |

## Architecture

```
cmd/api/main.go       ← wiring only (4 steps: ctx → telemetry → platform → routes)
internal/hello/        ← your domain (handler + service, transport-agnostic)
```

**`cmd/api/main.go` is intentionally tiny:**

1. Create application context (signal-aware)
2. Boot telemetry (`HELLNET_TELEMETRY_*` envs — runs in no-op mode without `ENDPOINT`)
3. Create fully-wired HTTP app (`platform.New(tel)`: config + gin + middleware + server)
4. Mount business routes + serve until SIGINT/SIGTERM

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

The workflows call reusable workflows from [templates](https://github.com/guilhermelinosp/templates), pinned by commit SHA. Releases need the `HELLNET_ACTIONS_PRIVATE_KEY` secret and the `HELLNET_ACTIONS_CLIENT_ID` variable (set them with `scripts/setup-repo.sh`).

## Versioning

Releases follow [Conventional Commits]. Hellnet libraries stay on **v1**;
this template ships as a normal application (`v1.x.x`).

## Contributing and license

See [CONTRIBUTING.md](CONTRIBUTING.md) and [SECURITY.md](SECURITY.md). Licensed under [Apache 2.0](LICENSE).

[Conventional Commits]: https://www.conventionalcommits.org/
