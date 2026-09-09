package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xesina/golang-gin-realworld-example-app/model"
	"github.com/xesina/golang-gin-realworld-example-app/utils"
)

// SignUp godoc
// @Summary Register a new user
// @Description Register a new user
// @ID sign-up
// @Tags user
// @Accept  json
// @Produce  json
// @Param user body userRegisterRequest true "User info for registration"
// @Success 201 {object} userResponse
// @Failure 400 {object} utils.Error
// @Failure 404 {object} utils.Error
// @Failure 500 {object} utils.Error
// @Router /users [post]
func (h *Handler) SignUp(c *gin.Context) {
	var u model.User
	req := &userRegisterRequest{}
	if err := req.bind(c, &u); err != nil {
		utils.JSON(c, http.StatusUnprocessableEntity, utils.NewError(err))
		return
	}
	if err := h.userStore.Create(&u); err != nil {
		utils.JSON(c, http.StatusUnprocessableEntity, utils.NewError(err))
		return
	}
	utils.JSON(c, http.StatusCreated, newUserResponse(&u))
	return
}

// Login godoc
// @Summary Login for existing user
// @Description Login for existing user
// @ID login
// @Tags user
// @Accept  json
// @Produce  json
// @Param user body userLoginRequest true "Credentials to use"
// @Success 200 {object} userResponse
// @Failure 400 {object} utils.Error
// @Failure 401 {object} utils.Error
// @Failure 422 {object} utils.Error
// @Failure 404 {object} utils.Error
// @Failure 500 {object} utils.Error
// @Router /users/login [post]
func (h *Handler) Login(c *gin.Context) {
	req := &userLoginRequest{}
	if err := req.bind(c); err != nil {
		utils.JSON(c, http.StatusUnprocessableEntity, utils.NewError(err))
		return
	}
	u, err := h.userStore.GetByEmail(req.User.Email)
	if err != nil {
		utils.JSON(c, http.StatusInternalServerError, utils.NewError(err))
		return
	}
	if u == nil {
		utils.JSON(c, http.StatusForbidden, utils.AccessForbidden())
		return
	}
	if !u.CheckPassword(req.User.Password) {
		utils.JSON(c, http.StatusForbidden, utils.AccessForbidden())
		return
	}
	utils.JSON(c, http.StatusOK, newUserResponse(u))
	return
}

// CurrentUser godoc
// @Summary Get the current user
// @Description Gets the currently logged-in user
// @ID current-user
// @Tags user
// @Accept  json
// @Produce  json
// @Success 200 {object} userResponse
// @Failure 400 {object} utils.Error
// @Failure 401 {object} utils.Error
// @Failure 422 {object} utils.Error
// @Failure 404 {object} utils.Error
// @Failure 500 {object} utils.Error
// @Security ApiKeyAuth
// @Router /user [get]
func (h *Handler) CurrentUser(c *gin.Context) {
	u, err := h.userStore.GetByID(userIDFromToken(c))
	if err != nil {
		utils.JSON(c, http.StatusInternalServerError, utils.NewError(err))
		return
	}
	if u == nil {
		utils.JSON(c, http.StatusNotFound, utils.NotFound())
		return
	}
	utils.JSON(c, http.StatusOK, newUserResponse(u))
	return
}

// UpdateUser godoc
// @Summary Update current user
// @Description Update user information for current user
// @ID update-user
// @Tags user
// @Accept  json
// @Produce  json
// @Param user body userUpdateRequest true "User details to update. At least **one** field is required."
// @Success 200 {object} userResponse
// @Failure 400 {object} utils.Error
// @Failure 401 {object} utils.Error
// @Failure 422 {object} utils.Error
// @Failure 404 {object} utils.Error
// @Failure 500 {object} utils.Error
// @Security ApiKeyAuth
// @Router /user [put]
func (h *Handler) UpdateUser(c *gin.Context) {
	u, err := h.userStore.GetByID(userIDFromToken(c))
	if err != nil {
		utils.JSON(c, http.StatusInternalServerError, utils.NewError(err))
		return
	}
	if u == nil {
		utils.JSON(c, http.StatusNotFound, utils.NotFound())
		return
	}
	req := newUserUpdateRequest()
	req.populate(u)
	if err := req.bind(c, u); err != nil {
		utils.JSON(c, http.StatusUnprocessableEntity, utils.NewError(err))
		return
	}
	if err := h.userStore.Update(u); err != nil {
		utils.JSON(c, http.StatusUnprocessableEntity, utils.NewError(err))
		return
	}
	utils.JSON(c, http.StatusOK, newUserResponse(u))
	return
}

// GetProfile godoc
// @Summary Get a profile
// @Description Get a profile of a user of the system. Auth is optional
// @ID get-profile
// @Tags profile
// @Accept  json
// @Produce  json
// @Param username path string true "Username of the profile to get"
// @Success 200 {object} userResponse
// @Failure 400 {object} utils.Error
// @Failure 401 {object} utils.Error
// @Failure 422 {object} utils.Error
// @Failure 404 {object} utils.Error
// @Failure 500 {object} utils.Error
// @Security ApiKeyAuth
// @Router /profiles/{username} [get]
func (h *Handler) GetProfile(c *gin.Context) {
	username := c.Param("username")
	u, err := h.userStore.GetByUsername(username)
	if err != nil {
		utils.JSON(c, http.StatusInternalServerError, utils.NewError(err))
		return
	}
	if u == nil {
		utils.JSON(c, http.StatusNotFound, utils.NotFound())
		return
	}
	utils.JSON(c, http.StatusOK, newProfileResponse(h.userStore, userIDFromToken(c), u))
	return
}

// Follow godoc
// @Summary Follow a user
// @Description Follow a user by username
// @ID follow
// @Tags follow
// @Accept  json
// @Produce  json
// @Param username path string true "Username of the profile you want to follow"
// @Success 200 {object} profileResponse
// @Failure 400 {object} utils.Error
// @Failure 401 {object} utils.Error
// @Failure 422 {object} utils.Error
// @Failure 404 {object} utils.Error
// @Failure 500 {object} utils.Error
// @Security ApiKeyAuth
// @Router /profiles/{username}/follow [post]
func (h *Handler) Follow(c *gin.Context) {
	followerID := userIDFromToken(c)
	username := c.Param("username")
	u, err := h.userStore.GetByUsername(username)
	if err != nil {
		utils.JSON(c, http.StatusInternalServerError, utils.NewError(err))
		return
	}
	if u == nil {
		utils.JSON(c, http.StatusNotFound, utils.NotFound())
		return
	}
	if err := h.userStore.AddFollower(u, followerID); err != nil {
		utils.JSON(c, http.StatusUnprocessableEntity, utils.NewError(err))
		return
	}
	utils.JSON(c, http.StatusOK, newProfileResponse(h.userStore, userIDFromToken(c), u))
	return
}

// Unfollow godoc
// @Summary Unfollow a user
// @Description Unfollow a user by username
// @ID unfollow
// @Tags follow
// @Accept  json
// @Produce  json
// @Param username path string true "Username of the profile you want to unfollow"
// @Success 201 {object} userResponse
// @Failure 400 {object} utils.Error
// @Failure 401 {object} utils.Error
// @Failure 422 {object} utils.Error
// @Failure 404 {object} utils.Error
// @Failure 500 {object} utils.Error
// @Security ApiKeyAuth
// @Router /profiles/{username}/follow [delete]
func (h *Handler) Unfollow(c *gin.Context) {
	followerID := userIDFromToken(c)
	username := c.Param("username")
	u, err := h.userStore.GetByUsername(username)
	if err != nil {
		utils.JSON(c, http.StatusInternalServerError, utils.NewError(err))
		return
	}
	if u == nil {
		utils.JSON(c, http.StatusNotFound, utils.NotFound())
		return
	}
	if err := h.userStore.RemoveFollower(u, followerID); err != nil {
		utils.JSON(c, http.StatusUnprocessableEntity, utils.NewError(err))
		return
	}
	utils.JSON(c, http.StatusOK, newProfileResponse(h.userStore, userIDFromToken(c), u))
	return
}

func userIDFromToken(c *gin.Context) uint {
	v, ok := c.Get("user")
	if !ok {
		return 0
	}
	id, ok := v.(uint)
	if !ok {
		return 0
	}
	return id
}
