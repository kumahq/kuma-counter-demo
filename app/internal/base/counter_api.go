package base

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/kumahq/kuma-counter-demo/pkg/api"
	"net/http"
	"strconv"
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
	req, _ := http.NewRequestWithContext(ctx, http.MethodDelete, s.kvUrl+"/api/key-value/"+COUNTER_KEY, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		Error(w, r, http.StatusInternalServerError, api.INTERNAL_ERROR, err, "failed sending request")
		return
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		Error(w, r, res.StatusCode, api.KV_NOT_FOUND, err, "Key not found")
		return
	}
	zone, err := s.getKey(ctx, ZONE_KEY)
	if err != nil {
		Error(w, r, http.StatusInternalServerError, api.INTERNAL_ERROR, err, "failed to retrieve zone")
		return
	}

	response := api.DeleteCounterResponse{
		Counter: 0,
		Zone:    zone,
	}

	writeResponse(w, r, http.StatusOK, response, nil)
}

func (s *ServerImpl) getKey(ctx context.Context, key string) (string, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, s.kvUrl+"/api/key-value/"+key, nil)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound {
		return "", nil
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
		Error(w, r, http.StatusInternalServerError, api.INTERNAL_ERROR, err, "failed to retrieve zone")
		return
	}
	zone, err := s.getKey(ctx, ZONE_KEY)
	if err != nil {
		Error(w, r, http.StatusInternalServerError, api.INTERNAL_ERROR, err, "failed to retrieve zone")
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

	writeResponse(w, r, http.StatusOK, response, nil)
}

func (s *ServerImpl) PostCounter(w http.ResponseWriter, r *http.Request) {
	if s.guardKvApi(w, r, true) {
		return
	}
	ctx := r.Context()
	zone, err := s.getKey(ctx, ZONE_KEY)
	if err != nil {
		Error(w, r, http.StatusInternalServerError, api.INTERNAL_ERROR, err, "failed to retrieve zone")
		return
	}
	for i := 0; i < 5; i++ {
		if s.tryIncrementCounter(w, r, zone) {
			return
		}
	}
	Error(w, r, http.StatusConflict, api.KV_CONFLICT, nil, "out of retries without success")

}

func (s *ServerImpl) tryIncrementCounter(w http.ResponseWriter, r *http.Request, zone string) bool {
	ctx := r.Context()
	counter, err := s.getKey(ctx, COUNTER_KEY)
	if err != nil {
		s.logger.InfoContext(ctx, "failed to retrieve counter", "error", err)
		Error(w, r, http.StatusInternalServerError, api.INTERNAL_ERROR, err, "failed to retrieve zone")
		return true
	}
	c := 0
	if counter != "" {
		c, _ = strconv.Atoi(counter)
	}

	// Now let's update
	b, _ := json.Marshal(api.KVPostRequest{Value: strconv.Itoa(c + 1), Expect: &counter})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, s.kvUrl+"/api/key-value/"+COUNTER_KEY, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		Error(w, r, http.StatusInternalServerError, api.INTERNAL_ERROR, err, "failed sending request")
		return true
	}
	switch res.StatusCode {
	case http.StatusConflict:
		return false
	case http.StatusNotFound:
		Error(w, r, res.StatusCode, api.KV_NOT_FOUND, nil, "counter key not found")
		return true
	case http.StatusOK:
		counterResponse := api.KVPostResponse{}
		err = json.NewDecoder(res.Body).Decode(&counterResponse)
		if err != nil {
			Error(w, r, http.StatusInternalServerError, api.INTERNAL_ERROR, nil, "failed to parse counter response")
			return true
		}
		c, _ := strconv.Atoi(counterResponse.Value)
		response := api.PostCounterResponse{
			Counter: c,
			Zone:    zone,
		}
		writeResponse(w, r, http.StatusOK, response, nil)
		return true
	default:
		Error(w, r, res.StatusCode, api.INTERNAL_ERROR, nil, "failed sending request")
		return true
	}

}
