// Package main bootstraps the API.
//
// Deliberately tiny: it wires explicit dependencies in lifecycle order and
// owns the shutdown sequence. Nearly all real decisions live in
// hellnet-lib-api (config, adapter, platform, server) and
// hellnet-lib-telemetry; this file only composes them with the business
// module (internal/hello).
//
//	context → config → telemetry → platform app → routes → run → shutdown
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/guilhermelinosp/hellnet-lib-api/api"
	"github.com/guilhermelinosp/hellnet-lib-api/platform"
	"github.com/guilhermelinosp/hellnet-lib-environments/environments"
	"github.com/guilhermelinosp/hellnet-lib-telemetry/telemetry"

	"github.com/guilhermelinosp/golang-api-template/internal/hello"
)

// Build metadata injected via -ldflags (see .goreleaser.yaml, Containerfile).
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", slog.Any("error", err))
		os.Exit(1)
	}
}

func run() error {
	// 1. Environment: dev convenience — loads .env if present, so the app
	//    boots with zero configuration (idempotent; ignored outside dev).
	_ = environments.LoadDotEnv()

	// 2. Application context — created ONCE here; server + telemetry inherit it
	//    for graceful shutdown on SIGINT/SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 2. Telemetry — env-first, self-contained (HELLNET_TELEMETRY_* with
	//    HELLNET_* fallback + .env in dev). Without HELLNET_TELEMETRY_ENDPOINT
	//    the SDK boots in no-op mode so the template runs with zero config.
	tel, err := telemetry.New()
	if err != nil {
		return err
	}
	logger := tel.Logger
	if logger == nil {
		logger = slog.Default()
	}
	logger.Info("starting",
		slog.String("version", version),
		slog.String("commit", commit),
		slog.String("date", date),
	)

	// 3. Fully wired HTTP application from hellnet-lib-api: environment-driven
	//    config (HELLNET_* with APP_* fallback) + gin adapter + telemetry
	//    middleware + HTTP server with graceful shutdown.
	app, err := platform.New(tel)
	if err != nil {
		return err
	}

	// 4. Business dependencies (composition, no DI framework). The template's
	//    domain module only speaks hellnet-lib-api contracts.
	helloHandler := hello.NewHandler(hello.NewService(logger))

	// 5. Mount platform probes (live/ready/health) + /api/v1 routes.
	app.Register(api.Deps{
		Platform: app.PlatformHandlers(),
		Routes:   helloHandler.Routes(),
	})

	// 6. Serve until ctx is cancelled, then flush telemetry LAST so final
	//    logs/traces/metrics still export.
	if err := app.Run(ctx); err != nil {
		logger.Error("runtime error", slog.Any("error", err))
	}
	logger.Info("shutting down: flushing telemetry")
	return app.Shutdown()
}
