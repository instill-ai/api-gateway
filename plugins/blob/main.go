package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
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

// HandlerRegisterer is the symbol the plugin loader will try to load. It must implement the Registerer interface
var HandlerRegisterer = blob(pluginName)

// This logger is replaced by the RegisterLogger method to load the one from KrakenD
var logger Logger = noopLogger{}

func (blob) RegisterLogger(v any) {
	l, ok := v.(Logger)
	if !ok {
		return
	}
	logger = l
	Debug("Logger loaded")
}

// Logger is the interface for the logger
type Logger interface {
	Debug(v ...any)
	Info(v ...any)
	Warning(v ...any)
	Error(v ...any)
	Critical(v ...any)
	Fatal(v ...any)
}

// Empty logger implementation
type noopLogger struct{}

func (n noopLogger) Debug(_ ...any)    {}
func (n noopLogger) Info(_ ...any)     {}
func (n noopLogger) Warning(_ ...any)  {}
func (n noopLogger) Error(_ ...any)    {}
func (n noopLogger) Critical(_ ...any) {}
func (n noopLogger) Fatal(_ ...any)    {}

// InfoTemplate is a template for logging
const InfoTemplate = "[PLUGIN: %s] %s "

// Info is a shortcut for logger.Info
func Info(v ...any) {
	logger.Info(fmt.Sprintf(InfoTemplate, HandlerRegisterer, fmt.Sprint(v...)))
}

// Debug is a shortcut for logger.Debug
func Debug(v ...any) {
	logger.Debug(fmt.Sprintf(InfoTemplate, HandlerRegisterer, fmt.Sprint(v...)))
}

// Warning is a shortcut for logger.Warning
func Warning(v ...any) {
	logger.Warning(fmt.Sprintf(InfoTemplate, HandlerRegisterer, fmt.Sprint(v...)))
}

// Error is a shortcut for logger.Error
func Error(v ...any) {
	logger.Error(fmt.Sprintf(InfoTemplate, HandlerRegisterer, fmt.Sprint(v...)))
}

// Critical is a shortcut for logger.Critical
func Critical(v ...any) {
	logger.Critical(fmt.Sprintf(InfoTemplate, HandlerRegisterer, fmt.Sprint(v...)))
}

// Fatal is a shortcut for logger.Fatal
func Fatal(v ...any) {
	logger.Fatal(fmt.Sprintf(InfoTemplate, HandlerRegisterer, fmt.Sprint(v...)))
}
