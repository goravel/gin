package gin

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// defaultBodyLimit matches fiber's default and is used when body_limit is not positive.
const defaultBodyLimit = 4096 << 10

// bodyLimitMiddleware rejects request bodies larger than limit bytes with 413.
// It has to run before anything reads the body, including the eager parsing in
// NewContextRequest, so it is registered as a plain gin handler ahead of the
// Goravel middleware.
func bodyLimitMiddleware(limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body == nil || c.Request.Body == http.NoBody {
			c.Next()
			return
		}

		if c.Request.ContentLength > limit {
			abortBodyTooLarge(c)
			return
		}

		// A body without Content-Length (chunked) is cut off while being read.
		// NewContextRequest turns that error into a 413, handlers that read the
		// body themselves get *http.MaxBytesError.
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		c.Next()
	}
}

func abortBodyTooLarge(c *gin.Context) {
	c.Abort()
	c.String(http.StatusRequestEntityTooLarge, http.StatusText(http.StatusRequestEntityTooLarge))
}

func isBodyTooLarge(err error) bool {
	var maxBytesErr *http.MaxBytesError
	return errors.As(err, &maxBytesErr)
}
