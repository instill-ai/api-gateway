package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/gofrs/uuid"
	artifactpb "github.com/instill-ai/protogen-go/artifact/artifact/v1alpha"
	mgmtPb "github.com/instill-ai/protogen-go/core/mgmt/v1beta"
)

type blobHandler struct {
	mgmtPublicClient      mgmtPb.MgmtPublicServiceClient
	mgmtPrivateClient     mgmtPb.MgmtPrivateServiceClient
	artifactPrivateClient artifactpb.ArtifactPrivateServiceClient
	minioAddr             string
}

func newBlobHandler(config map[string]any) (*blobHandler, error) {
	var (
		mgmtPublicAddr      string
		mgmtPrivateAddr     string
		artifactPrivateAddr string
	)
	var ok bool
	var rh blobHandler

	if rh.minioAddr, ok = config["minio_hostport"].(string); !ok {
		return nil, fmt.Errorf("the minio address is not set")
	}

	if mgmtPublicAddr, ok = config["mgmt_public_hostport"].(string); !ok {
		return nil, fmt.Errorf("the mgmt public address is not set")
	}
	if mgmtPrivateAddr, ok = config["mgmt_private_hostport"].(string); !ok {
		return nil, fmt.Errorf("the mgmt private address is not set")
	}
	if artifactPrivateAddr, ok = config["artifact_private_hostport"].(string); !ok {
		return nil, fmt.Errorf("the artifact private address is not set")
	}

	mgmtPublicConn, err := newGRPCConn(mgmtPublicAddr, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to connect with mgmt-backend: %w ", err)
	}
	mgmtPrivateConn, err := newGRPCConn(mgmtPrivateAddr, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to connect with mgmt-backend: %w ", err)
	}

	artifactPrivateConn, err := newGRPCConn(artifactPrivateAddr, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to connect with artifact-backend: %w", err)
	}

	rh.mgmtPublicClient = mgmtPb.NewMgmtPublicServiceClient(mgmtPublicConn)
	rh.mgmtPrivateClient = mgmtPb.NewMgmtPrivateServiceClient(mgmtPrivateConn)
	rh.artifactPrivateClient = artifactpb.NewArtifactPrivateServiceClient(artifactPrivateConn)
	return &rh, nil
}

type blobHandlerParams struct {
	writer http.ResponseWriter
	req    *http.Request
	// UserUID   string
	ObjectURL *artifactpb.ObjectURL

	// object
	Object *artifactpb.Object
}

// handler is the http handler for the blob plugin
func (rh *blobHandler) handler(ctx context.Context) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		Info(req.Method+" "+req.URL.Path, " start relaying request to blob backend")

		userUID := req.Header.Get("Instill-User-Uid")

		// check if the userUID is a valid uuid
		if _, err := uuid.FromString(userUID); err != nil {
			Error(req.URL.Path, " authorization failed. ", err)
			rh.handleError(req, w, err)
			return
		}

		// TODO: check if user uuid is member of the namespace or user self

		// NOTE: the object url uid is the last part of the request path
		parts := strings.Split(req.URL.Path, "/")
		objectURLUID := parts[len(parts)-1]

		if _, err := uuid.FromString(objectURLUID); err != nil {
			Error(req.URL.Path, " object url uid is not a uuid ", err)
			rh.handleError(req, w, err)
			return
		}
		objectURLInfo, err := rh.artifactPrivateClient.GetObjectURL(ctx, &artifactpb.GetObjectURLRequest{
			Uid: objectURLUID,
		})
		if err != nil {
			Error(req.URL.Path, " get object url info failed ", err)
			rh.handleError(req, w, err)
			return
		}

		// Note: first milestone will not check if the object url is expired
		// check if the object url is expired
		// if objectURLInfo.ObjectUrl.GetUrlExpireAt().AsTime().Before(time.Now()) {
		// 	rh.handleError(req, w, fmt.Errorf(" object url expired "))
		// 	return
		// }

		// get object info
		object, err := rh.artifactPrivateClient.GetObject(ctx, &artifactpb.GetObjectRequest{
			Uid: objectURLInfo.ObjectUrl.GetObjectUid(),
		})
		if err != nil {
			Error(req.URL.Path, " get object info failed ", err)
			rh.handleError(req, w, err)
			return
		}

		params := blobHandlerParams{
			writer: w,
			req:    req,
			// UserUID:   userUID,
			ObjectURL: objectURLInfo.ObjectUrl,
			Object:    object.Object,
		}

		rh.relay(ctx, params)
	})
}

// TODO: handle upload and download
func (rh *blobHandler) relay(ctx context.Context, p blobHandlerParams) {
	req := p.req
	w := p.writer

	// check if the method is PUT
	if req.Method == http.MethodPut {
		upload(ctx, req, w, rh, p.ObjectURL)
	} else if req.Method == http.MethodGet {
		download(ctx, req, w, rh, p.ObjectURL, p.Object)
	} else {
		rh.handleError(req, w, fmt.Errorf("method not supported "))
	}
}

// upload a test file with content "test" to the presigned URL using http client
func upload(ctx context.Context, req *http.Request, w http.ResponseWriter, rh *blobHandler, objectURL *artifactpb.ObjectURL) error {
	// rh.mgmtPrivateClient.CheckNamespaceAdmin()
	originalURL := req.URL.String()
	req.URL.Scheme = "http"
	req.URL.Host = rh.minioAddr
	req.RequestURI = ""
	pathAndQuery := objectURL.GetMinioUrlPath()
	// split the path and query then set to req
	parts := strings.Split(pathAndQuery, "?")
	req.URL.Path = parts[0]
	req.URL.RawQuery = parts[1]
	client := &http.Client{}
	defer req.Body.Close()
	var byteCounter int64
	teeReader := io.TeeReader(req.Body, &countingWriter{&byteCounter})

	newReq, err := http.NewRequest(http.MethodPut, req.URL.String(), teeReader)
	if err != nil {
		Error(newReq.URL.Path, "failed to create request", err)
		rh.handleError(newReq, w, err)
		return err
	}
	// set content type from original request
	// check if the content type is in the request header
	if contentType := req.Header.Get("Content-Type"); contentType != "" {
		newReq.Header.Set("Content-Type", contentType)
	} else {
		rh.handleError(req, w, fmt.Errorf("content type is not set in the header"))
		return fmt.Errorf("content type is not set in the header")
	}
	// set content length header from original request
	if contentLength := req.Header.Get("Content-Length"); contentLength != "" {
		newReq.Header.Set("Content-Length", contentLength)
	} else {
		rh.handleError(req, w, fmt.Errorf("content length is not set in the header"))
		return fmt.Errorf("content length is not set in the header")
	}
	// set keep alive from original request
	newReq.Header.Set("Connection", req.Header.Get("Connection"))
	// set accept encoding from original request
	newReq.Header.Set("Accept-Encoding", req.Header.Get("Accept-Encoding"))
	// cache control from original request
	newReq.Header.Set("Cache-Control", req.Header.Get("Cache-Control"))
	newReq.ContentLength = req.ContentLength

	// last modified time from original request
	if lastModifiedTime := req.Header.Get("Last-Modified"); lastModifiedTime != "" {
		newReq.Header.Set("Last-Modified", lastModifiedTime)
		lastModifiedTime, err := time.Parse(time.RFC1123, req.Header.Get("Last-Modified"))
		if err != nil {
			Error(req.URL.Path, " failed to parse last modified time ", err)
			rh.handleError(req, w, err)
			return err
		}
		newReq.Header.Set("Last-Modified", lastModifiedTime.Format(time.RFC1123))
	}
	newResp, err := client.Do(newReq)
	if err != nil {
		Error(req.URL.Path, "failed to upload file", err)
		rh.handleError(req, w, err)
		return err
	}
	defer newResp.Body.Close()

	// set the status code
	w.WriteHeader(newResp.StatusCode)
	// Copy headers, status codes, and body from the backend to the response writer
	for k, hs := range newResp.Header {
		for _, h := range hs {
			w.Header().Add(k, h)
		}
	}

	written, err := io.Copy(w, newResp.Body)
	if err != nil {
		Error("PUT ", originalURL, " upload file failed", err)
		rh.handleError(req, w, err)
		return err
	}

	// NOTE: if the written bytes is 0, it means the upload is successful.
	if written == 0 {
		Info(
			"PUT ",
			originalURL,
			" upload file success, ",
			byteCounter,
			" bytes transferred, Content-Length: ",
			req.ContentLength,
			" bytes, Content-Type: ",
			req.Header.Get("Content-Type"),
			", Object UID: ",
			objectURL.GetObjectUid(),
			" Namespace UID: ",
			objectURL.NamespaceUid,
		)
		isUploaded := true
		contentType := req.Header.Get("Content-Type")
		grpcReq := &artifactpb.UpdateObjectRequest{
			Uid:        objectURL.GetObjectUid(),
			Size:       &byteCounter,
			IsUploaded: &isUploaded,
			Type:       &contentType,
		}
		lastModifiedTime := req.Header.Get("Last-Modified")
		if lastModifiedTime != "" {
			lastModifiedTime, err := time.Parse(time.RFC1123, lastModifiedTime)
			if err != nil {
				Error(req.URL.Path, "failed to parse last modified time", err)
				rh.handleError(req, w, err)
				return err
			} else {
				grpcReq.LastModifiedTime = timestamppb.New(lastModifiedTime)
			}
		}
		_, err = rh.artifactPrivateClient.UpdateObject(ctx, grpcReq)
		if err != nil {
			Error(req.URL.Path, "failed to update object info", err)
			rh.handleError(req, w, err)
			return err
		}
	}

	return nil
}

func download(_ context.Context, req *http.Request, w http.ResponseWriter, rh *blobHandler, objectURL *artifactpb.ObjectURL, object *artifactpb.Object) error {
	originalURL := req.URL.String()
	req.URL.Scheme = "http"
	req.URL.Host = rh.minioAddr
	req.RequestURI = ""
	pathAndQuery := objectURL.GetMinioUrlPath()
	parts := strings.Split(pathAndQuery, "?")
	req.URL.Path = parts[0]
	req.URL.RawQuery = parts[1]

	client := &http.Client{}

	newReq, err := http.NewRequest(http.MethodGet, req.URL.String(), nil)
	if err != nil {
		Error(req.URL.Path, "failed to create request ", err)
		rh.handleError(req, w, err)
		return err
	}

	// Copy all headers from the original request to the new request
	for k, hs := range req.Header {
		if k != "Authorization" {
			for _, h := range hs {
				newReq.Header.Set(k, h)
			}
		}
	}

	newResp, err := client.Do(newReq)
	if err != nil {
		Error(req.URL.Path, "failed to download file ", err)
		rh.handleError(req, w, err)
		return err
	}
	defer newResp.Body.Close()

	// Copy headers from the backend response to the response writer
	for k, hs := range newResp.Header {
		if k != "Access-Control-Allow-Origin" {
			for _, h := range hs {
				w.Header().Add(k, h)
			}
		}
	}

	// set attachment header
	w.Header().Set("Content-Disposition", "attachment; filename="+object.GetName())

	w.WriteHeader(newResp.StatusCode)

	var written int64

	// Copy the body from the backend response to the response writer
	// Only copy the body if the status is not 304 Not Modified
	if newResp.StatusCode != http.StatusNotModified {
		written, err = io.Copy(w, newResp.Body)
		if err != nil {
			Error("GET ", originalURL, " download file failed ", err)
			rh.handleError(req, w, err)
			return err
		}
	}

	Info(
		"GET ",
		originalURL,
		" download file success, ",
		written,
		" bytes transferred, Content-Type: ",
		newResp.Header.Get("Content-Type"),
		", Object UID: ",
		objectURL.GetObjectUid(),
		" Namespace UID: ",
		objectURL.NamespaceUid,
	)

	return nil
}

func (rh *blobHandler) handleError(req *http.Request, w http.ResponseWriter, e error) {
	Error(req.URL.Path, e)

	// Set the status code
	w.WriteHeader(http.StatusInternalServerError)

	// Set the content type to JSON
	w.Header().Set("Content-Type", "application/json")

	// Create an error response
	errorResponse := struct {
		Message string `json:"message"`
	}{
		Message: e.Error(),
	}

	// Encode the error response as JSON
	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		Error(req.URL.Path, "failed to encode error response", err)
	}
}

type countingWriter struct {
	bytesWritten *int64
}

func (cw *countingWriter) Write(p []byte) (int, error) {
	*cw.bytesWritten += int64(len(p))
	return len(p), nil
}
