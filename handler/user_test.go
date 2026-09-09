package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xesina/golang-gin-realworld-example-app/router"
	"github.com/xesina/golang-gin-realworld-example-app/router/middleware"
	"github.com/xesina/golang-gin-realworld-example-app/utils"
)

func TestSignUpCaseSuccess(t *testing.T) {
	tearDown()
	setup()
	var (
		reqJSON = `{"user":{"username":"alice","email":"alice@realworld.io","password":"secret"}}`
	)
	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := newContext(rec, req)
	h.SignUp(c)
	if assert.Equal(t, http.StatusCreated, rec.Code) {
		m := responseMap(rec.Body.Bytes(), "user")
		assert.Equal(t, "alice", m["username"])
		assert.Equal(t, "alice@realworld.io", m["email"])
		assert.Nil(t, m["bio"])
		assert.Nil(t, m["image"])
		assert.NotEmpty(t, m["token"])
	}
}

func TestLoginCaseSuccess(t *testing.T) {
	tearDown()
	setup()
	var (
		reqJSON = `{"user":{"email":"user1@realworld.io","password":"secret"}}`
	)
	req := httptest.NewRequest(http.MethodPost, "/api/users/login", strings.NewReader(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := newContext(rec, req)
	h.Login(c)
	if assert.Equal(t, http.StatusOK, rec.Code) {
		m := responseMap(rec.Body.Bytes(), "user")
		assert.Equal(t, "user1", m["username"])
		assert.Equal(t, "user1@realworld.io", m["email"])
		assert.NotEmpty(t, m["token"])
	}
}

func TestLoginCaseFailed(t *testing.T) {
	tearDown()
	setup()
	var (
		reqJSON = `{"user":{"email":"userx@realworld.io","password":"secret"}}`
	)
	req := httptest.NewRequest(http.MethodPost, "/api/users/login", strings.NewReader(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := newContext(rec, req)
	h.Login(c)
	assert.Equal(t, http.StatusForbidden, rec.Code)
}

func TestCurrentUserCaseSuccess(t *testing.T) {
	tearDown()
	setup()
	jwtMiddleware := middleware.JWT(utils.JWTSecret)
	req := httptest.NewRequest(http.MethodGet, "/api/users/login", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(utils.GenerateJWT(1)))
	rec := httptest.NewRecorder()
	c := newContext(rec, req)
	runWithJWT(c, jwtMiddleware, h.CurrentUser)
	if assert.Equal(t, http.StatusOK, rec.Code) {
		m := responseMap(rec.Body.Bytes(), "user")
		assert.Equal(t, "user1", m["username"])
		assert.Equal(t, "user1@realworld.io", m["email"])
		assert.NotEmpty(t, m["token"])
	}
}

func TestCurrentUserCaseInvalid(t *testing.T) {
	tearDown()
	setup()
	jwtMiddleware := middleware.JWT(utils.JWTSecret)
	req := httptest.NewRequest(http.MethodGet, "/api/users/login", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(utils.GenerateJWT(100)))
	rec := httptest.NewRecorder()
	c := newContext(rec, req)
	runWithJWT(c, jwtMiddleware, h.CurrentUser)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestUpdateUserEmail(t *testing.T) {
	tearDown()
	setup()
	var (
		user1UpdateReq = `{"user":{"email":"user1@user1.me"}}`
	)
	jwtMiddleware := middleware.JWT(utils.JWTSecret)
	req := httptest.NewRequest(http.MethodPut, "/api/user", strings.NewReader(user1UpdateReq))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(utils.GenerateJWT(1)))
	rec := httptest.NewRecorder()
	c := newContext(rec, req)
	runWithJWT(c, jwtMiddleware, h.UpdateUser)
	if assert.Equal(t, http.StatusOK, rec.Code) {
		m := responseMap(rec.Body.Bytes(), "user")
		assert.Equal(t, "user1", m["username"])
		assert.Equal(t, "user1@user1.me", m["email"])
		assert.NotEmpty(t, m["token"])
	}
}

func TestUpdateUserMultipleFields(t *testing.T) {
	tearDown()
	setup()
	var (
		user1UpdateReq = `{"user":{"username":"user11","email":"user11@user11.me","bio":"user11 bio"}}`
	)
	jwtMiddleware := middleware.JWT(utils.JWTSecret)
	req := httptest.NewRequest(http.MethodPut, "/api/user", strings.NewReader(user1UpdateReq))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(utils.GenerateJWT(1)))
	rec := httptest.NewRecorder()
	c := newContext(rec, req)
	runWithJWT(c, jwtMiddleware, h.UpdateUser)
	if assert.Equal(t, http.StatusOK, rec.Code) {
		m := responseMap(rec.Body.Bytes(), "user")
		assert.Equal(t, "user11", m["username"])
		assert.Equal(t, "user11@user11.me", m["email"])
		assert.Equal(t, "user11 bio", m["bio"])
		assert.NotEmpty(t, m["token"])
	}
}

func TestGetProfileCaseSuccess(t *testing.T) {
	tearDown()
	setup()
	jwtMiddleware := middleware.JWT(utils.JWTSecret)
	req := httptest.NewRequest(http.MethodGet, "/api/profiles/:username", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(utils.GenerateJWT(1)))
	rec := httptest.NewRecorder()
	c := newContext(rec, req, "username", "user1")
	runWithJWT(c, jwtMiddleware, h.GetProfile)
	if assert.Equal(t, http.StatusOK, rec.Code) {
		m := responseMap(rec.Body.Bytes(), "profile")
		assert.Equal(t, "user1", m["username"])
		assert.Equal(t, "user1 bio", m["bio"])
		assert.Equal(t, "http://realworld.io/user1.jpg", m["image"])
		assert.Equal(t, false, m["following"])
	}
}

func TestGetProfileCaseNotFound(t *testing.T) {
	tearDown()
	setup()
	jwtMiddleware := middleware.JWT(utils.JWTSecret)
	req := httptest.NewRequest(http.MethodGet, "/api/profiles/:username", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(utils.GenerateJWT(1)))
	rec := httptest.NewRecorder()
	c := newContext(rec, req, "username", "userx")
	runWithJWT(c, jwtMiddleware, h.GetProfile)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestFollowCaseSuccess(t *testing.T) {
	tearDown()
	setup()
	jwtMiddleware := middleware.JWT(utils.JWTSecret)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(utils.GenerateJWT(1)))
	rec := httptest.NewRecorder()
	c := newContext(rec, req, "username", "user2")
	runWithJWT(c, jwtMiddleware, h.Follow)
	if assert.Equal(t, http.StatusOK, rec.Code) {
		m := responseMap(rec.Body.Bytes(), "profile")
		assert.Equal(t, "user2", m["username"])
		assert.Equal(t, "user2 bio", m["bio"])
		assert.Equal(t, "http://realworld.io/user2.jpg", m["image"])
		assert.Equal(t, true, m["following"])
	}
}

func TestFollowCaseInvalidUser(t *testing.T) {
	tearDown()
	setup()
	jwtMiddleware := middleware.JWT(utils.JWTSecret)
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(utils.GenerateJWT(1)))
	rec := httptest.NewRecorder()
	c := newContext(rec, req, "username", "userx")
	runWithJWT(c, jwtMiddleware, h.Follow)
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestUnfollow(t *testing.T) {
	tearDown()
	setup()
	jwtMiddleware := middleware.JWT(utils.JWTSecret)
	req := httptest.NewRequest(http.MethodDelete, "/api/profiles/:username/follow", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(utils.GenerateJWT(1)))
	rec := httptest.NewRecorder()
	c := newContext(rec, req, "username", "user2")
	runWithJWT(c, jwtMiddleware, h.Unfollow)
	if assert.Equal(t, http.StatusOK, rec.Code) {
		m := responseMap(rec.Body.Bytes(), "profile")
		assert.Equal(t, "user2", m["username"])
		assert.Equal(t, "user2 bio", m["bio"])
		assert.Equal(t, "http://realworld.io/user2.jpg", m["image"])
		assert.Equal(t, false, m["following"])
	}
}

// Reading a profile is optional-auth per the RealWorld spec. This drives the
// real route registration so the group's middleware wiring is what is tested.
func TestGetProfileWithoutAuth(t *testing.T) {
	tearDown()
	setup()
	r := router.New()
	h.Register(r.Group("/api"))

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/profiles/user1", nil))

	if assert.Equal(t, http.StatusOK, rec.Code) {
		m := responseMap(rec.Body.Bytes(), "profile")
		assert.Equal(t, "user1", m["username"])
		assert.Equal(t, false, m["following"], "an anonymous caller follows nobody")
	}
}

// ...while following still requires a token.
func TestFollowWithoutAuthIsRejected(t *testing.T) {
	tearDown()
	setup()
	r := router.New()
	h.Register(r.Group("/api"))

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/profiles/user2/follow", nil))
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
