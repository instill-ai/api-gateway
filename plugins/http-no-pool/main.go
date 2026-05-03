package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/luraproject/lura/v2/logging"
)

func main() {}

var logger = logging.NoOp

// HandlerRegisterer is the symbol the plugin loader will try to load for
// server-side plugin activation. This is a no-op passthrough — the real
// work happens in ClientRegisterer (client.go). The handler must be
// registered in plugin/http-server.name so KrakenD loads the .so and
// discovers the ClientRegisterer alongside it.
var HandlerRegisterer = handlerRegisterer("http-no-pool")

type handlerRegisterer string

func (r handlerRegisterer) RegisterHandlers(f func(
	name string,
	handler func(context.Context, map[string]any, http.Handler) (http.Handler, error),
)) {
	f(string(r), func(_ context.Context, _ map[string]any, next http.Handler) (http.Handler, error) {
		return next, nil
	})
}

func (r handlerRegisterer) RegisterLogger(v any) {
	l, ok := v.(logging.BasicLogger)
	if !ok {
		return
	}
	logger = l
	logger.Info(fmt.Sprintf("[PLUGIN: %s] Logger loaded (server+client)", HandlerRegisterer))
}
