package router

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/xesina/golang-gin-realworld-example-app/utils"
)

func New() *gin.Engine {
	r := gin.New()
	// Echo rewrote the path before routing, so a trailing slash never produced
	// a redirect. RemoveTrailingSlash below reproduces that, which means gin's
	// own redirecting behaviour has to be switched off.
	r.RedirectTrailingSlash = false
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	// echo's CORS middleware set Vary: Origin on every response; gin-contrib
	// omits it when all origins are allowed, and adds its own Vary list on a
	// preflight, so it is set here for everything except preflights.
	r.Use(func(c *gin.Context) {
		if c.Request.Method != http.MethodOptions {
			c.Writer.Header().Set("Vary", "Origin")
		}
		c.Next()
	})
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowHeaders:    []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodPut,
			http.MethodPatch,
			http.MethodPost,
			http.MethodDelete,
		},
	}))
	// echo answered an unmatched route with JSON, and every other response this
	// API produces is JSON; gin's defaults are plain text.
	r.HandleMethodNotAllowed = true
	r.NoRoute(func(c *gin.Context) {
		utils.JSON(c, http.StatusNotFound, gin.H{"message": "Not Found"})
	})
	r.NoMethod(func(c *gin.Context) {
		utils.JSON(c, http.StatusMethodNotAllowed, gin.H{"message": "Method Not Allowed"})
	})

	return r
}

// RemoveTrailingSlash strips a trailing slash from the request path before the
// engine routes it, the way echo's RemoveTrailingSlash pre-middleware did.
func RemoveTrailingSlash(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p := r.URL.Path; len(p) > 1 && strings.HasSuffix(p, "/") {
			trimmed := strings.TrimRight(p, "/")
			if trimmed == "" {
				trimmed = "/"
			}
			r.URL.Path = trimmed
		}
		h.ServeHTTP(w, r)
	})
}
