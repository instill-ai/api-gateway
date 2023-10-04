package main

import (
	"context"
	"fmt"
	"net/http"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// HandlerRegisterer is the symbol the plugin loader will try to load. It must implement the Registerer interface
var HandlerRegisterer = handlerRegisterer("grpc-proxy-server")

type handlerRegisterer string

type responseHijacker struct {
	w                    http.ResponseWriter
	grpcStatus           string
	grpcMessage          string
	grpcStatusDetailsBin string
}

type Trailer interface {
	WriteTrailer()
}

func NewResponseHijacker(w http.ResponseWriter) http.ResponseWriter {
	return &responseHijacker{w: w}
}

func (l *responseHijacker) Header() http.Header {
	return l.w.Header()
}

func (l *responseHijacker) Write(b []byte) (int, error) {
	return l.w.Write(b)
}

func (l *responseHijacker) WriteHeader(s int) {
	// Note: hijack the trailers and remove them from headers
	l.grpcStatus = l.w.Header().Get("Grpc-Status")
	l.grpcMessage = l.w.Header().Get("Grpc-Message")
	l.grpcStatusDetailsBin = l.w.Header().Get("Grpc-Status-Details-Bin")
	l.w.Header().Del("Grpc-Status")
	l.w.Header().Del("Grpc-Message")
	l.w.Header().Del("Grpc-Status-Details-Bin")

	l.w.Header().Set("Trailer", "Grpc-Status")
	if l.grpcMessage != "" {
		l.w.Header().Add("Trailer", "Grpc-Message")
	}
	if l.grpcStatusDetailsBin != "" {
		l.w.Header().Add("Trailer", "Grpc-Status-Details-Bin")
	}

	l.w.WriteHeader(s)
}

func (l *responseHijacker) WriteTrailer() {
	l.w.Header().Set("Grpc-Status", l.grpcStatus)
	if l.grpcMessage != "" {
		l.w.Header().Set("Grpc-Message", l.grpcMessage)
	}
	if l.grpcStatusDetailsBin != "" {
		l.w.Header().Set("Grpc-Status-Details-Bin", l.grpcStatusDetailsBin)
	}
}

func (r handlerRegisterer) RegisterHandlers(f func(
	name string,
	handler func(context.Context, map[string]interface{}, http.Handler) (http.Handler, error),
)) {
	f(string(r), r.registerHandlers)
}

func (r handlerRegisterer) registerHandlers(ctx context.Context, extra map[string]interface{}, h http.Handler) (http.Handler, error) {

	return h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w = NewResponseHijacker(w)
		h.ServeHTTP(w, req)
		w.(Trailer).WriteTrailer()
	}), &http2.Server{}), nil

}

func init() {
	fmt.Printf("Plugin: router handler \"%s\" loaded!!!\n", HandlerRegisterer)
}
