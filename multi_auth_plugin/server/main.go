package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// pluginName is the plugin name
var pluginName = "multi-auth"

// HandlerRegisterer is the symbol the plugin loader will try to load. It must implement the Registerer interface
var HandlerRegisterer = registerer(pluginName)

type registerer string

func (r registerer) RegisterHandlers(f func(
	name string,
	handler func(context.Context, map[string]interface{}, http.Handler) (http.Handler, error),
)) {
	f(string(r), r.registerHandlers)
}

type ValidateTokenResp struct {
	UserUid string `json:"user_uid"`
}

func writeStatusUnauthorized(req *http.Request, w http.ResponseWriter) {

	if req.ProtoMajor == 2 && strings.Contains(req.Header.Get("Content-Type"), "application/grpc") {
		w.Header().Set("Content-Type", "application/grpc")
		w.Header().Set("Trailer", "Grpc-Status, Grpc-Message")
		w.Header().Set("Grpc-Status", "16") // UNAUTHENTICATED
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
	}
}

func (r registerer) registerHandlers(ctx context.Context, extra map[string]interface{}, h http.Handler) (http.Handler, error) {

	logger, _ := GetZapLogger()
	defer func() {
		// can't handle the error due to https://github.com/uber-go/zap/issues/880
		_ = logger.Sync()
	}()
	grpc_zap.ReplaceGrpcLoggerV2(logger)

	config, ok := extra[pluginName].(map[string]interface{})
	if !ok {
		return h, errors.New("configuration not found")
	}

	return h2c.NewHandler(
		http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			authorization := req.Header.Get("Authorization")

			if req.URL.String() == "/__health" {
				h.ServeHTTP(w, req)
			} else if req.URL.String() == "/base/v1alpha/validate_token" {
				h.ServeHTTP(w, req)
			} else if req.URL.String() == "/base/v1alpha/auth/login" {
				h.ServeHTTP(w, req)
			} else if strings.HasPrefix(authorization, "Bearer instill_sk_") || strings.HasPrefix(authorization, "bearer instill_sk_") {
				reqValidate, err := http.NewRequest("POST", config["token_validation_endpoint"].(string), nil)

				if err != nil {
					writeStatusUnauthorized(req, w)
					return
				}
				reqValidate.Header = req.Header
				resValidate, err := http.DefaultClient.Do(reqValidate)

				if err != nil {
					writeStatusUnauthorized(req, w)
					return
				}
				defer resValidate.Body.Close()
				if resValidate.StatusCode == 200 {
					resValidateStruct := &ValidateTokenResp{}
					json.NewDecoder(resValidate.Body).Decode(resValidateStruct)
					req.Header.Set("jwt-sub", resValidateStruct.UserUid)
					h.ServeHTTP(w, req)
				} else {
					writeStatusUnauthorized(req, w)
				}

			} else {
				req.URL.Path = "/internal" + req.URL.Path
				h.ServeHTTP(w, req)
			}
		}),
		&http2.Server{},
	), nil

}

func init() {
	fmt.Printf("Plugin: router handler \"%s\" loaded!!!\n", HandlerRegisterer)
}

func main() {}
