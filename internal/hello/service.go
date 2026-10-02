// Package hello is the reference business module of this template.
//
// It exists to demonstrate the wiring contract — Handler → Service →
// (future) Repository — and how errors travel to the HTTP layer. Replace it
// with your actual domain; keep the same shape.
//
// The service depends only on the stdlib and the platform error type; all
// HTTP details stay in the handler.
package hello

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/guilhermelinosp/fast-platform/platform"
)

// maxNameLength bounds user input early (security default).
const maxNameLength = 64

// Service is the business port implemented by the domain layer. Handlers
// depend on this interface — never on concrete implementations — keeping
// them unit-testable with fakes.
type Service interface {
	Greet(ctx context.Context, name string) (string, error)
}

// Compile-time guard used by tests/fakes documentation.
var _ Service = (*BasicService)(nil)

// BasicService is the minimal reference implementation.
type BasicService struct {
	logger *slog.Logger
}

// NewService builds the basic greeter.
func NewService(logger *slog.Logger) *BasicService {
	if logger == nil {
		logger = slog.Default()
	}
	return &BasicService{logger: logger}
}

// Greet normalizes and validates the incoming name then produces the message.
//
// Normalization/size checks belong here (business rules), while JSON binding
// stays in the handler.
func (s *BasicService) Greet(_ context.Context, name string) (string, error) {
	name = strings.TrimSpace(name)
	if len(name) > maxNameLength {
		return "", platform.ValidationError("name", fmt.Sprintf("must be at most %d characters", maxNameLength))
	}
	s.logger.Debug("greeting generated", slog.String("name", name))
	return fmt.Sprintf("Hello, %s!", name), nil
}
