package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServeHTTP(t *testing.T) {
	assert := assert.New(t)
	plugin := Plugin{}
	plugin.router = plugin.initRouter()
	plugin.setConfiguration((&configuration{}).normalized())
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/status", nil)
	r.Header.Set("Mattermost-User-ID", "test-user-id")

	plugin.ServeHTTP(nil, w, r)

	result := w.Result()
	assert.NotNil(result)
	defer func() { _ = result.Body.Close() }()
	var body map[string]any
	err := json.NewDecoder(result.Body).Decode(&body)
	assert.Nil(err)
	assert.Equal("ok", body["status"])
	assert.Equal(pluginID, body["plugin_id"])
}
