package base

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kumahq/kuma-counter-demo/app/internal/api"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/trace"
	"log/slog"
	"net/http"
	"strconv"
)

const (
	COUNTER_KEY = "counter"
	ZONE_KEY    = "zone"
)

type ServerImpl struct {
	redisClient *redis.Client
	color       string
	version     string
}

var (
	DemoAppErrorType = "https://github.com/kumahq/kuma-counter-demo/blob/master/ERRORS.md#DEMOAPP-FAILURE"
)

func redisError(ctx context.Context, statusCode int, title string) api.Error {
	redisErrorType := "https://github.com/kumahq/kuma-counter-demo/blob/master/ERRORS.md#REDIS-FAILURE"
	span := trace.SpanFromContext(ctx)
	return api.Error{Type: &redisErrorType, Status: statusCode, Instance: span.SpanContext().TraceID().String(), Title: title}
}

func NewServerImpl(client *redis.Client, version string, color string) api.ServerInterface {
	return &ServerImpl{
		redisClient: client,
		version:     version,
		color:       color,
	}

}

func (s *ServerImpl) DeleteCounter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := s.redisClient.Del(ctx, COUNTER_KEY).Err(); err != nil {
		writeResponse(w, r, http.StatusInternalServerError, redisError(ctx, http.StatusInternalServerError, "failed to delete counter"), err)
		return
	}

	zone, err := s.redisClient.Get(ctx, ZONE_KEY).Result()
	if errors.Is(err, redis.Nil) {
		zone = ""
	} else if err != nil {
		writeResponse(w, r, http.StatusInternalServerError, redisError(ctx, http.StatusInternalServerError, "failed to retrieve zone"), err)
		return
	}

	response := api.DeleteCounterResponse{
		Counter: 0,
		Zone:    zone,
	}

	writeResponse(w, r, http.StatusOK, response, nil)
}

func (s *ServerImpl) GetCounter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	counter, err := s.redisClient.Get(ctx, COUNTER_KEY).Result()
	if errors.Is(err, redis.Nil) {
		counter = "0"
	} else if err != nil {
		writeResponse(w, r, http.StatusInternalServerError, redisError(ctx, http.StatusInternalServerError, "failed to retrieve counter"), err)
		return
	}

	zone, err := s.redisClient.Get(ctx, ZONE_KEY).Result()
	if errors.Is(err, redis.Nil) {
		zone = ""
	} else if err != nil {
		writeResponse(w, r, http.StatusInternalServerError, redisError(ctx, http.StatusInternalServerError, "failed to retrieve zone"), err)
		return
	}
	c, _ := strconv.Atoi(counter)

	response := api.GetCounterResponse{
		Counter: c,
		Zone:    zone,
	}

	writeResponse(w, r, http.StatusOK, response, nil)
}

func (s *ServerImpl) PostCounter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	counter, err := s.redisClient.Incr(ctx, COUNTER_KEY).Result()
	if err != nil {
		writeResponse(w, r, http.StatusInternalServerError, redisError(ctx, http.StatusInternalServerError, "failed to increment counter"), err)
		return
	}

	zone, err := s.redisClient.Get(ctx, ZONE_KEY).Result()
	if errors.Is(redis.Nil, err) {
		zone = ""
	} else if err != nil {
		writeResponse(w, r, http.StatusInternalServerError, redisError(ctx, http.StatusInternalServerError, "failed to retrieve zone"), err)
		return
	}

	response := api.PostCounterResponse{
		Counter: int(counter),
		Zone:    zone,
	}

	writeResponse(w, r, http.StatusOK, response, nil)
}

func (s *ServerImpl) GetVersion(w http.ResponseWriter, r *http.Request) {
	writeResponse(w, r, http.StatusOK, api.VersionResponse{
		Version: s.version,
		Color:   s.color,
	}, nil)
}

func writeResponse(w http.ResponseWriter, r *http.Request, originalStatusCode int, response interface{}, err error) {
	statusCode := originalStatusCode
	if originalStatusCode/100 < 5 { // If the original status is 5xx ignore our override
		statusStr := r.Header.Get("x-set-response-status-code")
		if st, _ := strconv.Atoi(statusStr); st != 0 {
			originalStatusCode = st
		}
	}
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
