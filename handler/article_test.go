package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xesina/golang-gin-realworld-example-app/model"
	"github.com/xesina/golang-gin-realworld-example-app/router/middleware"
	"github.com/xesina/golang-gin-realworld-example-app/utils"
)

func TestListArticlesCaseSuccess(t *testing.T) {
	tearDown()
	setup()
	req := httptest.NewRequest(http.MethodGet, "/api/articles", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := newContext(rec, req)
	h.Articles(c)
	if assert.Equal(t, http.StatusOK, rec.Code) {
		var aa articleListResponse
		err := json.Unmarshal(rec.Body.Bytes(), &aa)
		assert.NoError(t, err)
		assert.Equal(t, 2, aa.ArticlesCount)
	}
}

func TestGetArticlesCaseSuccess(t *testing.T) {
	tearDown()
	setup()
	req := httptest.NewRequest(http.MethodGet, "/api/articles/:slug", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := newContext(rec, req, "slug", "article1-slug")
	h.GetArticle(c)
	if assert.Equal(t, http.StatusOK, rec.Code) {
		var a singleArticleResponse
		err := json.Unmarshal(rec.Body.Bytes(), &a)
		assert.NoError(t, err)
		assert.Equal(t, "article1-slug", a.Article.Slug)
		assert.Equal(t, 2, len(a.Article.TagList))
	}
}

func TestCreateArticlesCaseSuccess(t *testing.T) {
	tearDown()
	setup()
	var (
		reqJSON = `{"article":{"title":"article2", "description":"article2", "body":"article2", "tagList":["tag1","tag2"]}}`
	)
	jwtMiddleware := middleware.JWT(utils.JWTSecret)
	req := httptest.NewRequest(http.MethodPost, "/api/articles", strings.NewReader(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(utils.GenerateJWT(1)))
	rec := httptest.NewRecorder()
	c := newContext(rec, req)
	runWithJWT(c, jwtMiddleware, h.CreateArticle)
	if assert.Equal(t, http.StatusCreated, rec.Code) {
		var a singleArticleResponse
		err := json.Unmarshal(rec.Body.Bytes(), &a)
		assert.NoError(t, err)
		assert.Equal(t, "article2", a.Article.Slug)
		assert.Equal(t, "article2", a.Article.Description)
		assert.Equal(t, "article2", a.Article.Title)
		assert.Equal(t, "user1", a.Article.Author.Username)
		assert.Equal(t, 2, len(a.Article.TagList))
	}
}

func TestUpdateArticlesCaseSuccess(t *testing.T) {
	tearDown()
	setup()
	var (
		reqJSON = `{"article":{"title":"article1 part 2", "tagList":["tag3"]}}`
	)
	jwtMiddleware := middleware.JWT(utils.JWTSecret)
	req := httptest.NewRequest(http.MethodPut, "/api/articles/:slug", strings.NewReader(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(utils.GenerateJWT(1)))
	rec := httptest.NewRecorder()
	c := newContext(rec, req, "slug", "article1-slug")
	runWithJWT(c, jwtMiddleware, h.UpdateArticle)
	if assert.Equal(t, http.StatusOK, rec.Code) {
		var a singleArticleResponse
		err := json.Unmarshal(rec.Body.Bytes(), &a)
		assert.NoError(t, err)
		assert.Equal(t, "article1 part 2", a.Article.Title)
		assert.Equal(t, "article1-part-2", a.Article.Slug)
		assert.Equal(t, 1, len(a.Article.TagList))
		assert.Equal(t, "tag3", a.Article.TagList[0])
	}
}

func TestFeedCaseSuccess(t *testing.T) {
	tearDown()
	setup()
	jwtMiddleware := middleware.JWT(utils.JWTSecret)
	req := httptest.NewRequest(http.MethodGet, "/api/articles/feed", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(utils.GenerateJWT(1)))
	rec := httptest.NewRecorder()
	c := newContext(rec, req)
	runWithJWT(c, jwtMiddleware, h.Feed)
	if assert.Equal(t, http.StatusOK, rec.Code) {
		var a articleListResponse
		err := json.Unmarshal(rec.Body.Bytes(), &a)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(a.Articles))
		assert.Equal(t, a.ArticlesCount, len(a.Articles))
		assert.Equal(t, "article2 title", a.Articles[0].Title)
		assert.Equal(t, "article2 title", a.Articles[0].Title)
	}
}

func TestDeleteArticleCaseSuccess(t *testing.T) {
	tearDown()
	setup()
	jwtMiddleware := middleware.JWT(utils.JWTSecret)
	req := httptest.NewRequest(http.MethodDelete, "/api/articles/:slug", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(utils.GenerateJWT(1)))
	rec := httptest.NewRecorder()
	c := newContext(rec, req, "slug", "article1-slug")
	runWithJWT(c, jwtMiddleware, h.DeleteArticle)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestGetCommentsCaseSuccess(t *testing.T) {
	tearDown()
	setup()
	jwtMiddleware := middleware.JWT(utils.JWTSecret)
	req := httptest.NewRequest(http.MethodGet, "/api/articles/:slug/comments", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(utils.GenerateJWT(2)))
	rec := httptest.NewRecorder()
	c := newContext(rec, req, "slug", "article1-slug")
	runWithJWT(c, jwtMiddleware, h.GetComments)
	if assert.Equal(t, http.StatusOK, rec.Code) {
		var cc commentListResponse
		err := json.Unmarshal(rec.Body.Bytes(), &cc)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(cc.Comments))
	}
}

func TestAddCommentCaseSuccess(t *testing.T) {
	tearDown()
	setup()
	var (
		reqJSON = `{"comment":{"body":"article1 comment2 by user2"}}`
	)
	jwtMiddleware := middleware.JWT(utils.JWTSecret)
	req := httptest.NewRequest(http.MethodPost, "/api/articles/:slug/comments", strings.NewReader(reqJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(utils.GenerateJWT(2)))
	rec := httptest.NewRecorder()
	c := newContext(rec, req, "slug", "article1-slug")
	runWithJWT(c, jwtMiddleware, h.AddComment)
	if assert.Equal(t, http.StatusCreated, rec.Code) {
		var c singleCommentResponse
		err := json.Unmarshal(rec.Body.Bytes(), &c)
		assert.NoError(t, err)
		assert.Equal(t, "article1 comment2 by user2", c.Comment.Body)
		assert.Equal(t, "user2", c.Comment.Author.Username)
	}
}

func TestDeleteCommentCaseSuccess(t *testing.T) {
	tearDown()
	setup()
	jwtMiddleware := middleware.JWT(utils.JWTSecret)
	req := httptest.NewRequest(http.MethodDelete, "/api/articles/:slug/comments/:id", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(utils.GenerateJWT(1)))
	rec := httptest.NewRecorder()
	c := newContext(rec, req, "slug", "article1-slug", "id", "1")
	runWithJWT(c, jwtMiddleware, h.DeleteComment)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestFavoriteCaseSuccess(t *testing.T) {
	tearDown()
	setup()
	jwtMiddleware := middleware.JWT(utils.JWTSecret)
	req := httptest.NewRequest(http.MethodPost, "/api/articles/:slug/favorite", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(utils.GenerateJWT(2)))
	rec := httptest.NewRecorder()
	c := newContext(rec, req, "slug", "article1-slug")
	runWithJWT(c, jwtMiddleware, h.Favorite)
	if assert.Equal(t, http.StatusOK, rec.Code) {
		var a singleArticleResponse
		err := json.Unmarshal(rec.Body.Bytes(), &a)
		assert.NoError(t, err)
		assert.Equal(t, "article1 title", a.Article.Title)
		assert.True(t, a.Article.Favorited)
		assert.Equal(t, 1, a.Article.FavoritesCount)
	}
}

func TestUnfavoriteCaseSuccess(t *testing.T) {
	tearDown()
	setup()
	jwtMiddleware := middleware.JWT(utils.JWTSecret)
	req := httptest.NewRequest(http.MethodDelete, "/api/articles/:slug/favorite", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(utils.GenerateJWT(1)))
	rec := httptest.NewRecorder()
	c := newContext(rec, req, "slug", "article2-slug")
	runWithJWT(c, jwtMiddleware, h.Unfavorite)
	if assert.Equal(t, http.StatusOK, rec.Code) {
		var a singleArticleResponse
		err := json.Unmarshal(rec.Body.Bytes(), &a)
		assert.NoError(t, err)
		assert.Equal(t, "article2 title", a.Article.Title)
		assert.False(t, a.Article.Favorited)
		assert.Equal(t, 0, a.Article.FavoritesCount)
	}
}

func TestGetTagsCaseSuccess(t *testing.T) {
	tearDown()
	setup()
	req := httptest.NewRequest(http.MethodGet, "/api/tags", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := newContext(rec, req)
	h.Tags(c)
	if assert.Equal(t, http.StatusOK, rec.Code) {
		var tt tagListResponse
		err := json.Unmarshal(rec.Body.Bytes(), &tt)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(tt.Tags))
		assert.Contains(t, tt.Tags, "tag1")
		assert.Contains(t, tt.Tags, "tag2")
	}
}

// The fixtures give user1 one article of their own and one article from the
// user they follow, so the original feed count - which counted the caller's own
// articles - was indistinguishable from the correct one. Adding a second
// article by the caller separates them.
func TestFeedCountCountsFollowedAuthorsNotCaller(t *testing.T) {
	tearDown()
	setup()
	own := model.Article{
		Slug:        "user1-second-slug",
		Title:       "user1 second title",
		Description: "d",
		Body:        "b",
		AuthorID:    1,
	}
	assert.NoError(t, as.CreateArticle(&own))

	req := httptest.NewRequest(http.MethodGet, "/api/articles/feed", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader(utils.GenerateJWT(1)))
	rec := httptest.NewRecorder()
	c := newContext(rec, req)
	runWithJWT(c, middleware.JWT(utils.JWTSecret), h.Feed)

	if assert.Equal(t, http.StatusOK, rec.Code) {
		var a articleListResponse
		assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &a))
		assert.Equal(t, 1, len(a.Articles), "feed holds only the followed author's article")
		assert.Equal(t, 1, a.ArticlesCount, "count must match the feed, not the caller's own article count")
	}
}
