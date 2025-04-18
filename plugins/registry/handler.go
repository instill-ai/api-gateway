package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/distribution/distribution/registry/api/errcode"
	"google.golang.org/grpc/metadata"

	grpccodes "google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"

	artifactpb "github.com/instill-ai/protogen-go/artifact/artifact/v1alpha"
	mgmtpb "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
	modelpb "github.com/instill-ai/protogen-go/model/model/v1alpha"
)

// urlRegexp will be applied to the paths involved in pushing an image. It
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
	mgmtPublicClient      mgmtpb.MgmtPublicServiceClient
	mgmtPrivateClient     mgmtpb.MgmtPrivateServiceClient
	modelPublicClient     modelpb.ModelPublicServiceClient
	modelPrivateClient    modelpb.ModelPrivateServiceClient
	artifactPrivateClient artifactpb.ArtifactPrivateServiceClient

	registryAddr string
}

func newRegistryHandler(config map[string]any) (*registryHandler, error) {
	var (
		mgmtPublicAddr      string
		mgmtPrivateAddr     string
		modelPublicAddr     string
		modelPrivateAddr    string
		artifactPrivateAddr string
	)
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
	if artifactPrivateAddr, ok = config["artifact_private_hostport"].(string); !ok {
		return nil, fmt.Errorf("invalid artifact private address")
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
	artifactPrivateConn, err := newGRPCConn(artifactPrivateAddr, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to connect with artifact-backend: %w", err)
	}

	rh.mgmtPublicClient = mgmtpb.NewMgmtPublicServiceClient(mgmtPublicConn)
	rh.mgmtPrivateClient = mgmtpb.NewMgmtPrivateServiceClient(mgmtPrivateConn)
	rh.modelPublicClient = modelpb.NewModelPublicServiceClient(modelPublicConn)
	rh.modelPrivateClient = modelpb.NewModelPrivateServiceClient(modelPrivateConn)
	rh.artifactPrivateClient = artifactpb.NewArtifactPrivateServiceClient(artifactPrivateConn)

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
			rh.handleError(req, w, authErr)
			return
		}

		// Validate the api key and the namespace authorization
		if !strings.HasPrefix(password, "instill_sk_") {
			rh.handleError(req, w, authErr)
			return
		}

		ctx = withBearerAuth(ctx, password)
		tokenValidation, err := rh.mgmtPublicClient.ValidateToken(ctx, &mgmtpb.ValidateTokenRequest{})
		if err != nil {
			switch grpcstatus.Convert(err).Code() {
			case grpccodes.Unauthenticated:
				rh.handleError(req, w, authErr)
			default:
				rh.handleError(req, w, fmt.Errorf("validating token: %w", err))
			}

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
	lookupReq := &mgmtpb.LookUpUserAdminRequest{UserUid: p.userUID}
	userLookup, err := rh.mgmtPrivateClient.LookUpUserAdmin(ctx, lookupReq)
	if err != nil {
		rh.handleError(req, w, fmt.Errorf("looking up user: %w", err))
		return
	}

	if userLookup.User.Id != p.userID {
		rh.handleError(req, w, authErr)
		return
	}
}

func (rh *registryHandler) relay(ctx context.Context, p registryHandlerParams) {
	req := p.req
	w := p.writer

	// Docker image tag format:
	// [registry]/[namespace]/[repository path]:[image tag]
	// The namespace is the user uid or the organization uid
	matches := urlRegexp.FindStringSubmatch(req.URL.Path)
	if len(matches) == 0 {
		msg := "Artifacts in Instill registry should have the format " +
			"<namespace>/<id>, where namespace can be a user or organization ID"
		rh.handleError(req, w, errcode.ErrorCodeDenied.WithMessage(msg))

		return
	}

	repository := matches[1]
	namespace := matches[2]
	contentID := matches[3]
	resourceType := matches[4]
	resourceID := matches[5]

	// If the username and the namespace is not the same, check if the
	// namespace is an organisation name where the user has the membership.
	if namespace != p.userID {
		ctx := withUserUIDAuth(ctx, p.userUID)

		resp, err := rh.mgmtPublicClient.ListUserMemberships(ctx, &mgmtpb.ListUserMembershipsRequest{UserId: p.userID})
		if err != nil {
			rh.handleError(req, w, fmt.Errorf("checking organization: %w", err))
			return
		}

		isValid := false
		for _, membership := range resp.Memberships {
			if namespace == membership.Organization.Id && membership.State == mgmtpb.MembershipState_MEMBERSHIP_STATE_ACTIVE {
				isValid = true
				break
			}
		}

		if !isValid {
			rh.handleError(req, w, authErr)
			return
		}
	}

	// Check the existence of the model namespace before continuing with the push.
	if req.Method == http.MethodHead {
		ctx := withUserUIDAuth(ctx, p.userUID)

		var name string
		var err error

		_, err = rh.modelPublicClient.GetNamespaceModel(ctx, &modelpb.GetNamespaceModelRequest{
			NamespaceId: namespace,
			ModelId:     contentID,
		})
		if err != nil {
			switch grpcstatus.Convert(err).Code() {
			case grpccodes.NotFound:
				logger.Warning(req, "model", name, "doesn't exist: ", err)
				rh.handleNameUnknown(w, "model "+name+" doesn't exist")
			default:
				rh.handleError(req, w, fmt.Errorf("validating namespace: %w", err))
			}
			return
		}
	}

	req.URL.Scheme = "http"
	req.URL.Host = rh.registryAddr
	req.RequestURI = ""

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		rh.handleError(req, w, fmt.Errorf("relaying request: %w", err))
		return
	}

	if req.Method == http.MethodPut && resourceType == "manifests" && resp.StatusCode == http.StatusCreated {
		digest := resp.Header.Get("Docker-Content-Digest")

		createTagReq := &artifactpb.CreateRepositoryTagRequest{
			Tag: &artifactpb.RepositoryTag{
				Digest: digest,
				Name:   fmt.Sprintf("repositories/%s/tags/%s", repository, resourceID),
				Id:     resourceID,
			},
		}
		if _, err := rh.artifactPrivateClient.CreateRepositoryTag(ctx, createTagReq); err != nil {
			rh.handleError(req, w, fmt.Errorf("creating tag: %w", err))
			return
		}

		// Deploy model.The previous operations are idempotent so it should be
		// safe to repeat them if we fail here.
		//
		// TODO in the future the registry will handle more than model images,
		// so this operation won't always be necessary. A much better pattern
		// is publishing the push operation success as an event and let the
		// clients to consume and act upon it (artifact to register the tag
		// creation time, model to deploy the image...).

		if _, err := rh.modelPrivateClient.DeployNamespaceModelAdmin(ctx, &modelpb.DeployNamespaceModelAdminRequest{
			NamespaceId: namespace,
			ModelId:     contentID,
			Version:     resourceID,
			Digest:      digest,
		}); err != nil {
			rh.handleError(req, w, fmt.Errorf("deploying model: %w", err))
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

func withBearerAuth(ctx context.Context, bearer string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "Authorization", fmt.Sprintf("Bearer %s", bearer))
}

func withUserUIDAuth(ctx context.Context, uid string) context.Context {
	return metadata.AppendToOutgoingContext(ctx,
		"Instill-Auth-Type", "user",
		"Instill-User-Uid", uid,
	)
}

var (
	authErr = errcode.ErrorCodeUnauthorized.WithDetail("Instill AI user authentication failed")
)

func (rh *registryHandler) handleError(req *http.Request, w http.ResponseWriter, e error) {
	logWarning(req, e)

	if err := errcode.ServeJSON(w, e); err != nil {
		logError(req, "failed to handle error;", "original error:", e, ", failure reason:", err)
	}
}

// handleNameUnknown should be the equivalent of
// `errcode.ServeJSON(w, errcode.ErrorCodeNameUnknown)`. That handler, however,
// produces a response that triggers a retry mechanism in the client. If the
// repository name is unknown, the outcome won't change by retrying the
// request, so this handler returns a response compliant with the v2 API that
// aborts the OCI image push.
func (rh *registryHandler) handleNameUnknown(w http.ResponseWriter, msg string) {
	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Content-Type", "application/json")

	resp := fmt.Sprintf(`{"errors": [{"code": "NAME_UNKNOWN", "message": "%s"}]}`, msg)
	fmt.Fprintln(w, resp)
}
