package apikeys

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/machinable/machinable/config"
	"github.com/machinable/machinable/dsi/interfaces"
	"github.com/machinable/machinable/dsi/models"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.ReleaseMode)
	code := m.Run()
	os.Exit(code)
}

func TestUpdateKey(t *testing.T) {
	ds := &interfaces.MockProjectAPIKeysDatastore{}
	handler := New(ds, &config.AppConfig{})

	tables := []struct {
		name       string
		updateErr  error
		newKey     *NewProjectKey
		statusCode int
	}{
		{
			"success",
			nil,
			&NewProjectKey{Role: "admin"},
			200,
		},
		{
			"empty role",
			nil,
			&NewProjectKey{Role: ""},
			200,
		},
		{
			"validation error - invalid role",
			nil,
			&NewProjectKey{Role: "Batman"},
			400,
		},
		{
			"update error",
			errors.New("unexpected error"),
			&NewProjectKey{Role: "user"},
			500,
		},
	}

	for _, tt := range tables {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.Default()
			ds.UpdateAPIKeyFunc = func(projectID, keyId string, read, write bool, role string) error {
				return tt.updateErr
			}

			setRoutes(router, handler, ds, func(c *gin.Context) { c.Set("projectId", "testing") })
			w := httptest.NewRecorder()

			jsonStr, _ := json.Marshal(tt.newKey)
			req, _ := http.NewRequest("PUT", "/keys/a-key-id", bytes.NewBuffer(jsonStr))
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}

func TestAddKey(t *testing.T) {
	ds := &interfaces.MockProjectAPIKeysDatastore{}
	handler := New(ds, &config.AppConfig{})

	tables := []struct {
		name       string
		addErr     error
		newKey     *NewProjectKey
		statusCode int
	}{
		{
			"success",
			nil,
			&NewProjectKey{Key: "1b7dfd69-bd3f-46c7-ac94-10bbdec1d96b"},
			201,
		},
		{
			"validation error",
			nil,
			&NewProjectKey{Key: "invalid-uuid-key"},
			400,
		},
		{
			"validation error - empty key",
			nil,
			&NewProjectKey{Key: ""},
			400,
		},
		{
			"create error",
			errors.New("unexpected error"),
			&NewProjectKey{Key: "1b7dfd69-bd3f-46c7-ac94-10bbdec1d96b"},
			500,
		},
	}

	for _, tt := range tables {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.Default()
			ds.CreateAPIKeyFunc = func(projectID, hash, description string, read, write bool, role string) (*models.ProjectAPIKey, error) {
				return nil, tt.addErr
			}

			setRoutes(router, handler, ds, func(c *gin.Context) { c.Set("projectId", "testing") })
			w := httptest.NewRecorder()

			jsonStr, _ := json.Marshal(tt.newKey)
			req, _ := http.NewRequest("POST", "/keys/", bytes.NewBuffer(jsonStr))
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}

func TestListKeys(t *testing.T) {
	ds := &interfaces.MockProjectAPIKeysDatastore{}
	handler := New(ds, &config.AppConfig{})

	tables := []struct {
		name       string
		listErr    error
		apiKeys    []*models.ProjectAPIKey
		statusCode int
	}{
		{
			"success",
			nil,
			[]*models.ProjectAPIKey{},
			200,
		},
		{
			"error",
			errors.New("unexpected error"),
			nil,
			500,
		},
	}

	for _, tt := range tables {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.Default()
			pid := ""
			ds.ListAPIKeysFunc = func(projectID string) ([]*models.ProjectAPIKey, error) {
				pid = projectID
				return tt.apiKeys, tt.listErr
			}

			setRoutes(router, handler, ds, func(c *gin.Context) { c.Set("projectId", "testing") })
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("GET", "/keys/", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, "testing", pid)
			assert.Equal(t, tt.statusCode, w.Code)

			if w.Code == 200 {
				respBody := struct {
					Items []*models.ProjectAPIKey `json:"items"`
				}{}

				json.Unmarshal(w.Body.Bytes(), &respBody)
				assert.Equal(t, tt.apiKeys, respBody.Items)
			}
		})
	}
}

func TestGenerate(t *testing.T) {
	router := gin.Default()
	ds := &interfaces.MockProjectAPIKeysDatastore{}
	handler := New(ds, &config.AppConfig{})

	setRoutes(router, handler, ds)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/keys/generate", nil)
	router.ServeHTTP(w, req)

	// check response code
	assert.Equal(t, 200, w.Code)

	// get response body
	response := struct {
		Key string `json:"key"`
	}{}

	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Error("failed parsing response")
	}

	// verify generated key
	_, err := uuid.FromString(response.Key)
	if err != nil {
		t.Error("invalid key")
	}
}

func TestDeleteKey(t *testing.T) {
	ds := &interfaces.MockProjectAPIKeysDatastore{}
	handler := New(ds, &config.AppConfig{})

	tables := []struct {
		name       string
		deleteErr  error
		statusCode int
	}{
		{
			"success",
			nil,
			204,
		},
		{
			"error",
			errors.New("unexpected error"),
			500,
		},
	}

	for _, tt := range tables {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.Default()
			pid, kid := "", ""
			ds.DeleteAPIKeyFunc = func(projectID, keyID string) error {
				pid = projectID
				kid = keyID
				return tt.deleteErr
			}

			setRoutes(router, handler, ds, func(c *gin.Context) { c.Set("projectId", "testing") })
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("DELETE", "/keys/first-api-key", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, "testing", pid)
			assert.Equal(t, "first-api-key", kid)
			assert.Equal(t, tt.statusCode, w.Code)
		})
	}
}
