// Package main bootstraps the API.
//
// Deliberately tiny: it wires explicit dependencies in lifecycle order and
// owns the shutdown sequence. The HTTP runtime (config, router, middleware,
// server, error envelope) comes from fast-platform/platform and telemetry from
// hellnet-lib-telemetry; this file only composes them with the business module
// (internal/hello).
//
//	context → config → telemetry → router → routes → run → shutdown
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/guilhermelinosp/fast-platform/platform"
	"github.com/guilhermelinosp/hellnet-lib-telemetry/telemetry"

	"github.com/guilhermelinosp/hellnet-api-template/internal/hello"
)

// defaultService is the service name until HELLNET_SERVICE is set.
const defaultService = "hellnet-api-template"

// Build metadata injected via -ldflags (see .goreleaser.yaml, Containerfile).
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	if err := run(); err != nil {
		platform.Fatal("api", err)
		os.Exit(1)
	}
}

func run() error {
	// 1. Process context: cancelled on SIGINT/SIGTERM (graceful shutdown). It
	//    also loads ./.env in development, so the app boots with zero config.
	ctx, stop, err := platform.Context()
	if err != nil {
		return err
	}
	defer stop()

	// 2. Config (HELLNET_SERVICE, HELLNET_PORT, timeouts, CORS, ...) and
	//    telemetry. HELLNET_SERVICE defaults to the template name and, without
	//    HELLNET_TELEMETRY_ENDPOINT, the SDK boots in no-op mode, so the
	//    template runs with zero configuration.
	if os.Getenv("HELLNET_SERVICE") == "" {
		_ = os.Setenv("HELLNET_SERVICE", defaultService)
	}
	cfg, err := platform.NewConfig()
	if err != nil {
		return err
	}
	ops, err := telemetry.New(ctx)
	if err != nil {
		return err
	}
	// Flush telemetry LAST so the final logs/traces/metrics still export.
	defer func() { _ = ops.Close(context.WithoutCancel(ctx)) }()

	// 3. Business dependencies (composition, no DI framework).
	helloHandler := hello.NewHandler(hello.NewService(slog.Default()))

	// 4. Router: platform probes (live/ready/health) + /api/v1 routes.
	router := platform.NewRouter(cfg, ops)
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": cfg.Name,
			"version": version,
			"commit":  commit,
			"builtAt": date,
			"go":      runtime.Version(),
		})
	})
	router.GET("/live", gin.WrapH(ops.Live()))
	router.GET("/ready", gin.WrapH(ops.Ready()))
	router.GET("/health", gin.WrapH(ops.Health()))
	helloHandler.Register(router.Group("/api/v1"))

	ops.Log(ctx).Info("starting",
		"version", version, "commit", commit, "date", date, "port", cfg.Port)

	// 5. Serve until ctx is cancelled.
	err = platform.Run(ctx, cfg, platform.NewServer(cfg, ops, router))
	if err != nil && !errors.Is(err, context.Canceled) {
		ops.Log(ctx).Error("runtime error", "error", err)
		return err
	}
	return nil
}
