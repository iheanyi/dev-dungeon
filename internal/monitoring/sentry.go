// Package monitoring provides error tracking and observability.
package monitoring

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/charmbracelet/log"
	"github.com/getsentry/sentry-go"
)

// Config holds Sentry configuration.
type Config struct {
	DSN         string
	Environment string
	Release     string
	Debug       bool
}

// Init initializes Sentry error tracking.
// Returns a cleanup function that should be deferred.
func Init(cfg Config) (func(), error) {
	if cfg.DSN == "" {
		log.Info("Sentry DSN not configured, error tracking disabled")
		return func() {}, nil
	}

	// Get version from build info if not provided
	release := cfg.Release
	if release == "" {
		if info, ok := debug.ReadBuildInfo(); ok {
			release = info.Main.Version
		}
		if release == "" || release == "(devel)" {
			release = "development"
		}
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.DSN,
		Environment:      cfg.Environment,
		Release:          release,
		Debug:            cfg.Debug,
		TracesSampleRate: 0.1, // Sample 10% of transactions
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			// Scrub any sensitive data before sending
			return event
		},
	})
	if err != nil {
		return nil, fmt.Errorf("sentry init failed: %w", err)
	}

	log.Info("Sentry error tracking initialized", "environment", cfg.Environment, "release", release)

	// Return cleanup function
	return func() {
		sentry.Flush(2 * time.Second)
	}, nil
}

// CaptureException sends an exception to Sentry with optional context.
func CaptureException(err error, tags map[string]string) {
	if err == nil {
		return
	}

	sentry.WithScope(func(scope *sentry.Scope) {
		for k, v := range tags {
			scope.SetTag(k, v)
		}
		sentry.CaptureException(err)
	})
}

// CaptureMessage sends a message to Sentry.
func CaptureMessage(message string, level sentry.Level, tags map[string]string) {
	sentry.WithScope(func(scope *sentry.Scope) {
		scope.SetLevel(level)
		for k, v := range tags {
			scope.SetTag(k, v)
		}
		sentry.CaptureMessage(message)
	})
}

// SetUser sets the current user context for Sentry events.
func SetUser(userID, username, fingerprint string) {
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{
			ID:       userID,
			Username: username,
			Data: map[string]string{
				"fingerprint": fingerprint,
			},
		})
	})
}

// ClearUser clears the user context.
func ClearUser() {
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetUser(sentry.User{})
	})
}

// RecoverPanic captures a panic and sends it to Sentry.
// Use with defer: defer monitoring.RecoverPanic(ctx, tags)
func RecoverPanic(ctx context.Context, tags map[string]string) {
	if r := recover(); r != nil {
		err, ok := r.(error)
		if !ok {
			err = fmt.Errorf("panic: %v", r)
		}

		sentry.WithScope(func(scope *sentry.Scope) {
			scope.SetLevel(sentry.LevelFatal)
			for k, v := range tags {
				scope.SetTag(k, v)
			}
			scope.SetContext("panic", map[string]interface{}{
				"value": r,
				"stack": string(debug.Stack()),
			})
			sentry.CaptureException(err)
		})

		// Re-panic after capturing
		panic(r)
	}
}

// HTTPRecoveryMiddleware wraps an HTTP handler with panic recovery and Sentry reporting.
func HTTPRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Capture to Sentry
				hub := sentry.GetHubFromContext(r.Context())
				if hub == nil {
					hub = sentry.CurrentHub().Clone()
				}

				hub.WithScope(func(scope *sentry.Scope) {
					scope.SetLevel(sentry.LevelFatal)
					scope.SetRequest(r)
					scope.SetContext("panic", map[string]interface{}{
						"value": err,
						"stack": string(debug.Stack()),
					})

					eventID := hub.RecoverWithContext(r.Context(), err)
					log.Error("HTTP panic recovered", "error", err, "event_id", eventID)
				})

				// Return 500 to client
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// SSHSessionTags returns common tags for SSH session events.
func SSHSessionTags(username, fingerprint, remoteAddr string) map[string]string {
	return map[string]string{
		"session.type":   "ssh",
		"user.username":  username,
		"user.fingerprint": fingerprint,
		"remote.addr":    remoteAddr,
	}
}

// AddBreadcrumb adds a breadcrumb for debugging.
func AddBreadcrumb(category, message string, data map[string]interface{}) {
	sentry.AddBreadcrumb(&sentry.Breadcrumb{
		Category: category,
		Message:  message,
		Data:     data,
		Level:    sentry.LevelInfo,
	})
}
