package utils

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// echo encoded responses with a streaming json encoder, which terminates each
// document with a newline, and labelled them with an upper case charset. gin
// marshals without the newline and lower cases the charset, so responses are
// rendered through this type instead of gin's own to keep both bytes identical
// to the pre-migration server.
type jsonRender struct {
	data interface{}
}

var jsonContentType = []string{"application/json; charset=UTF-8"}

func (r jsonRender) Render(w http.ResponseWriter) error {
	r.WriteContentType(w)
	b, err := json.Marshal(r.data)
	if err != nil {
		return err
	}
	_, err = w.Write(append(b, '\n'))
	return err
}

func (r jsonRender) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if len(header["Content-Type"]) == 0 {
		header["Content-Type"] = jsonContentType
	}
}

// JSON writes obj as the response body with the given status code.
func JSON(c *gin.Context, code int, obj interface{}) {
	c.Render(code, jsonRender{data: obj})
}

// AbortJSON is JSON, for a middleware that is refusing the request.
func AbortJSON(c *gin.Context, code int, obj interface{}) {
	c.Abort()
	JSON(c, code, obj)
}
