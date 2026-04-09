package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type cachedResponse struct {
	statusCode int
	body       []byte
	expiryTime time.Time
}

type IdempotencyCache struct {
	mu    sync.Mutex
	cache map[string]cachedResponse
}

type responseRecorder struct {
	gin.ResponseWriter
	body       []byte
	statusCode int
}

func NewIdempotencyCache() *IdempotencyCache {
	return &IdempotencyCache{
		cache: make(map[string]cachedResponse),
	}
}

func (ic *IdempotencyCache) Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		key := ctx.GetHeader("Idempotency-Key")
		if key == "" {
			ctx.Next()
			return
		}

		// lookup cache
		ic.mu.Lock()
		resp, found := ic.cache[key]
		if found {
			// found && not expired
			if time.Now().Before(resp.expiryTime) {
				ic.mu.Unlock()
				ctx.Data(resp.statusCode, "application/json", resp.body)
				ctx.Abort()
				return
			}
			// found but expired
			delete(ic.cache, key)
		}
		ic.mu.Unlock()

		// capture response
		rec := &responseRecorder{ResponseWriter: ctx.Writer}
		ctx.Writer = rec

		ctx.Next()

		ic.mu.Lock()
		ic.cache[key] = cachedResponse{
			statusCode: rec.statusCode,
			body:       rec.body,
			expiryTime: time.Now().Add(1 * time.Minute),
		}
		ic.mu.Unlock()
	}
}

func (r *responseRecorder) Write(data []byte) (int, error) {
	r.body = append(r.body, data...)    // capture response body
	return r.ResponseWriter.Write(data) // write to original ResponseWriter
}

func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}
