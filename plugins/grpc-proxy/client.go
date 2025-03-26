package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/luraproject/lura/v2/logging"
	"golang.org/x/net/http2"
)

// ClientRegisterer is the symbol the plugin loader will try to load. It must implement the RegisterClient interface
var ClientRegisterer = clientRegisterer("grpc-proxy-client")

type clientRegisterer string

func (r clientRegisterer) RegisterClients(f func(
	name string,
	handler func(context.Context, map[string]any) (http.Handler, error),
)) {
	f(string(r), r.registerClients)
}

func (r clientRegisterer) registerClients(_ context.Context, extra map[string]any) (http.Handler, error) {

	// check the passed configuration and initialize the plugin
	name, ok := extra["name"].(string)
	if !ok {
		return nil, errors.New("wrong config")
	}
	if name != string(r) {
		return nil, fmt.Errorf("unknown register %s", name)
	}

	// return the actual handler wrapping or your custom logic so it can be used as a replacement for the default http handler
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

		req.Header.Set("content-type", "application/grpc")

		tr := &http2.Transport{
			AllowHTTP: true,
			DialTLS: func(netw, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(netw, addr)
			},
		}

		httpClient := http.Client{Transport: tr}
		defer httpClient.CloseIdleConnections()

		resp, err := httpClient.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Warning(err.Error())
		}
		defer resp.Body.Close()

		if resp.Body == nil {
			return
		}
		for k, hs := range resp.Header {
			for _, h := range hs {
				w.Header().Add(k, h)
			}
		}

		// We can only get Trailer after reading the body
		for k, hs := range resp.Trailer {
			for _, h := range hs {
				w.Header().Add(k, h)
			}
		}

		w.WriteHeader(resp.StatusCode)
		w.Write(respBytes)

	}), nil
}

func (clientRegisterer) RegisterLogger(v any) {
	l, ok := v.(logging.BasicLogger)
	if !ok {
		return
	}
	logger = l
	logger.Debug(fmt.Sprintf("[PLUGIN: %s] Logger loaded", ClientRegisterer))
}
