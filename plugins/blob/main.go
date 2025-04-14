package main

import (
	"context"
	"fmt"
	"html"
	"net/http"
	"strings"

	"github.com/luraproject/lura/v2/logging"
)

var pluginName = "blob"

type blob string

func (r blob) RegisterHandlers(f func(
	name string,
	handler func(context.Context, map[string]any, http.Handler) (http.Handler, error),
)) {
	f(string(r), r.registerHandlers)
}

func (r blob) registerHandlers(ctx context.Context, extra map[string]any, h http.Handler) (http.Handler, error) {
	config, ok := extra[pluginName].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("configuration not found")
	}

	blobHandler, err := newBlobHandler(config)
	if err != nil {
		return nil, fmt.Errorf("failed to configure handler: %w", err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		parts := strings.Split(req.URL.Path, "/")
		if len(parts) < 6 || parts[1] != "v1alpha" || parts[2] != "namespaces" || parts[4] != "blob-urls" {
			h.ServeHTTP(w, req)
			return
		}

		blobHandler.handler(ctx)(w, req)

	}), nil

}

func main() {}

// HandlerRegisterer is the symbol the plugin loader will try to load. It must
// implement the Registerer interface.
var HandlerRegisterer = blob(pluginName)

// This logger will be replaced by the RegisterLogger method to load the one from
// KrakenD.
var logger logging.Logger = logging.NoOp

func (blob) RegisterLogger(v any) {
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
