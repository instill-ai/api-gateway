package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"

	"golang.org/x/net/http2"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"

	"plugin/internal/logger"
)

// ClientRegisterer is the symbol the plugin loader will try to load. It must implement the RegisterClient interface
var ClientRegisterer = registerer("grpc-proxy")

type registerer string

func (registerer) RegisterLogger(v interface{}) {
	logger, _ := logger.GetZapLogger()
	defer func() {
		// can't handle the error due to https://github.com/uber-go/zap/issues/880
		_ = logger.Sync()
	}()

	logger.Debug(fmt.Sprintf("[PLUGIN: %s] Logger loaded", ClientRegisterer))
}

func (r registerer) RegisterClients(f func(
	name string,
	handler func(context.Context, map[string]interface{}) (http.Handler, error),
)) {
	f(string(r), r.registerClients)
}

func (r registerer) registerClients(_ context.Context, extra map[string]interface{}) (http.Handler, error) {

	logger, _ := logger.GetZapLogger()
	defer func() {
		// can't handle the error due to https://github.com/uber-go/zap/issues/880
		_ = logger.Sync()
	}()
	grpc_zap.ReplaceGrpcLoggerV2(logger)

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

		resp, err := httpClient.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		respBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.Warn(err.Error())
		}
		defer resp.Body.Close()

		if resp.Body == nil {
			return
		}
		w.Header().Set("Trailer", "Grpc-Status, Grpc-Message")
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

func main() {}
