package middleware

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/xesina/golang-gin-realworld-example-app/utils"
)

type (
	JWTConfig struct {
		Skipper    Skipper
		SigningKey interface{}
	}
	Skipper      func(c *gin.Context) bool
	jwtExtractor func(*gin.Context) (string, error)
)

var (
	ErrJWTMissing = errors.New("missing or malformed jwt")
	ErrJWTInvalid = errors.New("invalid or expired jwt")
)

func JWT(key interface{}) gin.HandlerFunc {
	c := JWTConfig{}
	c.SigningKey = key
	return JWTWithConfig(c)
}

func JWTWithConfig(config JWTConfig) gin.HandlerFunc {
	extractor := jwtFromHeader("Authorization", "Token")
	return func(c *gin.Context) {
		auth, err := extractor(c)
		if err != nil {
			if config.Skipper != nil {
				if config.Skipper(c) {
					c.Next()
					return
				}
			}
			utils.AbortJSON(c, http.StatusUnauthorized, utils.NewError(err))
			return
		}
		token, err := jwt.Parse(auth, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return config.SigningKey, nil
		})
		if err != nil {
			utils.AbortJSON(c, http.StatusForbidden, utils.NewError(ErrJWTInvalid))
			return
		}
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userID := uint(claims["id"].(float64))
			c.Set("user", userID)
			c.Next()
			return
		}
		utils.AbortJSON(c, http.StatusForbidden, utils.NewError(ErrJWTInvalid))
	}
}

// jwtFromHeader returns a `jwtExtractor` that extracts token from the request header.
func jwtFromHeader(header string, authScheme string) jwtExtractor {
	return func(c *gin.Context) (string, error) {
		auth := c.Request.Header.Get(header)
		l := len(authScheme)
		if len(auth) > l+1 && auth[:l] == authScheme {
			return auth[l+1:], nil
		}
		return "", ErrJWTMissing
	}
}
