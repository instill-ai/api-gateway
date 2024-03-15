package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// pluginName is the plugin name
var pluginName = "registry"

// HandlerRegisterer is the symbol the plugin loader will try to load. It must implement the Registerer interface
var HandlerRegisterer = registerer(pluginName)

type registerer string

func (r registerer) RegisterHandlers(f func(
	name string,
	handler func(context.Context, map[string]interface{}, http.Handler) (http.Handler, error),
)) {
	f(string(r), r.registerHandlers)
}

func (r registerer) registerHandlers(_ context.Context, extra map[string]interface{}, h http.Handler) (http.Handler, error) {

	config, ok := extra[pluginName].(map[string]interface{})
	if !ok {
		return h, errors.New("configuration not found")
	}

	hostport, _ := config["hostport"].(string)
	prefix, _ := config["prefix"].(string)

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		// If the URL path starts with "/v2/" (indicating the first handshake request to confirm registry V2 API),
		// "/v2/prefix/" (before the registry prefix is applied), or "/prefix/v2/" (after the registry prefix is applied),
		// it means that the request is intended for the Instill Artifact registry. In this case, the traffic is hijacked
		// and directly relayed to the registry. Otherwise, if the URL path does not match any of these patterns,
		// the traffic is passed through to the next handler.
		if req.URL.Path != "/v2/" &&
			!strings.HasPrefix(req.URL.Path, fmt.Sprintf("/v2%s", prefix)) &&
			!strings.HasPrefix(req.URL.Path, fmt.Sprintf("%sv2/", prefix)) {
			h.ServeHTTP(w, req)
			return
		}

		req.URL.Scheme = "http"
		req.URL.Host = hostport
		req.URL.Path = strings.TrimSuffix(prefix, "/") + strings.Replace(req.URL.Path, prefix, "/", 1)
		req.RequestURI = ""

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			logger.Error(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Copy headers, status codes, and body from the backend to the response writer
		for k, hs := range resp.Header {
			for _, h := range hs {
				w.Header().Add(k, h)
			}
		}
		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
		resp.Body.Close()

	}), nil
}

func main() {}

// This logger is replaced by the RegisterLogger method to load the one from KrakenD
var logger Logger = noopLogger{}

func (registerer) RegisterLogger(v interface{}) {
	l, ok := v.(Logger)
	if !ok {
		return
	}
	logger = l
	logger.Info(fmt.Sprintf("[PLUGIN: %s] Logger loaded", HandlerRegisterer))
}

// Logger is an interface for logging functionality.
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
