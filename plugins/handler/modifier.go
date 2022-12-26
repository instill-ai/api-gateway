package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"

	"time"

	"github.com/chilts/sid"
	"go.uber.org/zap"
)

// HandlerRegisterer is the symbol the plugin loader will try to load. It must implement the Registerer interface
var HandlerRegisterer = registerer("modifier")

type registerer string

func (r registerer) RegisterHandlers(f func(
	name string,
	handler func(context.Context, map[string]interface{}, http.Handler) (http.Handler, error),
)) {
	f(string(r), r.registerHandlers)
}

// Error data model identically defined by OpenAPI in all api resources
type Error struct {

	// Error code
	Status int32 `json:"status,omitempty"`

	// Short description of the error
	Title string `json:"title,omitempty"`

	// Human-readable error message
	Detail string `json:"detail,omitempty"`

	// The duration in seconds (s) it takes for a request to be processed
	Duration float64 `json:"duration,omitempty"`
}

func errorFactory(status int32, title string) func(detail string, duration float64) []byte {
	return func(detail string, duration float64) []byte {

		if status == http.StatusInternalServerError {
			if string(detail[len(detail)-1]) != "." {
				detail = detail + "."
			}
			detail = detail + " Retry or contact support at support@instill.tech"
		}

		json, _ := json.Marshal(&Error{
			Status:   status,
			Title:    title,
			Detail:   detail,
			Duration: duration,
		})

		return json
	}
}

var newUnauthorizedError func(string, float64) []byte = errorFactory(http.StatusUnauthorized, "Unauthorized")
var newStatusForbiddenError func(string, float64) []byte = errorFactory(http.StatusForbidden, "Access Deny")
var newStatusNotFoundError func(string, float64) []byte = errorFactory(http.StatusNotFound, "Not Found")
var newStatusMethodNotAllowedError func(string, float64) []byte = errorFactory(http.StatusMethodNotAllowed, "Method Not Allowed")
var newInternalServerError func(string, float64) []byte = errorFactory(http.StatusInternalServerError, "Internal Server Error")

func replaceSelfHost(data interface{}, apiGatewayHost string) error {

	switch value := data.(type) {
	case *interface{}:
		if data, ok := (*value).(map[string]interface{}); ok {
			replaceSelfHost(&data, apiGatewayHost)
		}
	case *map[string]interface{}:
		for k, v := range *value {
			switch v.(type) {
			case string:
				if k == "self" {
					u, err := url.Parse(v.(string))
					if err != nil {
						return err
					}
					u.Host = apiGatewayHost
					(*value)[k] = u.String()
				}
			case interface{}:
				if data, ok := v.([]interface{}); ok {
					for i := range data {
						if data, ok := data[i].(map[string]interface{}); ok {
							err := replaceSelfHost(&data, apiGatewayHost)
							if err != nil {
								return err
							}
						}
					}
				} else if data, ok := v.(map[string]interface{}); ok {
					err := replaceSelfHost(&data, apiGatewayHost)
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func insertDuration(data interface{}, duration time.Duration) error {

	switch value := data.(type) {
	case *interface{}: // if root json is an array
		if data, ok := (*value).([]interface{}); ok {
			insertDuration(&data, duration)
		}
		if data, ok := (*value).(map[string]interface{}); ok {
			insertDuration(&data, duration)
		}
	case *[]interface{}:
		for i := range *value {
			v := (*value)[i].(map[string]interface{})
			v["duration"] = duration.Seconds()
		}
	case *map[string]interface{}:
		(*value)["duration"] = duration.Seconds()
	}

	return nil
}

type HTTPResponseInterceptor struct {
	http.ResponseWriter
	StatusCode int
	http.Flusher
}

// NewHTTPResponseInterceptor create new httpInterceptor
func NewHTTPResponseInterceptor(w http.ResponseWriter) *HTTPResponseInterceptor {
	return &HTTPResponseInterceptor{w, http.StatusOK, w.(http.Flusher)}
}

// WriteHeader override response WriteHeader
func (i *HTTPResponseInterceptor) WriteHeader(code int) {
	// log.Error(fmt.Sprintf("Status code %d\n", code))
	i.ResponseWriter.WriteHeader(code)
}

// func (i *HTTPResponseInterceptor) Write(bytes []byte) (int, error) {
// 	log.Error(fmt.Sprintf("Body %v\n", bytes))
// 	log.Error(fmt.Sprintf("Body %s\n", string(bytes)))
// 	return i.ResponseWriter.Write(bytes)
// }

// Header function overwrites the http.ResponseWriter Header() function
func (i *HTTPResponseInterceptor) Header() http.Header {
	// log.Error(fmt.Sprintf("Header %v\n", i.ResponseWriter.Header()))
	return i.ResponseWriter.Header()
}

func (r registerer) registerHandlers(ctx context.Context, extra map[string]interface{}, h http.Handler) (http.Handler, error) {
	// Check the passed configuration and initialize the plugin
	name, ok := extra["name"].(string)

	if !ok {
		return nil, errors.New("wrong config")
	}

	if name != string(r) {
		return nil, fmt.Errorf("unknown register %s", name)
	}

	// Return the actual handler wrapping or your custom logic so it can be used as a replacement for the default http handler
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// According to https://github.com/grpc/grpc/blob/master/doc/PROTOCOL-HTTP2.md#requests
		// The Content-Type of gRPC has the "application/grpc" prefix

		if req.Proto == "HTTP/2.0" && req.ProtoMajor == 2 {
			start := time.Now()

			// Insert a request id (23 char string) to the request header
			reqId := sid.IdBase64()
			req.Header.Set("Instill-Request-Id", reqId)
			req.Header.Set("Request-Id", reqId) // Deprecated, remove until all backends have removed it
			req.Header.Set("Content-Type", "application/grpc")

			ww := NewHTTPResponseInterceptor(rw)
			h.ServeHTTP(ww, req)
			rw.(http.Flusher).Flush()

			rw.Header().Add("grpc-status", "0")

			end := time.Now()
			duration := end.Sub(start)

			// Custom logger for cloud logging
			reqLogger, _ := zap.NewProduction(zap.Fields(
				zap.String("backend", ww.Header().Get("Backend")),
				zap.Int("statusCode", ww.StatusCode),
				zap.Int64("latencyMs", duration.Milliseconds()),
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path),
				zap.String("query", req.URL.RawQuery),
				zap.String("referer", req.Referer()),
				zap.String("userAgent", req.UserAgent()),
				zap.String("instillRequestId", req.Header.Get("Instill-Request-Id")),
				zap.String("jwtSub", req.Header.Get("Jwt-Sub")),
				zap.String("jwtClientId", req.Header.Get("Jwt-Client-Id")),
				zap.String("jwtScope", req.Header.Get("Jwt-Scope")),
				zap.String("jwtModels", req.Header.Get("Jwt-Models")),
				zap.String("time", end.Format(time.RFC3339)),
				zap.Int64("contentLength", req.ContentLength),
				zap.String("contentType", req.Header.Get("Content-Type")),
				zap.String("userName", req.Header.Get("Jwt-Username")),
			))
			defer reqLogger.Sync() // flushes buffer, if any

			reqLogger.Info("")
		} else {

			nr := httptest.NewRecorder()

			start := time.Now()

			// Insert a request id (23 char string) to the request header
			reqId := sid.IdBase64()
			req.Header.Set("Instill-Request-Id", reqId)
			req.Header.Set("Request-Id", reqId) // Deprecated, remove until all backends have removed it

			h.ServeHTTP(nr, req)

			end := time.Now()
			duration := end.Sub(start)

			res := nr.Result()

			// Copy the response header to the header of ResponseWriter
			for k, v := range res.Header {
				rw.Header()[k] = v
			}

			// Insert the request id to the response header
			rw.Header().Set("Instill-Request-Id", req.Header.Get("Instill-Request-Id"))

			// Custom logger for cloud logging
			reqLogger, _ := zap.NewProduction(zap.Fields(
				zap.String("backend", res.Header.Get("Backend")),
				zap.Int("statusCode", res.StatusCode),
				zap.Int64("latencyMs", duration.Milliseconds()),
				zap.String("method", req.Method),
				zap.String("path", req.URL.Path),
				zap.String("query", req.URL.RawQuery),
				zap.String("referer", req.Referer()),
				zap.String("userAgent", req.UserAgent()),
				zap.String("instillRequestId", req.Header.Get("Instill-Request-Id")),
				zap.String("jwtSub", req.Header.Get("Jwt-Sub")),
				zap.String("jwtClientId", req.Header.Get("Jwt-Client-Id")),
				zap.String("jwtScope", req.Header.Get("Jwt-Scope")),
				zap.String("jwtModels", req.Header.Get("Jwt-Models")),
				zap.String("time", end.Format(time.RFC3339)),
				zap.Int64("contentLength", req.ContentLength),
				zap.String("contentType", req.Header.Get("Content-Type")),
				zap.String("userName", req.Header.Get("Jwt-Username")),
			))
			defer reqLogger.Sync() // flushes buffer, if any

			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				reqLogger.Error(err.Error())
				rw.Header().Del("Backend")
				rw.Header().Set("Content-Type", "application/problem+json")
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write(newInternalServerError("Something went wrong", duration.Seconds()))
				return
			}
			defer res.Body.Close()

			// No logging for `__health` endpoint
			if req.URL.Path == "/__health" {
				rw.Header().Del("Backend")
				rw.WriteHeader(res.StatusCode)
				rw.Write(body)
				return
			}

			if len(body) == 0 {

				// If JWT validation failed in the first place
				if res.StatusCode == http.StatusUnauthorized {
					msg := "Request is unauthorized"
					// Log Authorization header for 401 and 403 responses
					reqLogger.Info(msg, zap.String("authorization", req.Header.Get("Authorization")))
					rw.Header().Del("Backend")
					rw.Header().Set("Content-Type", "application/problem+json")
					rw.WriteHeader(http.StatusUnauthorized)
					rw.Write(newUnauthorizedError(msg, duration.Seconds()))
					return
				}

				// If JWT validation failed in the first place (forbidden)
				if res.StatusCode == http.StatusForbidden {
					msg := "The request does not have access rights to the requested resource."
					// Log Authorization header for 401 and 403 responses
					reqLogger.Info(msg, zap.String("authorization", req.Header.Get("Authorization")))
					rw.Header().Del("Backend")
					rw.Header().Set("Content-Type", "application/problem+json")
					rw.WriteHeader(http.StatusForbidden)
					rw.Write(newStatusForbiddenError(msg, duration.Seconds()))
					return
				}

				// If endpoint not found
				if res.StatusCode == http.StatusNotFound {
					msg := "The requested resource is not found."
					reqLogger.Info(msg)
					rw.Header().Del("Backend")
					rw.Header().Set("Content-Type", "application/problem+json")
					rw.WriteHeader(http.StatusNotFound)
					rw.Write(newStatusNotFoundError(msg, duration.Seconds()))
					return
				}

				// If unsupported method
				if res.StatusCode == http.StatusMethodNotAllowed {
					msg := fmt.Sprintf("%s method not allowed", req.Method)
					reqLogger.Info(msg)
					rw.Header().Del("Backend")
					rw.Header().Set("Content-Type", "application/problem+json")
					rw.WriteHeader(http.StatusMethodNotAllowed)
					rw.Write(newStatusMethodNotAllowedError(msg, duration.Seconds()))
					return
				}

				// If internal server error
				if res.StatusCode == http.StatusInternalServerError {
					msg := "Something went wrong"
					reqLogger.Error(msg)
					rw.Header().Del("Backend")
					rw.Header().Set("Content-Type", "application/problem+json")
					rw.WriteHeader(http.StatusInternalServerError)
					rw.Write(newInternalServerError(msg, duration.Seconds()))
					return
				}

				// Others
				reqLogger.Info("")
				rw.Header().Del("Backend")
				rw.WriteHeader(res.StatusCode)
				rw.Write(body)
				return
			}

			var data interface{}
			err = json.Unmarshal(body, &data)
			if err != nil {
				reqLogger.Error(err.Error())
				rw.Header().Del("Backend")
				rw.Header().Set("Content-Type", "application/problem+json")
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write(newInternalServerError("Something went wrong", duration.Seconds()))
				return
			}

			// Recursively replace all self values with req.Host
			err = replaceSelfHost(&data, req.Host)
			if err != nil {
				reqLogger.Error(err.Error())
				rw.Header().Del("Backend")
				rw.Header().Set("Content-Type", "application/problem+json")
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write(newInternalServerError("Something went wrong", duration.Seconds()))
				return
			}

			// Insert duration field
			err = insertDuration(&data, duration)
			if err != nil {
				reqLogger.Error(err.Error())
				rw.Header().Del("Backend")
				rw.Header().Set("Content-Type", "application/problem+json")
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write(newInternalServerError("Something went wrong", duration.Seconds()))
				return
			}

			output, err := json.Marshal(data)
			if err != nil {
				reqLogger.Error(err.Error())
				rw.Header().Del("Backend")
				rw.Header().Set("Content-Type", "application/problem+json")
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write(newInternalServerError("Something went wrong", duration.Seconds()))
				return
			}

			// Update Content-Length after the modifier logic
			rw.Header().Del("Backend")
			rw.Header().Set("Content-Length", strconv.Itoa(len(output)))

			// Write back to the response body
			rw.WriteHeader(res.StatusCode)
			rw.Write(output)

			if res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden {
				// Log Authorization header for 401 and 403 responses
				reqLogger.Info(string(output), zap.String("authorization", req.Header.Get("Authorization")))
			} else if res.StatusCode >= http.StatusBadRequest {
				reqLogger.Info(string(output))
			} else {
				// No log body for non error response
				reqLogger.Info("")
			}
		}

	}), nil
}

func init() {
	fmt.Printf("Plugin: router handler \"%s\" loaded!!!\n", HandlerRegisterer)

}

func main() {}
