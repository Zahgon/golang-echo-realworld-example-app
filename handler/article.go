package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xesina/golang-gin-realworld-example-app/model"
	"github.com/xesina/golang-gin-realworld-example-app/utils"
)

// GetArticle godoc
// @Summary Get an article
// @Description Get an article. Auth not required
// @ID get-article
// @Tags article
// @Accept  json
// @Produce  json
// @Param slug path string true "Slug of the article to get"
// @Success 200 {object} singleArticleResponse
// @Failure 400 {object} utils.Error
// @Failure 500 {object} utils.Error
// @Router /articles/{slug} [get]
func (h *Handler) GetArticle(c *gin.Context) {
	slug := c.Param("slug")
	a, err := h.articleStore.GetBySlug(slug)

	if err != nil {
		utils.JSON(c, http.StatusInternalServerError, utils.NewError(err))
		return
	}

	if a == nil {
		utils.JSON(c, http.StatusNotFound, utils.NotFound())
		return
	}

	utils.JSON(c, http.StatusOK, newArticleResponse(c, a))
	return
}

// Articles godoc
// @Summary Get recent articles globally
// @Description Get most recent articles globally. Use query parameters to filter results. Auth is optional
// @ID get-articles
// @Tags article
// @Accept  json
// @Produce  json
// @Param tag query string false "Filter by tag"
// @Param author query string false "Filter by author (username)"
// @Param favorited query string false "Filter by favorites of a user (username)"
// @Param limit query integer false "Limit number of articles returned (default is 20)"
// @Param offset query integer false "Offset/skip number of articles (default is 0)"
// @Success 200 {object} articleListResponse
// @Failure 500 {object} utils.Error
// @Router /articles [get]
func (h *Handler) Articles(c *gin.Context) {
	var (
		articles []model.Article
		count    int
	)

	tag := c.Query("tag")
	author := c.Query("author")
	favoritedBy := c.Query("favorited")

	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		offset = 0
	}

	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		limit = 20
	}

	if tag != "" {
		articles, count, err = h.articleStore.ListByTag(tag, offset, limit)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, nil)
			return
		}
	} else if author != "" {
		articles, count, err = h.articleStore.ListByAuthor(author, offset, limit)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, nil)
			return
		}
	} else if favoritedBy != "" {
		articles, count, err = h.articleStore.ListByWhoFavorited(favoritedBy, offset, limit)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, nil)
			return
		}
	} else {
		articles, count, err = h.articleStore.List(offset, limit)
		if err != nil {
			utils.JSON(c, http.StatusInternalServerError, nil)
			return
		}
	}

	utils.JSON(c, http.StatusOK, newArticleListResponse(h.userStore, userIDFromToken(c), articles, count))
	return
}

// Feed godoc
// @Summary Get recent articles from users you follow
// @Description Get most recent articles from users you follow. Use query parameters to limit. Auth is required
// @ID feed
// @Tags article
// @Accept  json
// @Produce  json
// @Param limit query integer false "Limit number of articles returned (default is 20)"
// @Param offset query integer false "Offset/skip number of articles (default is 0)"
// @Success 200 {object} articleListResponse
// @Failure 401 {object} utils.Error
// @Failure 500 {object} utils.Error
// @Security ApiKeyAuth
// @Router /articles/feed [get]
func (h *Handler) Feed(c *gin.Context) {
	var (
		articles []model.Article
		count    int
	)

	offset, err := strconv.Atoi(c.Query("offset"))
	if err != nil {
		offset = 0
	}

	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		limit = 20
	}

	articles, count, err = h.articleStore.ListFeed(userIDFromToken(c), offset, limit)
	if err != nil {
		utils.JSON(c, http.StatusInternalServerError, nil)
		return
	}

	utils.JSON(c, http.StatusOK, newArticleListResponse(h.userStore, userIDFromToken(c), articles, count))
	return
}

// CreateArticle godoc
// @Summary Create an article
// @Description Create an article. Auth is require
// @ID create-article
// @Tags article
// @Accept  json
// @Produce  json
// @Param article body articleCreateRequest true "Article to create"
// @Success 201 {object} singleArticleResponse
// @Failure 401 {object} utils.Error
// @Failure 422 {object} utils.Error
// @Failure 500 {object} utils.Error
// @Security ApiKeyAuth
// @Router /articles [post]
func (h *Handler) CreateArticle(c *gin.Context) {
	var a model.Article

	req := &articleCreateRequest{}
	if err := req.bind(c, &a); err != nil {
		utils.JSON(c, http.StatusUnprocessableEntity, utils.NewError(err))
		return
	}

	a.AuthorID = userIDFromToken(c)

	err := h.articleStore.CreateArticle(&a)
	if err != nil {
		utils.JSON(c, http.StatusUnprocessableEntity, utils.NewError(err))
		return
	}

	utils.JSON(c, http.StatusCreated, newArticleResponse(c, &a))
	return
}

// UpdateArticle godoc
// @Summary Update an article
// @Description Update an article. Auth is required
// @ID update-article
// @Tags article
// @Accept  json
// @Produce  json
// @Param slug path string true "Slug of the article to update"
// @Param article body articleUpdateRequest true "Article to update"
// @Success 200 {object} singleArticleResponse
// @Failure 400 {object} utils.Error
// @Failure 401 {object} utils.Error
// @Failure 422 {object} utils.Error
// @Failure 404 {object} utils.Error
// @Failure 500 {object} utils.Error
// @Security ApiKeyAuth
// @Router /articles/{slug} [put]
func (h *Handler) UpdateArticle(c *gin.Context) {
	slug := c.Param("slug")

	a, err := h.articleStore.GetUserArticleBySlug(userIDFromToken(c), slug)
	if err != nil {
		utils.JSON(c, http.StatusInternalServerError, utils.NewError(err))
		return
	}

	if a == nil {
		utils.JSON(c, http.StatusNotFound, utils.NotFound())
		return
	}

	req := &articleUpdateRequest{}
	req.populate(a)

	if err := req.bind(c, a); err != nil {
		utils.JSON(c, http.StatusUnprocessableEntity, utils.NewError(err))
		return
	}

	if err = h.articleStore.UpdateArticle(a, req.Article.Tags); err != nil {
		utils.JSON(c, http.StatusInternalServerError, utils.NewError(err))
		return
	}

	utils.JSON(c, http.StatusOK, newArticleResponse(c, a))
	return
}

// DeleteArticle godoc
// @Summary Delete an article
// @Description Delete an article. Auth is required
// @ID delete-article
// @Tags article
// @Accept  json
// @Produce  json
// @Param slug path string true "Slug of the article to delete"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} utils.Error
// @Failure 404 {object} utils.Error
// @Failure 500 {object} utils.Error
// @Security ApiKeyAuth
// @Router /articles/{slug} [delete]
func (h *Handler) DeleteArticle(c *gin.Context) {
	slug := c.Param("slug")

	a, err := h.articleStore.GetUserArticleBySlug(userIDFromToken(c), slug)
	if err != nil {
		utils.JSON(c, http.StatusInternalServerError, utils.NewError(err))
		return
	}

	if a == nil {
		utils.JSON(c, http.StatusNotFound, utils.NotFound())
		return
	}

	err = h.articleStore.DeleteArticle(a)
	if err != nil {
		utils.JSON(c, http.StatusInternalServerError, utils.NewError(err))
		return
	}

	utils.JSON(c, http.StatusOK, map[string]interface{}{"result": "ok"})
	return
}

// AddComment godoc
// @Summary Create a comment for an article
// @Description Create a comment for an article. Auth is required
// @ID add-comment
// @Tags comment
// @Accept  json
// @Produce  json
// @Param slug path string true "Slug of the article that you want to create a comment for"
// @Param comment body createCommentRequest true "Comment you want to create"
// @Success 201 {object} singleCommentResponse
// @Failure 400 {object} utils.Error
// @Failure 401 {object} utils.Error
// @Failure 422 {object} utils.Error
// @Failure 404 {object} utils.Error
// @Failure 500 {object} utils.Error
// @Security ApiKeyAuth
// @Router /articles/{slug}/comments [post]
func (h *Handler) AddComment(c *gin.Context) {
	slug := c.Param("slug")

	a, err := h.articleStore.GetBySlug(slug)
	if err != nil {
		utils.JSON(c, http.StatusInternalServerError, utils.NewError(err))
		return
	}

	if a == nil {
		utils.JSON(c, http.StatusNotFound, utils.NotFound())
		return
	}

	var cm model.Comment

	req := &createCommentRequest{}
	if err := req.bind(c, &cm); err != nil {
		utils.JSON(c, http.StatusUnprocessableEntity, utils.NewError(err))
		return
	}

	if err = h.articleStore.AddComment(a, &cm); err != nil {
		utils.JSON(c, http.StatusInternalServerError, utils.NewError(err))
		return
	}

	utils.JSON(c, http.StatusCreated, newCommentResponse(c, &cm))
	return
}

// GetComments godoc
// @Summary Get the comments for an article
// @Description Get the comments for an article. Auth is optional
// @ID get-comments
// @Tags comment
// @Accept  json
// @Produce  json
// @Param slug path string true "Slug of the article that you want to get comments for"
// @Success 200 {object} commentListResponse
// @Failure 422 {object} utils.Error
// @Failure 500 {object} utils.Error
// @Router /articles/{slug}/comments [get]
func (h *Handler) GetComments(c *gin.Context) {
	slug := c.Param("slug")

	cm, err := h.articleStore.GetCommentsBySlug(slug)
	if err != nil {
		utils.JSON(c, http.StatusInternalServerError, utils.NewError(err))
		return
	}

	utils.JSON(c, http.StatusOK, newCommentListResponse(c, cm))
	return
}

// DeleteComment godoc
// @Summary Delete a comment for an article
// @Description Delete a comment for an article. Auth is required
// @ID delete-comments
// @Tags comment
// @Accept  json
// @Produce  json
// @Param slug path string true "Slug of the article that you want to delete a comment for"
// @Param id path integer true "ID of the comment you want to delete"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} utils.Error
// @Failure 401 {object} utils.Error
// @Failure 422 {object} utils.Error
// @Failure 404 {object} utils.Error
// @Failure 500 {object} utils.Error
// @Security ApiKeyAuth
// @Router /articles/{slug}/comments/{id} [delete]
func (h *Handler) DeleteComment(c *gin.Context) {
	id64, err := strconv.ParseUint(c.Param("id"), 10, 32)
	id := uint(id64)

	if err != nil {
		utils.JSON(c, http.StatusBadRequest, utils.NewError(err))
		return
	}

	cm, err := h.articleStore.GetCommentByID(id)
	if err != nil {
		utils.JSON(c, http.StatusInternalServerError, utils.NewError(err))
		return
	}

	if cm == nil {
		utils.JSON(c, http.StatusNotFound, utils.NotFound())
		return
	}

	if cm.UserID != userIDFromToken(c) {
		utils.JSON(c, http.StatusUnauthorized, utils.NewError(errors.New("unauthorized action")))
		return
	}

	if err := h.articleStore.DeleteComment(cm); err != nil {
		utils.JSON(c, http.StatusInternalServerError, utils.NewError(err))
		return
	}

	utils.JSON(c, http.StatusOK, map[string]interface{}{"result": "ok"})
	return
}

// Favorite godoc
// @Summary Favorite an article
// @Description Favorite an article. Auth is required
// @ID favorite
// @Tags favorite
// @Accept  json
// @Produce  json
// @Param slug path string true "Slug of the article that you want to favorite"
// @Success 200 {object} singleArticleResponse
// @Failure 400 {object} utils.Error
// @Failure 401 {object} utils.Error
// @Failure 422 {object} utils.Error
// @Failure 404 {object} utils.Error
// @Failure 500 {object} utils.Error
// @Security ApiKeyAuth
// @Router /articles/{slug}/favorite [post]
func (h *Handler) Favorite(c *gin.Context) {
	slug := c.Param("slug")
	a, err := h.articleStore.GetBySlug(slug)

	if err != nil {
		utils.JSON(c, http.StatusInternalServerError, utils.NewError(err))
		return
	}

	if a == nil {
		utils.JSON(c, http.StatusNotFound, utils.NotFound())
		return
	}

	if err := h.articleStore.AddFavorite(a, userIDFromToken(c)); err != nil {
		utils.JSON(c, http.StatusUnprocessableEntity, utils.NewError(err))
		return
	}

	utils.JSON(c, http.StatusOK, newArticleResponse(c, a))
	return
}

// Unfavorite godoc
// @Summary Unfavorite an article
// @Description Unfavorite an article. Auth is required
// @ID unfavorite
// @Tags favorite
// @Accept  json
// @Produce  json
// @Param slug path string true "Slug of the article that you want to unfavorite"
// @Success 200 {object} singleArticleResponse
// @Failure 400 {object} utils.Error
// @Failure 401 {object} utils.Error
// @Failure 422 {object} utils.Error
// @Failure 404 {object} utils.Error
// @Failure 500 {object} utils.Error
// @Security ApiKeyAuth
// @Router /articles/{slug}/favorite [delete]
func (h *Handler) Unfavorite(c *gin.Context) {
	slug := c.Param("slug")

	a, err := h.articleStore.GetBySlug(slug)
	if err != nil {
		utils.JSON(c, http.StatusInternalServerError, utils.NewError(err))
		return
	}

	if a == nil {
		utils.JSON(c, http.StatusNotFound, utils.NotFound())
		return
	}

	if err := h.articleStore.RemoveFavorite(a, userIDFromToken(c)); err != nil {
		utils.JSON(c, http.StatusUnprocessableEntity, utils.NewError(err))
		return
	}

	utils.JSON(c, http.StatusOK, newArticleResponse(c, a))
	return
}

// Tags godoc
// @Summary Get tags
// @Description Get tags
// @ID tags
// @Tags tag
// @Accept  json
// @Produce  json
// @Success 201 {object} tagListResponse
// @Failure 400 {object} utils.Error
// @Failure 422 {object} utils.Error
// @Failure 404 {object} utils.Error
// @Failure 500 {object} utils.Error
// @Security ApiKeyAuth
// @Router /tags [get]
func (h *Handler) Tags(c *gin.Context) {
	tags, err := h.articleStore.ListTags()
	if err != nil {
		utils.JSON(c, http.StatusInternalServerError, err)
		return
	}

	utils.JSON(c, http.StatusOK, newTagListResponse(tags))
	return
}
