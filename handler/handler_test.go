package handler

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"encoding/json"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/xesina/golang-gin-realworld-example-app/article"
	"github.com/xesina/golang-gin-realworld-example-app/db"
	"github.com/xesina/golang-gin-realworld-example-app/model"
	"github.com/xesina/golang-gin-realworld-example-app/router"
	"github.com/xesina/golang-gin-realworld-example-app/store"
	"github.com/xesina/golang-gin-realworld-example-app/user"
)

var (
	d  *gorm.DB
	us user.Store
	as article.Store
	h  *Handler
	e  *gin.Engine
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	setup()
	code := m.Run()
	tearDown()
	os.Exit(code)
}

func authHeader(token string) string {
	return "Token " + token
}

// newContext builds a gin context bound to req/rec, standing in for echo's
// e.NewContext together with SetParamNames/SetParamValues. Route parameters
// are passed as alternating key/value arguments.
func newContext(rec *httptest.ResponseRecorder, req *http.Request, params ...string) *gin.Context {
	c := gin.CreateTestContextOnly(rec, e)
	c.Request = req
	for i := 0; i+1 < len(params); i += 2 {
		c.Params = append(c.Params, gin.Param{Key: params[i], Value: params[i+1]})
	}
	return c
}

// runWithJWT stands in for echo's middleware(handler)(c) composition: the
// middleware runs first, and the handler runs only if it did not abort.
func runWithJWT(c *gin.Context, mw gin.HandlerFunc, h gin.HandlerFunc) {
	mw(c)
	if !c.IsAborted() {
		h(c)
	}
}

func setup() {
	d = db.TestDB()
	db.AutoMigrate(d)
	us = store.NewUserStore(d)
	as = store.NewArticleStore(d)
	h = NewHandler(us, as)
	e = router.New()
	loadFixtures()
}

func tearDown() {
	_ = d.Close()
	if err := db.DropTestDB(); err != nil {
		log.Fatal(err)
	}
}

func responseMap(b []byte, key string) map[string]interface{} {
	var m map[string]interface{}
	json.Unmarshal(b, &m)
	return m[key].(map[string]interface{})
}

func loadFixtures() error {
	u1bio := "user1 bio"
	u1image := "http://realworld.io/user1.jpg"
	u1 := model.User{
		Username: "user1",
		Email:    "user1@realworld.io",
		Bio:      &u1bio,
		Image:    &u1image,
	}
	u1.Password, _ = u1.HashPassword("secret")
	if err := us.Create(&u1); err != nil {
		return err
	}

	u2bio := "user2 bio"
	u2image := "http://realworld.io/user2.jpg"
	u2 := model.User{
		Username: "user2",
		Email:    "user2@realworld.io",
		Bio:      &u2bio,
		Image:    &u2image,
	}
	u2.Password, _ = u2.HashPassword("secret")
	if err := us.Create(&u2); err != nil {
		return err
	}
	us.AddFollower(&u2, u1.ID)

	a := model.Article{
		Slug:        "article1-slug",
		Title:       "article1 title",
		Description: "article1 description",
		Body:        "article1 body",
		AuthorID:    1,
		Tags: []model.Tag{
			{
				Tag: "tag1",
			},
			{
				Tag: "tag2",
			},
		},
	}
	as.CreateArticle(&a)
	as.AddComment(&a, &model.Comment{
		Body:      "article1 comment1",
		ArticleID: 1,
		UserID:    1,
	})

	a2 := model.Article{
		Slug:        "article2-slug",
		Title:       "article2 title",
		Description: "article2 description",
		Body:        "article2 body",
		AuthorID:    2,
		Favorites: []model.User{
			u1,
		},
		Tags: []model.Tag{
			{
				Tag: "tag1",
			},
		},
	}
	as.CreateArticle(&a2)
	as.AddComment(&a2, &model.Comment{
		Body:      "article2 comment1 by user1",
		ArticleID: 2,
		UserID:    1,
	})
	as.AddFavorite(&a2, 1)

	return nil
}

// Every test in this package rebuilds the fixtures and then asserts against
// them, so a fixture that silently fails to load turns into a confusing
// failure somewhere else. This asserts the seeded state directly.
func TestFixturesLoaded(t *testing.T) {
	tearDown()
	setup()

	u1, err := us.GetByUsername("user1")
	assert.NoError(t, err)
	assert.NotNil(t, u1)
	assert.Equal(t, "user1@realworld.io", u1.Email)
	assert.Equal(t, "user1 bio", *u1.Bio)

	u2, err := us.GetByUsername("user2")
	assert.NoError(t, err)
	assert.NotNil(t, u2)

	following, err := us.IsFollower(u2.ID, u1.ID)
	assert.NoError(t, err)
	assert.True(t, following, "user1 must follow user2")

	notFollowing, err := us.IsFollower(u1.ID, u2.ID)
	assert.NoError(t, err)
	assert.False(t, notFollowing, "the follow must not be mutual")

	a1, err := as.GetBySlug("article1-slug")
	assert.NoError(t, err)
	assert.NotNil(t, a1)
	assert.Equal(t, u1.ID, a1.AuthorID)
	assert.Len(t, a1.Tags, 2)

	a2, err := as.GetBySlug("article2-slug")
	assert.NoError(t, err)
	assert.NotNil(t, a2)
	assert.Equal(t, u2.ID, a2.AuthorID)
	assert.Len(t, a2.Favorites, 1, "article2 is favorited by user1")

	comments, err := as.GetCommentsBySlug("article1-slug")
	assert.NoError(t, err)
	assert.Len(t, comments, 1)

	tags, err := as.ListTags()
	assert.NoError(t, err)
	assert.Len(t, tags, 2)
}
