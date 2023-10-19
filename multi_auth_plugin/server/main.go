package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
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

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken struct {
		Sub string `json:"sub"`
	} `json:"access_token"`
}

type ValidateTokenResp struct {
	UserUid string `json:"user_uid"`
}

func writeStatusUnauthorized(req *http.Request, w http.ResponseWriter) {

	if req.ProtoMajor == 2 && strings.Contains(req.Header.Get("Content-Type"), "application/grpc") {
		w.Header().Set("Content-Type", "application/grpc")
		w.Header().Set("Trailer", "Grpc-Status")
		w.Header().Add("Trailer", "Grpc-Message")
		w.Header().Set("Grpc-Status", "16")               // UNAUTHENTICATED
		w.Header().Set("Grpc-Message", "Unauthenticated") // UNAUTHENTICATED
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

	client := &http.Client{}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		authorization := req.Header.Get("Authorization")

		if req.URL.String() == "/__health" {
			h.ServeHTTP(w, req)
		} else if req.URL.String() == "/core/v1alpha/validate_token" {
			h.ServeHTTP(w, req)
		} else if req.URL.String() == "/core/v1alpha/auth/login" {
			h.ServeHTTP(w, req)
		} else if strings.HasPrefix(authorization, "Basic ") || strings.HasPrefix(authorization, "basic ") {
			basicAuth := strings.Split(authorization, " ")[1]

			basicAuthDecoded, err := base64.StdEncoding.DecodeString(basicAuth)
			if err != nil {
				writeStatusUnauthorized(req, w)
				return
			}

			loginRequest := LoginRequest{
				Username: strings.Split(string(basicAuthDecoded), ":")[0],
				Password: strings.Split(string(basicAuthDecoded), ":")[1],
			}
			loginRequestJson, err := json.Marshal(loginRequest)
			if err != nil {
				writeStatusUnauthorized(req, w)
				return
			}

			loginReq, err := http.NewRequest("POST", config["token_issuer_endpoint"].(string), bytes.NewBuffer(loginRequestJson))
			if err != nil {
				writeStatusUnauthorized(req, w)
				return
			}

			loginResponseJson, err := client.Do(loginReq)
			if err != nil {
				writeStatusUnauthorized(req, w)
				return
			}

			defer loginResponseJson.Body.Close()

			respBody, err := ioutil.ReadAll(loginResponseJson.Body)
			if err != nil {
				writeStatusUnauthorized(req, w)
				return
			}
			var loginResponse LoginResponse

			err = json.Unmarshal(respBody, &loginResponse)

			if err != nil {
				writeStatusUnauthorized(req, w)
				return
			}

			req.Header.Set("jwt-sub", loginResponse.AccessToken.Sub)
			h.ServeHTTP(w, req)

		} else if strings.HasPrefix(authorization, "Bearer instill_sk_") || strings.HasPrefix(authorization, "bearer instill_sk_") {
			reqValidate, err := http.NewRequest("POST", config["token_validation_endpoint"].(string), nil)

			if err != nil {
				writeStatusUnauthorized(req, w)
				return
			}
			reqValidate.Header = req.Header
			resValidate, err := client.Do(reqValidate)

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
	}), nil

}

func init() {
	fmt.Printf("Plugin: router handler \"%s\" loaded!!!\n", HandlerRegisterer)
}

func main() {}
