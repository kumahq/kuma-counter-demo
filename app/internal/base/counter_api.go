package base

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/kumahq/kuma-counter-demo/pkg/api"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// This API uses 2 entry in the KV COUNTER and ZONE
const (
	COUNTER_KEY = "counter"
	ZONE_KEY    = "zone"
)

func (s *ServerImpl) DeleteCounter(w http.ResponseWriter, r *http.Request) {
	if s.guardKvApi(w, r, true) {
		return
	}
	ctx := r.Context()
	path, _ := url.JoinPath(s.kvUrl, "/api/key-value", COUNTER_KEY)
	req, _ := http.NewRequestWithContext(ctx, http.MethodDelete, path, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		s.writeErrorResponse(w, r, http.StatusInternalServerError, api.INTERNAL_ERROR, err, "failed sending request")
		return
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		s.writeErrorResponse(w, r, res.StatusCode, api.KV_NOT_FOUND, err, "Key not found")
		return
	}
	zone, err := s.getKey(ctx, ZONE_KEY)
	if err != nil {
		s.writeErrorResponse(w, r, http.StatusInternalServerError, api.INTERNAL_ERROR, err, "failed to retrieve zone")
		return
	}

	response := api.DeleteCounterResponse{
		Counter: 0,
		Zone:    zone,
	}

	s.writeResponse(w, r, http.StatusOK, response, nil)
}

func (s *ServerImpl) getKey(ctx context.Context, key string) (string, error) {
	path, _ := url.JoinPath(s.kvUrl, "/api/key-value", key)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return "", nil
	}
	if res.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("request to kv failed with statusCode: %q body: %q", res.Status, b)
	}
	zoneResponse := api.KVGetResponse{}
	err = json.NewDecoder(res.Body).Decode(&zoneResponse)
	if err != nil {
		return "", err
	}
	return zoneResponse.Value, nil
}

func (s *ServerImpl) GetCounter(w http.ResponseWriter, r *http.Request) {
	if s.guardKvApi(w, r, true) {
		return
	}
	ctx := r.Context()
	counter, err := s.getKey(ctx, COUNTER_KEY)
	if err != nil {
		s.writeErrorResponse(w, r, http.StatusInternalServerError, api.INTERNAL_ERROR, err, "failed to retrieve zone")
		return
	}
	zone, err := s.getKey(ctx, ZONE_KEY)
	if err != nil {
		s.writeErrorResponse(w, r, http.StatusInternalServerError, api.INTERNAL_ERROR, err, "failed to retrieve zone")
		return
	}
	c := 0
	if counter != "" {
		c, _ = strconv.Atoi(counter)
	}

	response := api.GetCounterResponse{
		Counter: c,
		Zone:    zone,
	}

	s.writeResponse(w, r, http.StatusOK, response, nil)
}

func (s *ServerImpl) PostCounter(w http.ResponseWriter, r *http.Request) {
	if s.guardKvApi(w, r, true) {
		return
	}
	ctx := r.Context()
	zone, err := s.getKey(ctx, ZONE_KEY)
	if err != nil {
		s.writeErrorResponse(w, r, http.StatusInternalServerError, api.INTERNAL_ERROR, err, "failed to retrieve zone")
		return
	}
	for i := 0; ; i++ {
		if s.tryIncrementCounter(w, r, zone) {
			return
		}
		if i == 5 {
			s.writeErrorResponse(w, r, http.StatusConflict, api.KV_CONFLICT, nil, "out of retries without success")
			return
		}
		time.Sleep(time.Duration(int64(rand.Intn(50)+50)) * time.Millisecond)
	}

}

func (s *ServerImpl) tryIncrementCounter(w http.ResponseWriter, r *http.Request, zone string) bool {
	ctx := r.Context()
	counter, err := s.getKey(ctx, COUNTER_KEY)
	if err != nil {
		s.logger.InfoContext(ctx, "failed to retrieve counter", "error", err)
		s.writeErrorResponse(w, r, http.StatusInternalServerError, api.INTERNAL_ERROR, err, "failed to retrieve zone")
		return true
	}
	c := 0
	if counter != "" {
		c, _ = strconv.Atoi(counter)
	}

	// Now let's update
	b, _ := json.Marshal(api.KVPostRequest{Value: strconv.Itoa(c + 1), Expect: &counter})
	path, _ := url.JoinPath(s.kvUrl, "/api/key-value", COUNTER_KEY)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		s.writeErrorResponse(w, r, http.StatusInternalServerError, api.INTERNAL_ERROR, err, "failed sending request")
		return true
	}
	switch res.StatusCode {
	case http.StatusOK:
		counterResponse := api.KVPostResponse{}
		err = json.NewDecoder(res.Body).Decode(&counterResponse)
		if err != nil {
			s.writeErrorResponse(w, r, http.StatusInternalServerError, api.INTERNAL_ERROR, nil, "failed to parse counter response")
			return true
		}
		c, _ := strconv.Atoi(counterResponse.Value)
		response := api.PostCounterResponse{
			Counter: c,
			Zone:    zone,
		}
		s.writeResponse(w, r, http.StatusOK, response, nil)
		return true
	case http.StatusConflict:
		return false
	case http.StatusNotFound:
		s.writeErrorResponse(w, r, res.StatusCode, api.KV_NOT_FOUND, nil, "counter key not found")
		return true
	default:
		s.writeErrorResponse(w, r, res.StatusCode, api.INTERNAL_ERROR, nil, "failed sending request")
		return true
	}

}
