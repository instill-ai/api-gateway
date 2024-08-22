package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

/*

the following configuration must be present in the krakend.json file:

"extra_config": {
    "plugin/http-server": {
      "name": ["sse-streaming"],
      "sse-streaming": {
        "backend_host": "http://localhost:9081"
      }
    }
  }

the provided endpoint overwrites any existing endpoint in the configuration file.

*/

// pluginName is the name of the plugin, used as a key in the configuration map.
var pluginName = "sse-streaming"

// HandlerRegisterer is the symbol the plugin loader will try to load. It must implement the Registerer interface.
var HandlerRegisterer = registerer(pluginName)

type registerer string

// RegisterHandlers registers the handler function with the given name.
func (r registerer) RegisterHandlers(f func(
	name string,
	handler func(context.Context, map[string]interface{}, http.Handler) (http.Handler, error),
)) {
	f(string(r), r.registerHandlers)
}

// registerHandlers extracts configuration and sets up the HTTP handler.
func (r registerer) registerHandlers(ctx context.Context, extra map[string]interface{}, h http.Handler) (http.Handler, error) {
	config, ok := extra[pluginName].(map[string]interface{})
	if !ok {
		return h, errors.New("configuration not found")
	}

	// Extract configuration values
	backendHost, backendHostOk := config["backend_host"].(string)

	// Check if all required configuration values are present
	if !backendHostOk {
		return h, errors.New("missing required configuration values")
	}

	if backendHost == "" {
		return h, errors.New("backend_host cannot be empty")
	}

	// Validate the backend host
	if _, err := url.ParseRequestURI(backendHost); err != nil {
		return h, errors.New("invalid backend_host URL")
	}

	// Return a new HTTP handler that wraps the original handler with custom logic.
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		httpClient := http.Client{Transport: http.DefaultTransport}

		// This is a quick solution since we only support sse for pipeline trigger endpoint
		if req.Header.Get("instill-use-sse") == "true" {
			proxyHandler(w, req, httpClient, backendHost)
		} else {
			h.ServeHTTP(w, req)
		}

	}), nil
}

// proxyHandler forwards the request to the actual SSE server and streams the response back to the client.
func proxyHandler(w http.ResponseWriter, r *http.Request, httpClient http.Client, backendHost string) {

	url := string(r.URL.Path)
	url = strings.ReplaceAll(url, "/internal", "")
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s%s", backendHost, url), r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header = r.Header
	resp, err := httpClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Create a buffered reader to read the SSE stream
	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		// Write the event to the client
		_, err = w.Write([]byte(line))
		if err != nil {
			break
		}

		// Flush the response writer to ensure the event is sent immediately
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}
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
	logger.Debug(fmt.Sprintf("[PLUGIN: %s] Logger loaded", HandlerRegisterer))
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
