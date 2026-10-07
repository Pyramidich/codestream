package testutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

// RegisterAndLogin creates a user, logs in and returns access token and user ID.
func RegisterAndLogin(t *testing.T, server *httptest.Server) (accessToken, userID string) {
	t.Helper()

	registerBody := map[string]string{
		"email":        "test@example.com",
		"password":     "password123",
		"display_name": "Test User",
	}
	registerResp := postJSON(t, server, "/auth/register", registerBody, "")
	require.Equal(t, http.StatusCreated, registerResp.StatusCode)

	loginBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	loginResp := postJSON(t, server, "/auth/login", loginBody, "")
	require.Equal(t, http.StatusOK, loginResp.StatusCode)

	var result struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	decodeJSON(t, loginResp, &result)

	return result.AccessToken, ""
}

// CreateProject creates a project and returns its ID.
func CreateProject(t *testing.T, server *httptest.Server, token, name string) string {
	t.Helper()

	body := map[string]string{"name": name}
	resp := postJSON(t, server, "/projects", body, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]interface{}
	decodeJSON(t, resp, &result)

	id, ok := result["id"].(string)
	require.True(t, ok, "project id should be string")
	return id
}

// CreateFile creates a file and returns its ID.
func CreateFile(t *testing.T, server *httptest.Server, token, projectID, name, path, language string) string {
	t.Helper()

	body := map[string]string{
		"name":     name,
		"path":     path,
		"language": language,
	}
	resp := postJSON(t, server, fmt.Sprintf("/projects/%s/files", projectID), body, token)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var result map[string]interface{}
	decodeJSON(t, resp, &result)

	id, ok := result["id"].(string)
	require.True(t, ok, "file id should be string")
	return id
}

// ConnectWebSocket connects to the WebSocket endpoint.
func ConnectWebSocket(t *testing.T, server *httptest.Server, token, fileID string) *websocket.Conn {
	t.Helper()

	wsURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	wsURL.Scheme = "ws"
	wsURL.RawQuery = url.Values{
		"token":   {token},
		"file_id": {fileID},
	}.Encode()
	wsURL.Path = "/ws"

	conn, _, err := websocket.DefaultDialer.Dial(wsURL.String(), nil)
	require.NoError(t, err)
	return conn
}

func postJSON(t *testing.T, server *httptest.Server, path string, body map[string]string, token string) *http.Response {
	t.Helper()

	payload, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, server.URL+path, bytes.NewReader(payload))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func getJSON(t *testing.T, server *httptest.Server, path, token string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, server.URL+path, nil)
	require.NoError(t, err)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func decodeJSON(t *testing.T, resp *http.Response, target interface{}) {
	t.Helper()
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	require.NoError(t, json.Unmarshal(body, target))
}
