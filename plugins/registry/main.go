package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/luraproject/lura/logging"
	"google.golang.org/grpc/metadata"

	mgmtPB "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
)

var (
	pluginName = "registry"

	// HandlerRegisterer is the symbol the plugin loader will try to load. It must implement the Registerer interface
	HandlerRegisterer = registerer(pluginName)

	// urlRegexp will be aplied to the paths involved in pushing an image. It
	// will capture the following fields:
	// 1. Repository name
	// 2. The namespace segment in the repository, determined by its owner owner.
	// 3. The ID of the repository (e.g. the model ID)
	// 4. The type of resource in the upload (blob or manifest).
	// 5. The resource ID of the updated object. We can extract the tag or
	//    digest of the object from here.
	//
	// For example, given the request `PUT /v2/funky-wombat/llava-34b/manifests/1.0.3-alpha`:
	// matches[0]: /v2/funky-wombat/llava-34b/manifests/1.0.3-alpha`
	// matches[1]: funky-wombat/llava-34b
	// matches[2]: funky-wombat
	// matches[3]: llava-34b
	// matches[4]: manifests
	// matches[5]: 1.0.3-alpha
	urlRegexp = regexp.MustCompile(`/v2/(([^/]+)/([^/]+))/(blobs|manifests)/(.*)`)
)

type registerer string

func (r registerer) RegisterHandlers(f func(
	name string,
	handler func(context.Context, map[string]interface{}, http.Handler) (http.Handler, error),
)) {
	f(string(r), r.registerHandlers)
}

func (r registerer) registerHandlers(ctx context.Context, extra map[string]interface{}, h http.Handler) (http.Handler, error) {
	config, ok := extra[pluginName].(map[string]interface{})
	if !ok {
		return h, fmt.Errorf("configuration not found")
	}

	hostport, _ := config["hostport"].(string)
	mgmtPublicClient, _ := initMgmtPublicServiceClient(ctx, config["mgmt_public_hostport"].(string), "", "")
	mgmtPrivateClient, _ := initMgmtPrivateServiceClient(ctx, config["mgmt_private_hostport"].(string), "", "")

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		url := req.URL

		// If the URL path starts with "/v2/" (exactly /v2/ indicating the
		// first handshake request to confirm registry V2 API), it means that
		// the request is intended for the Instill Artifact registry. In this
		// case, the traffic is hijacked and directly relayed to the registry.
		// Otherwise, if the URL path does not match any of these patterns, the
		// traffic is passed through to the next handler.
		if !strings.HasPrefix(url.Path, "/v2/") {
			h.ServeHTTP(w, req)
			return
		}

		// Authenticate the user via docker login
		username, password, ok := req.BasicAuth()
		if !ok {
			// Challenge the user for basic authentication
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			writeStatusUnauthorized(req, w, "Instill AI user authentication failed")
			return
		}

		if url.Path == "/v2/" {
			// Validate the api key and the namespace authorization
			if !strings.HasPrefix(password, "instill_sk_") {
				writeStatusUnauthorized(req, w, "Instill AI user authentication failed")
				return
			}

			ctx := metadata.AppendToOutgoingContext(ctx, "Authorization", fmt.Sprintf("Bearer %s", password))
			tokenValidation, err := mgmtPublicClient.ValidateToken(
				ctx,
				&mgmtPB.ValidateTokenRequest{},
			)
			if err != nil {
				writeStatusInternalError(req, w)
				return
			}

			userUID := tokenValidation.UserUid
			// Check if the login username is the same with the user id retrieved from the token validation response
			userLookup, err := mgmtPrivateClient.LookUpUserAdmin(
				ctx,
				&mgmtPB.LookUpUserAdminRequest{
					Permalink: "users/" + userUID,
				},
			)
			if err != nil {
				logger.Error(err.Error())
				writeStatusInternalError(req, w)
				return
			}

			if userLookup.User.Id != username {
				writeStatusUnauthorized(req, w, "Instill AI user authentication failed")
				return
			}

			// To this point, if the url.Path is "/v2/", return 200 OK to the client for login success
			writeStatusOK(req, w)
			return
		}

		// Docker image tag format:
		// [registry]/[namespace]/[repository path]:[image tag]
		// The namespace is the user uid or the organization uid
		var namespace string
		matches := urlRegexp.FindStringSubmatch(req.URL.Path)
		if len(matches) == 0 {
			errStr := "Namespace is not found in the image name. " +
				"Docker registry hosted in Instill Artifact has a format " +
				"[registry]/[namespace]/[repository path]:[image tag]. " +
				"A namespace can be a user or organization ID."
			writeStatusUnauthorized(req, w, errStr)
			return
		}

		namespace = matches[2]
		resourceType := matches[4]
		resourceID := matches[5]

		// If the username and the namespace is not the same, check if the
		// namespace is an organisation name where the user has the membership.
		if namespace != username {
			parent := fmt.Sprintf("users/%s", username)
			resp, err := mgmtPublicClient.ListUserMemberships(ctx, &mgmtPB.ListUserMembershipsRequest{Parent: parent})
			if err != nil {
				writeStatusUnauthorized(req, w, "Instill AI user authentication failed")
				return
			}

			isValid := false
			for _, membership := range resp.Memberships {
				if namespace == membership.Organization.Name && membership.State == mgmtPB.MembershipState_MEMBERSHIP_STATE_ACTIVE {
					isValid = true
					break
				}
			}

			if !isValid {
				writeStatusUnauthorized(req, w, "Instill AI user authentication failed")
				return
			}
		}

		url.Scheme = "http"
		url.Host = hostport
		req.RequestURI = ""

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			logger.Error(err.Error())
			writeStatusInternalError(req, w)
			return
		}

		if req.Method == http.MethodPut && resourceType == "manifests" {
			logger.Info("tag:", resourceID)
			logger.Info("model ID:", matches[1])
			// DIGEST AT dockerContentDigest := resp.Header.Get("Docker-Content-Digest")
			// CREATION TIME createdAt := time.Now()

			// 1. Call create tag endpoint in artifact-backend (should be idempotent)
			// 2. Call model for deployment
		}

		// Copy headers, status codes, and body from the backend to the response writer
		for k, hs := range resp.Header {
			for _, h := range hs {
				w.Header().Add(k, h)
			}
		}

		w.WriteHeader(resp.StatusCode)
		w.Header().Set("Content-Type", "application/json")
		io.Copy(w, resp.Body)
		resp.Body.Close()

	}), nil
}

func writeStatusUnauthorized(req *http.Request, w http.ResponseWriter, error string) {
	if req.ProtoMajor == 2 && strings.Contains(req.Header.Get("Content-Type"), "application/grpc") {
		w.Header().Set("Content-Type", "application/grpc")
		w.Header().Set("Trailer", "Grpc-Status")
		w.Header().Add("Trailer", "Grpc-Message")
		w.Header().Set("Grpc-Status", "16")
		w.Header().Set("Grpc-Message", "Unauthenticated")
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
	}
	fmt.Fprintln(w, error)
}

func writeStatusInternalError(req *http.Request, w http.ResponseWriter) {
	if req.ProtoMajor == 2 && strings.Contains(req.Header.Get("Content-Type"), "application/grpc") {
		w.Header().Set("Content-Type", "application/grpc")
		w.Header().Set("Trailer", "Grpc-Status")
		w.Header().Add("Trailer", "Grpc-Message")
		w.Header().Set("Grpc-Status", "13")
		w.Header().Set("Grpc-Message", "INTERNAL")
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
	}
}

func writeStatusOK(req *http.Request, w http.ResponseWriter) {
	if req.ProtoMajor == 2 && strings.Contains(req.Header.Get("Content-Type"), "application/grpc") {
		w.Header().Set("Content-Type", "application/grpc")
		w.Header().Set("Trailer", "Grpc-Status")
		w.Header().Add("Trailer", "Grpc-Message")
		w.Header().Set("Grpc-Status", "0")
		w.Header().Set("Grpc-Message", "OK")
	} else {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
	}
}

func main() {}

// This logger is replaced by the RegisterLogger method to load the one from KrakenD
var logger = logging.NoOp

func (registerer) RegisterLogger(v interface{}) {
	l, ok := v.(logging.Logger)
	if !ok {
		return
	}
	logger = l
	logger.Info(fmt.Sprintf("[PLUGIN: %s] Logger loaded", HandlerRegisterer))
}
