package gin

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	// defaultBodyLimitKB matches fiber's default and is used when body_limit is
	// missing or not positive.
	defaultBodyLimitKB = 4096
	defaultBodyLimit   = defaultBodyLimitKB << 10
)

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

		// A body without Content-Length (chunked) is cut off while being read, so it
		// is rejected only once something reads it: NewContextRequest turns the error
		// into a 413 for the content types it parses, and a handler reading the body
		// itself gets *http.MaxBytesError. A chunked body of another content type
		// that nobody reads is never rejected, but it is never buffered either.
		//
		// MaxBytesReader reports the overrun to the server through the writer it is
		// given and the server only recognises its own, so unwrap gin's wrapper. Left
		// wrapped, the server drains the rest of the body before writing the response
		// and stalls on a client that stopped sending.
		c.Request.Body = http.MaxBytesReader(unwrapWriter(c.Writer), c.Request.Body, limit)
		c.Next()
	}
}

// unwrapWriter returns the http.ResponseWriter the given one wraps, if any.
func unwrapWriter(writer http.ResponseWriter) http.ResponseWriter {
	if unwrapper, ok := writer.(interface{ Unwrap() http.ResponseWriter }); ok {
		return unwrapper.Unwrap()
	}

	return writer
}

func abortBodyTooLarge(c *gin.Context) {
	c.Abort()
	c.String(http.StatusRequestEntityTooLarge, http.StatusText(http.StatusRequestEntityTooLarge))
}

func isBodyTooLarge(err error) bool {
	var maxBytesErr *http.MaxBytesError
	return errors.As(err, &maxBytesErr)
}
