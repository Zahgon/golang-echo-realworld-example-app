package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/xesina/golang-gin-realworld-example-app/utils"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

// run drives mw with the given Authorization header and method, reporting
// whether the chain was allowed to continue and what the user id ended up as.
func run(t *testing.T, mw gin.HandlerFunc, method, authHeader string) (*httptest.ResponseRecorder, bool, uint) {
	t.Helper()
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(method, "/api/user", nil)
	if authHeader != "" {
		c.Request.Header.Set("Authorization", authHeader)
	}
	mw(c)
	var id uint
	if v, ok := c.Get("user"); ok {
		id, _ = v.(uint)
	}
	return rec, !c.IsAborted(), id
}

func errorBody(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var e struct {
		Errors struct {
			Body string `json:"body"`
		} `json:"errors"`
	}
	assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &e))
	return e.Errors.Body
}

func TestJWTMissingToken(t *testing.T) {
	rec, proceeded, _ := run(t, JWT(utils.JWTSecret), http.MethodGet, "")
	assert.False(t, proceeded, "chain must be aborted when no token is present")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Equal(t, "missing or malformed jwt", errorBody(t, rec))
}

func TestJWTMalformedScheme(t *testing.T) {
	// the extractor only accepts the "Token" scheme
	rec, proceeded, _ := run(t, JWT(utils.JWTSecret), http.MethodGet, "Bearer "+utils.GenerateJWT(1))
	assert.False(t, proceeded)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestJWTInvalidToken(t *testing.T) {
	rec, proceeded, _ := run(t, JWT(utils.JWTSecret), http.MethodGet, "Token not-a-jwt-at-all")
	assert.False(t, proceeded, "chain must be aborted for an unparseable token")
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Equal(t, "invalid or expired jwt", errorBody(t, rec))
}

func TestJWTWrongSigningKey(t *testing.T) {
	// a well formed token signed by somebody else must not be accepted
	rec, proceeded, _ := run(t, JWTWithConfig(JWTConfig{SigningKey: []byte("not-the-secret")}),
		http.MethodGet, "Token "+utils.GenerateJWT(1))
	assert.False(t, proceeded)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestJWTValidTokenSetsUser(t *testing.T) {
	rec, proceeded, id := run(t, JWT(utils.JWTSecret), http.MethodGet, "Token "+utils.GenerateJWT(42))
	assert.True(t, proceeded, "chain must continue for a valid token")
	assert.Equal(t, uint(42), id, "the claim's user id must reach the handler")
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestJWTSkipperAllowsAnonymous(t *testing.T) {
	mw := JWTWithConfig(JWTConfig{
		Skipper:    func(c *gin.Context) bool { return c.Request.Method == http.MethodGet },
		SigningKey: utils.JWTSecret,
	})
	_, proceeded, id := run(t, mw, http.MethodGet, "")
	assert.True(t, proceeded, "skipper must let an anonymous GET through")
	assert.Equal(t, uint(0), id, "an anonymous request carries no user id")
}

func TestJWTSkipperStillRejectsOtherMethods(t *testing.T) {
	mw := JWTWithConfig(JWTConfig{
		Skipper:    func(c *gin.Context) bool { return c.Request.Method == http.MethodGet },
		SigningKey: utils.JWTSecret,
	})
	rec, proceeded, _ := run(t, mw, http.MethodPost, "")
	assert.False(t, proceeded, "a skipper for GET must not exempt POST")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestJWTSkipperNotConsultedWhenTokenPresent(t *testing.T) {
	// the skipper only governs the missing-token case; a bad token is still bad
	mw := JWTWithConfig(JWTConfig{
		Skipper:    func(c *gin.Context) bool { return true },
		SigningKey: utils.JWTSecret,
	})
	rec, proceeded, _ := run(t, mw, http.MethodGet, "Token garbage")
	assert.False(t, proceeded)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}
