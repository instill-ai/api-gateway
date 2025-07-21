package main

import (
	"context"
	"fmt"
	"html"
	"net/http"
	"strings"

	"github.com/luraproject/lura/v2/logging"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var pluginName = "registry"

type registerer string

func (r registerer) RegisterHandlers(f func(
	name string,
	handler func(context.Context, map[string]any, http.Handler) (http.Handler, error),
)) {
	f(string(r), r.registerHandlers)
}

func (r registerer) registerHandlers(ctx context.Context, extra map[string]any, h http.Handler) (http.Handler, error) {
	config, ok := extra[pluginName].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("configuration not found")
	}

	registryHandler, err := newRegistryHandler(config)
	if err != nil {
		return nil, fmt.Errorf("failed to configure handler: %w", err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Extract OpenTelemetry context from incoming request
		otelCtx := req.Context()

		// Create a span for the registry plugin
		tracer := trace.SpanFromContext(otelCtx).TracerProvider().Tracer("registry")
		spanCtx, span := tracer.Start(otelCtx, "registry.handle_request",
			trace.WithAttributes(
				attribute.String("http.method", req.Method),
				attribute.String("http.url", req.URL.String()),
				attribute.String("http.user_agent", req.UserAgent()),
				attribute.String("plugin.name", pluginName),
			),
		)
		defer span.End()

		// Add span context to request
		req = req.WithContext(spanCtx)

		// If the URL path starts with "/v2/" (exactly /v2/ indicating the
		// first handshake request to confirm registry V2 API), it means that
		// the request is intended for the Instill Artifact registry. In this
		// case, the traffic is hijacked and directly relayed to the registry.
		// Otherwise, if the URL path does not match any of these patterns, the
		// traffic is passed through to the next handler.
		if !strings.HasPrefix(req.URL.Path, "/v2/") {
			span.SetAttributes(attribute.String("registry.skip_reason", "non_v2_path"))
			h.ServeHTTP(w, req)
			return
		}

		span.SetAttributes(attribute.String("registry.action", "process_v2_request"))
		registryHandler.handler(spanCtx)(w, req)

	}), nil
}

func main() {}

// HandlerRegisterer is the symbol the plugin loader will try to load. It must
// implement the Registerer interface.
var HandlerRegisterer = registerer(pluginName)

// This logger is replaced by the RegisterLogger method to load the one from
// KrakenD.
var logger logging.Logger = logging.NoOp

func (registerer) RegisterLogger(v any) {
	l, ok := v.(logging.Logger)
	if !ok {
		return
	}
	logger = l
	logger.Info(logPrefix, "Logger loaded")
}

// The following functions are shortcuts for formatted logging.
var logPrefix = fmt.Sprintf("[PLUGIN: %s]", HandlerRegisterer)

func sanitize(s string) string {
	// html.EscapeString doesn't fully prevent log injection.
	sanitized := strings.ReplaceAll(s, "\n", "")
	sanitized = strings.ReplaceAll(sanitized, "\r", "")
	sanitized = strings.ReplaceAll(sanitized, "\t", "")
	sanitized = strings.ReplaceAll(sanitized, "\x00", "")

	return html.EscapeString(sanitized)
}

func logReq(req *http.Request, v []any) []any {
	logFields := make([]any, 3, len(v)+3)
	logFields[0] = logPrefix
	logFields[1] = req.Method
	logFields[2] = sanitize(req.URL.Path)

	return append(logFields, v...)
}

func logDebug(req *http.Request, v ...any)    { logger.Debug(logReq(req, v)...) }
func logInfo(req *http.Request, v ...any)     { logger.Info(logReq(req, v)...) }
func logWarning(req *http.Request, v ...any)  { logger.Warning(logReq(req, v)...) }
func logError(req *http.Request, v ...any)    { logger.Error(logReq(req, v)...) }
func logCritical(req *http.Request, v ...any) { logger.Critical(logReq(req, v)...) }
func logFatal(req *http.Request, v ...any)    { logger.Fatal(logReq(req, v)...) }
