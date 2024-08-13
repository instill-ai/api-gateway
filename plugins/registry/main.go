package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/luraproject/lura/v2/logging"
)

var pluginName = "registry"

type registerer string

func (r registerer) RegisterHandlers(f func(
	name string,
	handler func(context.Context, map[string]interface{}, http.Handler) (http.Handler, error),
)) {
	f(string(r), r.registerHandlers)
}

func (r registerer) registerHandlers(ctx context.Context, extra map[string]interface{}, h http.Handler) (http.Handler, error) {
	config, ok := extra[pluginName].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("configuration not found")
	}

	registryHandler, err := newRegistryHandler(config)
	if err != nil {
		return nil, fmt.Errorf("failed to configure handler: %w", err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// If the URL path starts with "/v2/" (exactly /v2/ indicating the
		// first handshake request to confirm registry V2 API), it means that
		// the request is intended for the Instill Artifact registry. In this
		// case, the traffic is hijacked and directly relayed to the registry.
		// Otherwise, if the URL path does not match any of these patterns, the
		// traffic is passed through to the next handler.
		if !strings.HasPrefix(req.URL.Path, "/v2/") {
			h.ServeHTTP(w, req)
			return
		}

		registryHandler.handler(ctx)(w, req)

	}), nil

}

func main() {}

// HandlerRegisterer is the symbol the plugin loader will try to load. It must implement the Registerer interface
var HandlerRegisterer = registerer(pluginName)

// This logger is replaced by the RegisterLogger method to load the one from KrakenD
var logger = logging.NoOp

func (registerer) RegisterLogger(v interface{}) {
	l, ok := v.(logging.BasicLogger)
	if !ok {
		return
	}
	logger = l
	logger.Info(fmt.Sprintf("[PLUGIN: %s] Logger loaded", HandlerRegisterer))
}
