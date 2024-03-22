package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	mgmtPB "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
	"google.golang.org/grpc/metadata"
)

// urlRegexp will be aplied to the paths involved in pushing an image. It
// will capture the following fields:
//  1. Repository name
//  2. The namespace segment in the repository, determined by its owner owner.
//  3. The ID of the repository (e.g. the model ID)
//  4. The type of resource in the upload (blob or manifest).
//  5. The resource ID of the updated object. We can extract the tag or
//     digest of the object from here.
//
// For example, given the request `PUT /v2/funky-wombat/llava-34b/manifests/1.0.3-alpha`:
// matches[0]: /v2/funky-wombat/llava-34b/manifests/1.0.3-alpha`
// matches[1]: funky-wombat/llava-34b
// matches[2]: funky-wombat
// matches[3]: llava-34b
// matches[4]: manifests
// matches[5]: 1.0.3-alpha
var urlRegexp = regexp.MustCompile(`/v2/(([^/]+)/([^/]+))/(blobs|manifests)/(.*)`)

type registryHandler struct {
	mgmtPublicClient  mgmtPB.MgmtPublicServiceClient
	mgmtPrivateClient mgmtPB.MgmtPrivateServiceClient

	registryAddr string
}

func newRegistryHandler(config map[string]any) (*registryHandler, error) {
	var mgmtPublicAddr, mgmtPrivateAddr string
	var ok bool
	var rh registryHandler

	if rh.registryAddr, ok = config["hostport"].(string); !ok {
		return nil, fmt.Errorf("invalid registry address")
	}
	if mgmtPublicAddr, ok = config["mgmt_public_hostport"].(string); !ok {
		return nil, fmt.Errorf("invalid mgmt public address")
	}
	if mgmtPrivateAddr, ok = config["mgmt_private_hostport"].(string); !ok {
		return nil, fmt.Errorf("invalid mgmt private address")
	}

	mgmtPublicConn, err := newGRPCConn(mgmtPublicAddr, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to connect with mgmt-backend: %w", err)
	}
	mgmtPrivateConn, err := newGRPCConn(mgmtPrivateAddr, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to connect with mgmt-backend: %w", err)
	}

	rh.mgmtPublicClient = mgmtPB.NewMgmtPublicServiceClient(mgmtPublicConn)
	rh.mgmtPrivateClient = mgmtPB.NewMgmtPrivateServiceClient(mgmtPrivateConn)

	return &rh, nil
}

type registryHandlerParams struct {
	writer   http.ResponseWriter
	req      *http.Request
	username string
	password string
}

func (rh *registryHandler) handler(ctx context.Context) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Authenticate the user via docker login
		username, password, ok := req.BasicAuth()
		if !ok {
			// Challenge the user for basic authentication
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			writeStatusUnauthorized(req, w, "Instill AI user authentication failed")
			return
		}

		params := registryHandlerParams{
			writer:   w,
			req:      req,
			username: username,
			password: password,
		}

		if req.URL.Path == "/v2/" {
			rh.login(ctx, params)
			return
		}

		rh.relay(ctx, params)
	})
}

func (rh *registryHandler) login(ctx context.Context, p registryHandlerParams) {
	req := p.req
	w := p.writer

	// Validate the api key and the namespace authorization
	if !strings.HasPrefix(p.password, "instill_sk_") {
		writeStatusUnauthorized(req, w, "Instill AI user authentication failed")
		return
	}

	ctxWithBearer := metadata.AppendToOutgoingContext(ctx, "Authorization", fmt.Sprintf("Bearer %s", p.password))
	tokenValidation, err := rh.mgmtPublicClient.ValidateToken(ctxWithBearer, &mgmtPB.ValidateTokenRequest{})
	if err != nil {
		writeStatusInternalError(req, w)
		return
	}

	userUID := tokenValidation.UserUid
	// Check if the login username is the same with the user id retrieved from the token validation response
	userLookup, err := rh.mgmtPrivateClient.LookUpUserAdmin(
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

	if userLookup.User.Id != p.username {
		writeStatusUnauthorized(req, w, "Instill AI user authentication failed")
		return
	}

	// To this point, if the url.Path is "/v2/", return 200 OK to the client for login success
	writeStatusOK(p.req, p.writer)
}

func (rh *registryHandler) relay(ctx context.Context, p registryHandlerParams) {
	req := p.req
	w := p.writer

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
	if namespace != p.username {
		parent := fmt.Sprintf("users/%s", p.username)
		resp, err := rh.mgmtPublicClient.ListUserMemberships(ctx, &mgmtPB.ListUserMembershipsRequest{Parent: parent})
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

	req.URL.Scheme = "http"
	req.URL.Host = rh.registryAddr
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
