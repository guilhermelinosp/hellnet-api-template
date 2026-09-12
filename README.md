# golang-api-template

> Opinionated, production-ready GitHub template for Go HTTP APIs.
> Built entirely on [hellnet-lib-api](https://github.com/guilhermelinosp/hellnet-lib-api) — you only write business logic.

[![pipeline](https://github.com/guilhermelinosp/golang-api-template/actions/workflows/pipeline.yml/badge.svg)](https://github.com/guilhermelinosp/golang-api-template/actions/workflows/pipeline.yml)
[![pr-check](https://github.com/guilhermelinosp/golang-api-template/actions/workflows/pr-check.yml/badge.svg)](https://github.com/guilhermelinosp/golang-api-template/actions/workflows/pr-check.yml)
[![CodeQL](https://github.com/guilhermelinosp/golang-api-template/actions/workflows/codeql.yml/badge.svg)](https://github.com/guilhermelinosp/golang-api-template/actions/workflows/codeql.yml)

Three non-negotiable statements about this codebase:

```text
Gin is an implementation detail (hidden inside hellnet-lib-api).
hellnet-lib-api provides config, routing, middleware, server and platform probes.
hellnet-lib-telemetry is the standard observability layer.
```

---

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

---

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

---

## Quick start

```bash
# 1. Create your repo from this template, then:
git clone https://github.com/<you>/my-project && cd my-project
go mod edit -module github.com/<you>/my-project
go mod tidy

# 2. Rename the domain (optional): replace internal/hello with your own module.
#    Keep the same shape: Handler + Service using hellnet-lib-api/api contracts.

# 3. Run:
go run ./cmd/api/

# 4. Verify:
curl -s localhost:8080/live
curl -s localhost:8080/ready
curl -s localhost:8080/health
curl -s localhost:8080/metrics
curl -s 'localhost:8080/api/v1/hello?name=you'
```

### Environment variables

| Variable | Purpose | Default |
|---|---|---|
| `HELLNET_APP_NAME` / `APP_NAME` | service name (also telemetry fallback) | `golang-api-template` |
| `HELLNET_APP_PORT` / `APP_PORT` | listen port | `8080` |
| `HELLNET_TELEMETRY_ENDPOINT` / `HELLNET_ENDPOINT` | OTLP collector URL (no-op without it) | *empty* |
| `HELLNET_TELEMETRY_SERVICE` / `HELLNET_SERVICE` | telemetry service name | app name |
| `HELLNET_ENVIRONMENT` / `APP_ENV` | `development` / `production` | `development` |

---

## Develop

```bash
go test ./...                    # all tests (unit only)
go test -race ./...              # race detector
go vet ./...                     # static analysis
golangci-lint run                # linter
```

Install git hooks once:

```bash
lefthook install
```

---

## Versioning

Releases follow [Conventional Commits]. Hellnet libraries stay on **v1**;
this template ships as a normal application (`v1.x.x`).

## License

[Apache 2.0](LICENSE)

[Conventional Commits]: https://www.conventionalcommits.org/
# auto-pr selftest
