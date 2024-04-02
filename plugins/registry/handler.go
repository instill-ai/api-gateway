package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"google.golang.org/grpc/metadata"

	mgmtPB "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
	modelPB "github.com/instill-ai/protogen-go/model/model/v1alpha"
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
	mgmtPublicClient   mgmtPB.MgmtPublicServiceClient
	mgmtPrivateClient  mgmtPB.MgmtPrivateServiceClient
	modelPublicClient  modelPB.ModelPublicServiceClient
	modelPrivateClient modelPB.ModelPrivateServiceClient

	registryAddr string
}

func newRegistryHandler(config map[string]any) (*registryHandler, error) {
	var mgmtPublicAddr, mgmtPrivateAddr, modelPublicAddr, modelPrivateAddr string
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
	if modelPublicAddr, ok = config["model_public_hostport"].(string); !ok {
		return nil, fmt.Errorf("invalid model public address")
	}
	if modelPrivateAddr, ok = config["model_private_hostport"].(string); !ok {
		return nil, fmt.Errorf("invalid model private address")
	}

	mgmtPublicConn, err := newGRPCConn(mgmtPublicAddr, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to connect with mgmt-backend: %w", err)
	}
	mgmtPrivateConn, err := newGRPCConn(mgmtPrivateAddr, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to connect with mgmt-backend: %w", err)
	}
	modelPublicConn, err := newGRPCConn(modelPublicAddr, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to connect with model-backend: %w", err)
	}
	modelPrivateConn, err := newGRPCConn(modelPrivateAddr, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to connect with model-backend: %w", err)
	}

	rh.mgmtPublicClient = mgmtPB.NewMgmtPublicServiceClient(mgmtPublicConn)
	rh.mgmtPrivateClient = mgmtPB.NewMgmtPrivateServiceClient(mgmtPrivateConn)
	rh.modelPublicClient = modelPB.NewModelPublicServiceClient(modelPublicConn)
	rh.modelPrivateClient = modelPB.NewModelPrivateServiceClient(modelPrivateConn)

	return &rh, nil
}

type registryHandlerParams struct {
	writer  http.ResponseWriter
	req     *http.Request
	userID  string
	userUID string
}

func (rh *registryHandler) handler(ctx context.Context) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Authenticate the user via docker login
		username, password, ok := req.BasicAuth()
		if !ok {
			// Challenge the user for basic authentication
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			writeStatus(req, w, http.StatusUnauthorized, "Unauthenticated", "Instill AI user authentication failed")
			return
		}

		// Validate the api key and the namespace authorization
		if !strings.HasPrefix(password, "instill_sk_") {
			writeStatus(req, w, http.StatusUnauthorized, "Unauthenticated", "Instill AI user authentication failed")
			return
		}

		ctx = withBearerAuth(ctx, password)
		tokenValidation, err := rh.mgmtPublicClient.ValidateToken(ctx, &mgmtPB.ValidateTokenRequest{})
		if err != nil {
			writeStatus(req, w, http.StatusInternalServerError, "INTERNAL", "")
			return
		}

		params := registryHandlerParams{
			writer:  w,
			req:     req,
			userID:  username,
			userUID: tokenValidation.UserUid,
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

	// Check if the login username is the same with the user id retrieved from the token validation response
	userLookup, err := rh.mgmtPrivateClient.LookUpUserAdmin(
		ctx,
		&mgmtPB.LookUpUserAdminRequest{
			Permalink: "users/" + p.userUID,
		},
	)
	if err != nil {
		logger.Error(err.Error())
		writeStatus(req, w, http.StatusInternalServerError, "INTERNAL", "")
		return
	}

	if userLookup.User.Id != p.userID {
		writeStatus(req, w, http.StatusUnauthorized, "Unauthenticated", "Instill AI user authentication failed")
		return
	}

	// To this point, if the url.Path is "/v2/", return 200 OK to the client for login success
	writeStatus(p.req, p.writer, http.StatusOK, "OK", "")
}

func (rh *registryHandler) relay(ctx context.Context, p registryHandlerParams) {
	req := p.req
	w := p.writer

	// Docker image tag format:
	// [registry]/[namespace]/[repository path]:[image tag]
	// The namespace is the user uid or the organization uid
	matches := urlRegexp.FindStringSubmatch(req.URL.Path)
	if len(matches) == 0 {
		errStr := "Namespace is not found in the image name. " +
			"Docker registry hosted in Instill Artifact has a format " +
			"[registry]/[namespace]/[repository path]:[image tag]. " +
			"A namespace can be a user or organization ID."
		writeStatus(req, w, http.StatusUnauthorized, "Unauthenticated", errStr)
		return
	}

	namespace := matches[2]
	contentID := matches[3]
	resourceType := matches[4]
	resourceID := matches[5]

	// If the username and the namespace is not the same, check if the
	// namespace is an organisation name where the user has the membership.
	isOrganizationRepository := false
	if namespace != p.userID {
		isOrganizationRepository = true

		parent := fmt.Sprintf("users/%s", p.userID)
		resp, err := rh.mgmtPublicClient.ListUserMemberships(ctx, &mgmtPB.ListUserMembershipsRequest{Parent: parent})
		if err != nil {
			writeStatus(req, w, http.StatusInternalServerError, "INTERNAL", "")
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
			writeStatus(req, w, http.StatusUnauthorized, "Unauthenticated", "Instill AI user authentication failed")
			return
		}
	}

	// Check the existence of the model namespace before continuing with the push.
	ctx = withUserUIDAuth(ctx, p.userUID)
	if req.Method == http.MethodHead {
		var err error
		if isOrganizationRepository {
			name := fmt.Sprintf("organizations/%s/models/%s", namespace, contentID)
			_, err = rh.modelPublicClient.GetOrganizationModel(ctx, &modelPB.GetOrganizationModelRequest{
				Name: name,
				View: modelPB.View_VIEW_BASIC.Enum(),
			})
		} else {
			name := fmt.Sprintf("users/%s/models/%s", namespace, contentID)
			_, err = rh.modelPublicClient.GetUserModel(ctx, &modelPB.GetUserModelRequest{
				Name: name,
				View: modelPB.View_VIEW_BASIC.Enum(),
			})
		}
		if err != nil {
			writeStatus(req, w, http.StatusPreconditionFailed, "FAILED_PRECONDITION", "Resource namespace does not exist")
			return
		}
	}

	req.URL.Scheme = "http"
	req.URL.Host = rh.registryAddr
	req.RequestURI = ""

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logger.Error(err.Error())
		writeStatus(req, w, http.StatusInternalServerError, "INTERNAL", "")
		return
	}

	if req.Method == http.MethodPut && resourceType == "manifests" && resp.StatusCode == http.StatusCreated {
		digest := resp.Header.Get("Docker-Content-Digest")

		// TODO Call create tag endpoint in artifact-backend.

		// Deploy model.The previous operations are idempotent so it should be
		// safe to repeat them if we fail here.
		//
		// TODO in the future the registry will handle more than model images,
		// so this operation won't always be necessary. A much better pattern
		// is publishing the push operation success as an event and let the
		// clients to consume and act upon it (artifact to register the tag
		// creation time, model to deploy the image...).
		prefix := "users"
		if isOrganizationRepository {
			prefix = "organizations"
		}
		deployReq := &modelPB.DeployModelAdminRequest{
			Name:    fmt.Sprintf("%s/%s/models/%s", prefix, namespace, contentID),
			Version: resourceID,
			Digest:  digest,
		}
		if _, err := rh.modelPrivateClient.DeployModelAdmin(ctx, deployReq); err != nil {
			logger.Error(err.Error())
			writeStatus(req, w, http.StatusInternalServerError, "INTERNAL", "")
			return
		}
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

func writeStatus(req *http.Request, w http.ResponseWriter, statusCode int, message string, err string) {
	if req.ProtoMajor == 2 && strings.Contains(req.Header.Get("Content-Type"), "application/grpc") {
		var grpcStatus string
		switch statusCode {
		case http.StatusOK:
			grpcStatus = "0"
		case http.StatusPreconditionFailed:
			grpcStatus = "9"
		case http.StatusInternalServerError:
			grpcStatus = "13"
		case http.StatusUnauthorized:
			grpcStatus = "16"
		default:
			grpcStatus = "13"
		}
		w.Header().Set("Content-Type", "application/grpc")
		w.Header().Set("Trailer", "Grpc-Status")
		w.Header().Add("Trailer", "Grpc-Message")
		w.Header().Set("Grpc-Status", grpcStatus)
		w.Header().Set("Grpc-Message", message)
	} else {
		w.WriteHeader(statusCode)
		w.Header().Set("Content-Type", "application/json")
	}

	if err != "" {
		fmt.Fprintln(w, err)
	}
}

func withBearerAuth(ctx context.Context, bearer string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "Authorization", fmt.Sprintf("Bearer %s", bearer))
}

func withUserUIDAuth(ctx context.Context, uid string) context.Context {
	return metadata.AppendToOutgoingContext(ctx,
		"Instill-Auth-Type", "user",
		"Instill-User-Uid", uid,
	)
}
