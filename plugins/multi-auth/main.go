package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/luraproject/lura/v2/logging"
	"google.golang.org/grpc/metadata"

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

	config, ok := extra[pluginName].(map[string]interface{})
	if !ok {
		return h, errors.New("configuration not found")
	}

	mgmtClient, _ := InitMgmtPublicServiceClient(context.Background(), config["grpc_server"].(string), "", "")

	httpClient := http.Client{Transport: http.DefaultTransport}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		authorization := req.Header.Get("Authorization")

		if req.URL.String() == "/__health" {
			h.ServeHTTP(w, req)
		} else if req.URL.String() == "/v1beta/validate_token" || req.URL.String() == "/core/v1beta/validate_token" {
			h.ServeHTTP(w, req)
		} else if req.URL.String() == "/v1beta/auth/login" || req.URL.String() == "/core/v1beta/auth/login" {
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

		} else if req.Header.Get("instill-use-sse") == "true" {
			// Currently, KrakenD doesn’t support event-stream. To make
			// authentication work, we send a request to the management API
			// first for verification.
			r, err := http.NewRequest("GET", "http://localhost:8080/v1beta/user", nil)
			r.Header = req.Header
			r.Header.Del("instill-use-sse")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			resp, err := httpClient.Do(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if resp.StatusCode == 401 {
				writeStatusUnauthorized(req, w)
				return
			}
			type user struct {
				User struct {
					UID string `json:"uid"`
				} `json:"user"`
			}
			respBytes, err := io.ReadAll(resp.Body)
			if err != nil {
				writeStatusUnauthorized(req, w)
				return
			}
			defer resp.Body.Close()

			u := user{}
			err = json.Unmarshal(respBytes, &u)
			if err != nil {
				writeStatusUnauthorized(req, w)
				return
			}

			req.Header.Set("Instill-Auth-Type", "user")
			req.Header.Set("Instill-User-Uid", u.User.UID)
			req.Header.Set("instill-Use-SSE", "true")
			h.ServeHTTP(w, req)

		} else {
			req.Header.Set("Instill-Auth-Type", "user")
			req.URL.Path = "/internal" + req.URL.Path
			h.ServeHTTP(w, req)
		}
	}), nil

}

func main() {}

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
