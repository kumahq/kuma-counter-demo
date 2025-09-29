package base

import (
	"encoding/json"
	"fmt"
	"github.com/kumahq/kuma-counter-demo/pkg/api"
	"go.opentelemetry.io/otel/trace"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
)

type ServerImpl struct {
	logger  *slog.Logger
	kvUrl   string
	kv      map[string]api.KV
	version string
	sync.Mutex
}

func NewServerImpl(logger *slog.Logger, kvUrl string, version string) api.ServerInterface {
	logger.Info("starting handler with", "kv-url", kvUrl, "version", version)
	return &ServerImpl{
		logger:  logger,
		kv:      map[string]api.KV{},
		kvUrl:   kvUrl,
		version: version,
	}

}

func (s *ServerImpl) GetVersion(w http.ResponseWriter, r *http.Request) {
	res := api.VersionResponse{Version: s.version}
	s.writeResponse(w, r, http.StatusOK, res, nil)
}

func (s *ServerImpl) writeResponse(w http.ResponseWriter, r *http.Request, originalStatusCode int, response interface{}, err error) {
	statusCode := originalStatusCode
	if originalStatusCode/100 < 5 { // If the original status is 5xx ignore our override
		statusStr := r.Header.Get("x-set-response-status-code")
		if st, _ := strconv.Atoi(statusStr); st != 0 {
			originalStatusCode = st
		}
	}
	
	// Handle fault injection via x-failure-percentage header
	if originalStatusCode/100 < 4 { // Only inject faults for successful responses (2xx, 3xx)
		failurePercentageStr := r.Header.Get("x-failure-percentage")
		if failurePercentage, parseErr := strconv.Atoi(failurePercentageStr); parseErr == nil && failurePercentage > 0 {
			if failurePercentage >= 100 || rand.Intn(100) < failurePercentage {
				originalStatusCode = http.StatusInternalServerError
				qType := api.INTERNAL_ERROR.QualifiedType()
				span := trace.SpanFromContext(r.Context())
				response = api.Error{
					Type:     &qType,
					Status:   http.StatusInternalServerError,
					Instance: span.SpanContext().TraceID().String(),
					Title:    fmt.Sprintf("Simulated failure (x-failure-percentage: %d%%)", failurePercentage),
				}
				err = fmt.Errorf("simulated failure with %d%% chance", failurePercentage)
				slog.DebugContext(r.Context(), "injecting simulated failure", "failurePercentage", failurePercentage)
			}
		}
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("X-Demo-App-Version", s.version)
	w.WriteHeader(originalStatusCode)
	if err != nil {
		slog.ErrorContext(r.Context(), "failed request with error", "err", err, "status", statusCode, "response", response)
	} else {
		slog.DebugContext(r.Context(), "successful request", "statusCode", statusCode, "originalStatusCode", originalStatusCode, "response", response)
	}
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		fmt.Printf("failed to write response: %v\n", err)
	}
}

func (s *ServerImpl) writeErrorResponse(w http.ResponseWriter, r *http.Request, statusCode int, errorType api.ErrorType, err error, format string, args ...any) {
	qType := errorType.QualifiedType()
	span := trace.SpanFromContext(r.Context())
	s.writeResponse(w, r, statusCode, api.Error{Type: &qType, Status: statusCode, Instance: span.SpanContext().TraceID().String(), Title: fmt.Sprintf(format, args...)}, err)
}
