package handler

import (
	"github.com/xesina/golang-gin-realworld-example-app/article"
	"github.com/xesina/golang-gin-realworld-example-app/user"
)

type Handler struct {
	userStore    user.Store
	articleStore article.Store
}

func NewHandler(us user.Store, as article.Store) *Handler {
	return &Handler{
		userStore:    us,
		articleStore: as,
	}
}
