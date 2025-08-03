package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/distribution/distribution/registry/api/errcode"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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
		// Extract OpenTelemetry context from incoming request
		otelCtx := req.Context()

		// Create a span for the registry handler
		tracer := trace.SpanFromContext(otelCtx).TracerProvider().Tracer("registry")
		spanCtx, span := tracer.Start(otelCtx, "registry.handler",
			trace.WithAttributes(
				attribute.String("registry.action", "authenticate_and_process"),
			),
		)
		defer span.End()

		// Authenticate the user via docker login
		username, password, ok := req.BasicAuth()
		if !ok {
			// Challenge the user for basic authentication
			span.SetAttributes(attribute.String("auth.error", "missing_basic_auth"))
			span.SetStatus(codes.Error, "Missing basic authentication")
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			rh.handleError(req, w, authErr)
			return
		}

		span.SetAttributes(attribute.String("auth.username", username))

		// Validate the api key and the namespace authorization
		if !strings.HasPrefix(password, "instill_sk_") {
			span.SetAttributes(attribute.String("auth.error", "invalid_api_key_format"))
			span.SetStatus(codes.Error, "Invalid API key format")
			rh.handleError(req, w, authErr)
			return
		}

		// Create a child span for token validation
		_, tokenSpan := tracer.Start(spanCtx, "registry.validate_token",
			trace.WithAttributes(
				attribute.String("grpc.method", "ValidateToken"),
				attribute.String("grpc.service", "mgmtpb.MgmtPublicService"),
			),
		)
		defer tokenSpan.End()

		// Use the original context with bearer auth, not the span context
		authCtx := withBearerAuth(ctx, password)
		tokenValidation, err := rh.mgmtPublicClient.ValidateToken(authCtx, &mgmtpb.ValidateTokenRequest{})
		if err != nil {
			tokenSpan.RecordError(err)
			tokenSpan.SetStatus(codes.Error, err.Error())
			span.RecordError(err)

			switch grpcstatus.Convert(err).Code() {
			case grpccodes.Unauthenticated:
				span.SetAttributes(attribute.String("auth.error", "unauthorized"))
				rh.handleError(req, w, authErr)
			default:
				span.SetAttributes(attribute.String("auth.error", "validation_failed"))
				rh.handleError(req, w, fmt.Errorf("validating token: %w", err))
			}

			return
		}

		// Record successful authentication
		tokenSpan.SetAttributes(attribute.String("auth.user_uid", tokenValidation.UserUid))
		span.SetAttributes(attribute.String("auth.user_uid", tokenValidation.UserUid))

		params := registryHandlerParams{
			writer:  w,
			req:     req,
			userID:  username,
			userUID: tokenValidation.UserUid,
		}

		if req.URL.Path == "/v2/" {
			span.SetAttributes(attribute.String("registry.action", "login"))
			rh.login(ctx, params)
			return
		}

		span.SetAttributes(attribute.String("registry.action", "relay"))
		rh.relay(ctx, params)
	})
}

func (rh *registryHandler) login(ctx context.Context, p registryHandlerParams) {
	req := p.req
	w := p.writer

	// Create a child span for login processing
	tracer := trace.SpanFromContext(ctx).TracerProvider().Tracer("registry")
	loginSpanCtx, loginSpan := tracer.Start(ctx, "registry.login",
		trace.WithAttributes(
			attribute.String("registry.action", "user_lookup"),
		),
	)
	defer loginSpan.End()

	// Check if the login username is the same with the user id retrieved from the token validation response
	lookupReq := &mgmtpb.LookUpUserAdminRequest{UserUid: p.userUID}

	// Create a child span for the gRPC call
	_, grpcSpan := tracer.Start(loginSpanCtx, "registry.grpc_lookup_user_admin",
		trace.WithAttributes(
			attribute.String("grpc.method", "LookUpUserAdmin"),
			attribute.String("grpc.service", "mgmtpb.MgmtPrivateService"),
		),
	)
	defer grpcSpan.End()

	userLookup, err := rh.mgmtPrivateClient.LookUpUserAdmin(ctx, lookupReq)
	if err != nil {
		grpcSpan.RecordError(err)
		grpcSpan.SetStatus(codes.Error, err.Error())
		loginSpan.RecordError(err)
		loginSpan.SetStatus(codes.Error, err.Error())
		rh.handleError(req, w, fmt.Errorf("looking up user: %w", err))
		return
	}

	grpcSpan.SetAttributes(attribute.String("user.id", userLookup.User.Id))
	loginSpan.SetAttributes(attribute.String("user.id", userLookup.User.Id))

	if userLookup.User.Id != p.userID {
		loginSpan.SetAttributes(attribute.String("auth.error", "username_mismatch"))
		loginSpan.SetStatus(codes.Error, "Username mismatch")
		rh.handleError(req, w, authErr)
		return
	}

	loginSpan.SetAttributes(attribute.String("auth.result", "success"))
}

func (rh *registryHandler) relay(ctx context.Context, p registryHandlerParams) {
	req := p.req
	w := p.writer

	// Create a child span for relay processing
	tracer := trace.SpanFromContext(ctx).TracerProvider().Tracer("registry")
	_, relaySpan := tracer.Start(ctx, "registry.relay",
		trace.WithAttributes(
			attribute.String("registry.action", "relay_request"),
		),
	)
	defer relaySpan.End()

	// Docker image tag format:
	// [registry]/[namespace]/[repository path]:[image tag]
	// The namespace is the user uid or the organization uid
	matches := urlRegexp.FindStringSubmatch(req.URL.Path)
	if len(matches) == 0 {
		relaySpan.SetAttributes(attribute.String("registry.error", "invalid_url_format"))
		relaySpan.SetStatus(codes.Error, "Invalid URL format")
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

	relaySpan.SetAttributes(
		attribute.String("registry.repository", repository),
		attribute.String("registry.namespace", namespace),
		attribute.String("registry.content_id", contentID),
		attribute.String("registry.resource_type", resourceType),
		attribute.String("registry.resource_id", resourceID),
	)

	// If the username and the namespace is not the same, check if the
	// namespace is an organization name where the user has the membership.
	if namespace != p.userID {
		// Create a child span for membership check
		_, membershipSpan := tracer.Start(ctx, "registry.check_membership",
			trace.WithAttributes(
				attribute.String("registry.action", "check_organization_membership"),
			),
		)
		defer membershipSpan.End()

		authCtx := withUserUIDAuth(ctx, p.userUID)

		// Create a child span for the gRPC call
		_, grpcSpan := tracer.Start(ctx, "registry.grpc_list_user_memberships",
			trace.WithAttributes(
				attribute.String("grpc.method", "ListUserMemberships"),
				attribute.String("grpc.service", "mgmtpb.MgmtPublicService"),
			),
		)
		defer grpcSpan.End()

		resp, err := rh.mgmtPublicClient.ListUserMemberships(authCtx, &mgmtpb.ListUserMembershipsRequest{UserId: p.userID})
		if err != nil {
			grpcSpan.RecordError(err)
			grpcSpan.SetStatus(codes.Error, err.Error())
			membershipSpan.RecordError(err)
			membershipSpan.SetStatus(codes.Error, err.Error())
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
			membershipSpan.SetAttributes(attribute.String("auth.error", "no_organization_membership"))
			membershipSpan.SetStatus(codes.Error, "No valid organization membership")
			rh.handleError(req, w, authErr)
			return
		}

		membershipSpan.SetAttributes(attribute.String("auth.result", "valid_membership"))
	}

	// Check the existence of the model namespace before continuing with the push.
	if req.Method == http.MethodHead {
		authCtx := withUserUIDAuth(ctx, p.userUID)

		var name string
		var err error

		_, err = rh.modelPublicClient.GetNamespaceModel(authCtx, &modelpb.GetNamespaceModelRequest{
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
