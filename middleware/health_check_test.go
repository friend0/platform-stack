package middleware

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

const mockResponse = "I fly drones!"

func MockHandler(c *gin.Context) {
	c.String(http.StatusOK, mockResponse)
}

func TestHealthCheckNotAvailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()

	r.Use(gin.Logger(), gin.Recovery())
	r.GET("/flights", MockHandler)

	assertRoutesNotHidden(t, r)
	assertHealthCheckNotAvailable(t, r)
}

func TestHealthCheckAvailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()

	r.Use(HealthCheck("/healthz"), gin.Logger(), gin.Recovery())
	r.GET("/flights", MockHandler)

	assertRoutesNotHidden(t, r)
	assertHealthCheckAvailable(t, r)
}

func assertRoutesNotHidden(t *testing.T, r *gin.Engine) {
	w, err := get(r, "/flights")

	assert.Nil(t, err)
	assert.Equal(t, w.Code, http.StatusOK)

	b, _ := ioutil.ReadAll(w.Result().Body)

	assert.Equal(t, string(b), mockResponse)
}

func assertHealthCheckAvailable(t *testing.T, r *gin.Engine) {
	w, err := get(r, "/healthz")

	assert.Nil(t, err)
	assert.Equal(t, w.Code, http.StatusOK)
}

func assertHealthCheckNotAvailable(t *testing.T, r *gin.Engine) {
	w, err := get(r, "/healthz")

	assert.Nil(t, err)
	assert.Equal(t, w.Code, http.StatusNotFound)
}

func get(r *gin.Engine, path string) (*httptest.ResponseRecorder, error) {
	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("Couldn't create GET request to %s: %v\n", path, err)
	}

	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	return w, nil
}
