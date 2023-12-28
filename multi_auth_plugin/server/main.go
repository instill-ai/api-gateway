package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gofrs/uuid"
	"google.golang.org/grpc/metadata"

	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	mgmtPB "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
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

	mgmtClient, _ := InitMgmtPublicServiceClient(context.Background(), config["grpc_server"].(string), "", "")

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		authorization := req.Header.Get("Authorization")

		if req.URL.String() == "/__health" {
			h.ServeHTTP(w, req)
		} else if req.URL.String() == "/core/v1beta/validate_token" {
			h.ServeHTTP(w, req)
		} else if req.URL.String() == "/core/v1beta/auth/login" {
			h.ServeHTTP(w, req)
		} else if strings.HasPrefix(authorization, "Basic ") || strings.HasPrefix(authorization, "basic ") {
			basicAuth := strings.Split(authorization, " ")[1]

			basicAuthDecoded, err := base64.StdEncoding.DecodeString(basicAuth)
			if err != nil {
				writeStatusUnauthorized(req, w)
				return
			}

			resp, err := mgmtClient.AuthTokenIssuer(ctx, &mgmtPB.AuthTokenIssuerRequest{
				Username: strings.Split(string(basicAuthDecoded), ":")[0],
				Password: strings.Split(string(basicAuthDecoded), ":")[1],
			})
			if err != nil {
				writeStatusUnauthorized(req, w)
				return
			}

			req.Header.Set("Instill-Auth-Type", "user")
			req.Header.Set("Instill-User-Uid", resp.AccessToken.Sub)
			h.ServeHTTP(w, req)

		} else if strings.HasPrefix(authorization, "Bearer instill_sk_") || strings.HasPrefix(authorization, "bearer instill_sk_") {

			ctx = metadata.AppendToOutgoingContext(context.Background(), "Authorization", req.Header.Get("authorization"))
			resp, err := mgmtClient.ValidateToken(ctx, &mgmtPB.ValidateTokenRequest{})
			if err != nil {
				writeStatusUnauthorized(req, w)
				return
			}
			req.Header.Set("Instill-Auth-Type", "user")
			req.Header.Set("Instill-User-Uid", resp.UserUid)
			h.ServeHTTP(w, req)

		} else if authorization == "" {
			visitorID, _ := uuid.NewV4()
			req.Header.Set("Instill-Auth-Type", "visitor")
			req.Header.Set("Instill-Visitor-Uid", visitorID.String())
			h.ServeHTTP(w, req)

		} else {
			req.Header.Set("Instill-Auth-Type", "user")
			req.URL.Path = "/internal" + req.URL.Path
			h.ServeHTTP(w, req)
		}
	}), nil

}

func init() {
	fmt.Printf("Plugin: router handler \"%s\" loaded!!!\n", HandlerRegisterer)
}

func main() {}
