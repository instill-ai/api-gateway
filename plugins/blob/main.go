package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
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

		// TODO: we should be able to implement this to directly rely on KrakenD instead of writing our own proxy.

		// We encode the MinIO presigned URL to a base64 string in the format:
		// schema://host:port/v1alpha/blob-urls/base64_encoded_presigned_url
		// Here in this plugin, we decode the base64 string to the presigned URL
		// and forward the request to MinIO.
		if len(parts) > 2 && parts[2] == "blob-urls" {

			blobURLBytes, err := base64.StdEncoding.DecodeString(parts[3])
			if err != nil {
				return
			}
			blobURL, err := url.Parse(string(blobURLBytes))
			if err != nil {
				return
			}
			blobURL.Scheme = "http"
			req.Header.Set("host", blobURL.Host)
			req.Host = blobURL.Host
			req.URL = blobURL
			req.RequestURI = ""
			req.Header.Del("Authorization")
			resp, err := http.DefaultClient.Do(req)

			if err != nil {
				http.Error(w, "Failed to proxy request", http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()

			// Copy response headers
			for k, hs := range resp.Header {
				if k != "Access-Control-Allow-Origin" {
					for _, h := range hs {
						w.Header().Add(k, h)
					}
				}
			}

			// Set status code
			w.WriteHeader(resp.StatusCode)

			// Stream response body
			_, err = io.Copy(w, resp.Body)
			if err != nil {
				logger.Error(logPrefix, "Failed to stream response body:", err)
			}
			return
		}

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
