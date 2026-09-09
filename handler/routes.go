package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xesina/golang-gin-realworld-example-app/router/middleware"
	"github.com/xesina/golang-gin-realworld-example-app/utils"
)

func (h *Handler) Register(v1 *gin.RouterGroup) {
	jwtMiddleware := middleware.JWT(utils.JWTSecret)
	guestUsers := v1.Group("/users")
	guestUsers.POST("", h.SignUp)
	guestUsers.POST("/login", h.Login)

	user := v1.Group("/user", jwtMiddleware)
	user.GET("", h.CurrentUser)
	user.PUT("", h.UpdateUser)

	// reading a profile is optional-auth per the RealWorld spec; following and
	// unfollowing still require a token.
	profiles := v1.Group("/profiles", middleware.JWTWithConfig(
		middleware.JWTConfig{
			Skipper: func(c *gin.Context) bool {
				return c.Request.Method == http.MethodGet
			},
			SigningKey: utils.JWTSecret,
		},
	))
	profiles.GET("/:username", h.GetProfile)
	profiles.POST("/:username/follow", h.Follow)
	profiles.DELETE("/:username/follow", h.Unfollow)

	articles := v1.Group("/articles", middleware.JWTWithConfig(
		middleware.JWTConfig{
			Skipper: func(c *gin.Context) bool {
				if c.Request.Method == http.MethodGet && c.FullPath() != "/api/articles/feed" {
					return true
				}
				return false
			},
			SigningKey: utils.JWTSecret,
		},
	))
	articles.POST("", h.CreateArticle)
	articles.GET("/feed", h.Feed)
	articles.PUT("/:slug", h.UpdateArticle)
	articles.DELETE("/:slug", h.DeleteArticle)
	articles.POST("/:slug/comments", h.AddComment)
	articles.DELETE("/:slug/comments/:id", h.DeleteComment)
	articles.POST("/:slug/favorite", h.Favorite)
	articles.DELETE("/:slug/favorite", h.Unfavorite)
	articles.GET("", h.Articles)
	articles.GET("/:slug", h.GetArticle)
	articles.GET("/:slug/comments", h.GetComments)

	tags := v1.Group("/tags")
	tags.GET("", h.Tags)
}
