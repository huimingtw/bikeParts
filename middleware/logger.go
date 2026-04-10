package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

const XRequestIDKey = "X-Request-Id"

func generateRequestID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func Logger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(XRequestIDKey)
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Header(XRequestIDKey, requestID)

		start := time.Now()
		c.Next()
		latency := time.Since(start)

		path := c.Request.URL.Path
		if raw := c.Request.URL.RawQuery; raw != "" {
			path = path + "?" + raw
		}

		status := c.Writer.Status()
		attrs := []any{
			"request_id", requestID,
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"ip", c.ClientIP(),
			"latency", latency.String(),
			"size", c.Writer.Size(),
		}
		if errs := c.Errors.ByType(gin.ErrorTypePrivate).String(); errs != "" {
			attrs = append(attrs, "errors", errs)
		}

		switch {
		case status >= 500:
			logger.ErrorContext(c.Request.Context(), "request", attrs...)
		case status >= 400:
			logger.WarnContext(c.Request.Context(), "request", attrs...)
		default:
			logger.InfoContext(c.Request.Context(), "request", attrs...)
		}
	}
}
