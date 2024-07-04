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
        "endpoint": "/sse/{id}",
        "backend_url_pattern": "/events-stream/{id}",
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
	endpoint, endpointOk := config["endpoint"].(string)
	backendURLPattern, backendURLPatternOk := config["backend_url_pattern"].(string)
	backendHost, backendHostOk := config["backend_host"].(string)

	// Check if all required configuration values are present
	if !endpointOk || !backendURLPatternOk || !backendHostOk {
		return h, errors.New("missing required configuration values")
	}

	// Basic sanity checks on the configuration values
	if endpoint == "" {
		return h, errors.New("endpoint cannot be empty")
	}
	if backendURLPattern == "" {
		return h, errors.New("backend_url_pattern cannot be empty")
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
		// TODO: Performance optimize matchStrings; critical, every request to the API gateway uses this.
		matchPaths, id := matchStrings(endpoint, req.URL.Path)

		if !matchPaths {
			h.ServeHTTP(w, req)
			return
		}

		// Construct serverURL using the extracted ID
		serverURL := fmt.Sprintf("http://%s%s", backendHost, strings.Replace(backendURLPattern, "{id}", id, 1))
		// Call proxyHandler if the path matches
		proxyHandler(w, req, serverURL)
	}), nil
}

// proxyHandler forwards the request to the actual SSE server and streams the response back to the client.
func proxyHandler(w http.ResponseWriter, r *http.Request, serverURL string) {
	logger.Debug("server URL", serverURL)
	// Forward the request to the actual SSE server
	resp, err := http.Get(serverURL)
	if err != nil {
		errM := "failed to connect to downstream SSE server"
		logger.Critical(errM)
		http.Error(w, errM, http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Set headers for the client
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

// matchStrings checks if the request path matches the pattern and extracts the ID.
func matchStrings(pattern, str string) (bool, string) {
	patternParts := strings.Split(pattern, "/")
	strParts := strings.Split(str, "/")

	if len(patternParts) != len(strParts) {
		return false, ""
	}

	var id string
	for i := 0; i < len(patternParts); i++ {
		if patternParts[i] != strParts[i] && patternParts[i] != "{id}" {
			return false, ""
		}
		if patternParts[i] == "{id}" {
			id = strParts[i]
		}
	}

	return true, id
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
