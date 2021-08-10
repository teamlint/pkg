package http

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"github.com/teamlint/pkg/config"
)

// 未完成版本

type CacheWriter struct {
	cfg   *config.Config
	cache *cache.Cache
}

func NewCacheWriter(cfg *config.Config) *CacheWriter {
	cache := cache.New(5*time.Minute, 10*time.Minute)
	return &CacheWriter{cfg: cfg, cache: cache}
}
func (c *CacheWriter) Configure(s *Server) {
	s.Use(CachePage(c.cache))
}

// bodyCacheWriter is used to cache responses in gin.
type bodyCacheWriter struct {
	gin.ResponseWriter
	cache      *cache.Cache
	requestURI string
}

// Write a JSON response to gin and cache the response.
func (w bodyCacheWriter) Write(b []byte) (int, error) {
	// Write the response to the cache only if a success code
	status := w.Status()
	if 200 <= status && status <= 299 {
		w.cache.Set(w.requestURI, b, cache.DefaultExpiration)
	}

	// Then write the response to gin
	return w.ResponseWriter.Write(b)
}

// CachePage sees if there are any cached responses and returns
// the cached response if one is available.
func CachePage(cache *cache.Cache) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the ignoreCache parameter
		ignoreCache := strings.ToLower(c.Query("ignoreCache")) == "true"

		// See if we have a cached response
		response, exists := cache.Get(c.Request.RequestURI)
		if !ignoreCache && exists {
			// If so, use it
			// c.Data(200, "application/json", response.([]byte))
			// c.Abort()
			c.Writer.Write(response.([]byte))
			c.Next()
		} else {
			// If not, pass our cache writer to the next middleware
			bcw := &bodyCacheWriter{cache: cache, requestURI: c.Request.RequestURI, ResponseWriter: c.Writer}
			c.Writer = bcw
			c.Next()
		}
	}
}
