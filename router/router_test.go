package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

// RemoveTrailingSlash stands in for echo's pre-routing middleware, which
// rewrote the path rather than redirecting. These assert the rewrite, and that
// no redirect is produced.
func TestRemoveTrailingSlashRewritesPath(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"strips one trailing slash", "/api/tags/", "/api/tags"},
		{"strips repeated trailing slashes", "/api/tags///", "/api/tags"},
		{"leaves a path without one alone", "/api/tags", "/api/tags"},
		{"preserves root", "/", "/"},
		{"collapses an all-slash path back to root", "//", "/"},
		{"collapses a longer all-slash path back to root", "////", "/"},
		{"only touches the tail", "/api/articles/some-slug/", "/api/articles/some-slug"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var seen string
			h := RemoveTrailingSlash(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				seen = r.URL.Path
				w.WriteHeader(http.StatusOK)
			}))
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.in, nil))
			assert.Equal(t, tc.want, seen, "path handed to the router")
			assert.Equal(t, http.StatusOK, rec.Code, "must rewrite, never redirect")
		})
	}
}

// A route registered without a trailing slash must be reachable with one, and
// answer directly rather than with gin's 301.
func TestRemoveTrailingSlashReachesTheRoute(t *testing.T) {
	r := New()
	r.GET("/api/tags", func(c *gin.Context) { c.String(http.StatusOK, "tags") })

	rec := httptest.NewRecorder()
	RemoveTrailingSlash(r).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/tags/", nil))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "tags", rec.Body.String())
}

func TestValidatorRejectsAndAccepts(t *testing.T) {
	type payload struct {
		Email string `validate:"required,email"`
	}
	assert.Error(t, Validate(&payload{Email: "not-an-email"}))
	assert.NoError(t, Validate(&payload{Email: "user@example.com"}))
}

// echo answered unmatched routes and wrong methods with JSON; gin's defaults
// are plain text, which would have been the only non-JSON response the API
// produced.
func TestNoRouteAndNoMethodAnswerJSON(t *testing.T) {
	r := New()
	r.GET("/api/tags", func(c *gin.Context) { c.String(http.StatusOK, "tags") })

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/nope", nil))
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.JSONEq(t, `{"message":"Not Found"}`, rec.Body.String())
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")

	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/tags", nil))
	assert.Equal(t, http.StatusMethodNotAllowed, rec.Code)
	assert.JSONEq(t, `{"message":"Method Not Allowed"}`, rec.Body.String())
}
