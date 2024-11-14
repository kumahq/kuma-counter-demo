package base

import (
	"encoding/json"
	"github.com/kumahq/kuma-counter-demo/pkg/api"
	"maps"
	"net/http"
	"slices"
)

func (s *ServerImpl) guardKvApi(w http.ResponseWriter, r *http.Request, present bool) bool {
	if present && s.kvUrl == "" {
		Error(w, r, http.StatusBadRequest, api.KV_DISABLED, nil, "kv api url is not set")
		return true
	}
	if !present && s.kvUrl != "" {
		Error(w, r, http.StatusBadRequest, api.KV_DISABLED, nil, "kv api endpoints disabled")
		return true
	}
	return false
}

func (s *ServerImpl) KvList(w http.ResponseWriter, r *http.Request) {
	if s.guardKvApi(w, r, false) {
		return
	}
	s.Lock()
	keys := slices.Collect(maps.Keys(s.kv))
	s.Unlock()
	if keys == nil {
		keys = []string{}
	}

	response := api.KVListResponse{
		Keys: keys,
	}
	writeResponse(w, r, http.StatusOK, response, nil)
}

func (s *ServerImpl) KvDelete(w http.ResponseWriter, r *http.Request, key string) {
	if s.guardKvApi(w, r, false) {
		return
	}
	s.Lock()
	val, exists := s.kv[key]
	delete(s.kv, key)
	s.Unlock()
	if !exists {
		Error(w, r, http.StatusNotFound, api.KV_NOT_FOUND, nil, "no key %q", key)
		return
	}

	response := val
	writeResponse(w, r, http.StatusOK, response, nil)
}

func (s *ServerImpl) KvGet(w http.ResponseWriter, r *http.Request, key string) {
	if s.guardKvApi(w, r, false) {
		return
	}
	s.Lock()
	val, exists := s.kv[key]
	s.Unlock()
	if !exists {
		Error(w, r, http.StatusNotFound, api.KV_NOT_FOUND, nil, "no key %q", key)
		return
	}

	response := val
	writeResponse(w, r, http.StatusOK, response, nil)
}

func (s *ServerImpl) KvPost(w http.ResponseWriter, r *http.Request, key string) {
	if s.guardKvApi(w, r, false) {
		return
	}
	in := api.KVPostRequest{}
	err := json.NewDecoder(r.Body).Decode(&in)
	if err != nil {
		Error(w, r, http.StatusBadRequest, api.INVALID_JSON, err, "failed to parse: %v", err)
		return
	}
	s.Lock()
	defer s.Unlock()
	if in.Expect != nil && s.kv[key].Value != *in.Expect {
		Error(w, r, http.StatusConflict, api.KV_CONFLICT, nil, "CaS failed for key %q expect %q has %q", key, *in.Expect, s.kv[key].Value)
		return
	}
	out := api.KV{
		Value: in.Value,
	}
	s.kv[key] = out
	response := out
	writeResponse(w, r, http.StatusOK, response, nil)
}
