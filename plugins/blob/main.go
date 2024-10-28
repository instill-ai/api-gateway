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
	handler func(context.Context, map[string]interface{}, http.Handler) (http.Handler, error),
)) {
	f(string(r), r.registerHandlers)
}

func (r blob) registerHandlers(ctx context.Context, extra map[string]interface{}, h http.Handler) (http.Handler, error) {
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

func (blob) RegisterLogger(v interface{}) {
	l, ok := v.(Logger)
	if !ok {
		return
	}
	logger = l
	Debug("Logger loaded")
}

type Logger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Warning(v ...interface{})
	Error(v ...interface{})
	Critical(v ...interface{})
	Fatal(v ...interface{})
}

// Empty logger implementation
type noopLogger struct{}

func (n noopLogger) Debug(_ ...interface{})    {}
func (n noopLogger) Info(_ ...interface{})     {}
func (n noopLogger) Warning(_ ...interface{})  {}
func (n noopLogger) Error(_ ...interface{})    {}
func (n noopLogger) Critical(_ ...interface{}) {}
func (n noopLogger) Fatal(_ ...interface{})    {}

// InfoTemplate is a template for logging
const InfoTemplate = "[PLUGIN: %s] %s "

// Info is a shortcut for logger.Info
func Info(v ...interface{}) {
	logger.Info(fmt.Sprintf(InfoTemplate, HandlerRegisterer, fmt.Sprint(v...)))
}

// Debug is a shortcut for logger.Debug
func Debug(v ...interface{}) {
	logger.Debug(fmt.Sprintf(InfoTemplate, HandlerRegisterer, fmt.Sprint(v...)))
}

// Warning is a shortcut for logger.Warning
func Warning(v ...interface{}) {
	logger.Warning(fmt.Sprintf(InfoTemplate, HandlerRegisterer, fmt.Sprint(v...)))
}

// Error is a shortcut for logger.Error
func Error(v ...interface{}) {
	logger.Error(fmt.Sprintf(InfoTemplate, HandlerRegisterer, fmt.Sprint(v...)))
}

// Critical is a shortcut for logger.Critical
func Critical(v ...interface{}) {
	logger.Critical(fmt.Sprintf(InfoTemplate, HandlerRegisterer, fmt.Sprint(v...)))
}

// Fatal is a shortcut for logger.Fatal
func Fatal(v ...interface{}) {
	logger.Fatal(fmt.Sprintf(InfoTemplate, HandlerRegisterer, fmt.Sprint(v...)))
}
