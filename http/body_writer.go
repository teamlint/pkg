package http

import (
	"bytes"

	"github.com/gin-gonic/gin"
	"github.com/teamlint/pkg/config"
	"github.com/teamlint/pkg/log"
)

type BodyWriter struct {
	cfg *config.Config
}

func NewBodyWriter(cfg *config.Config) *BodyWriter {
	return &BodyWriter{cfg: cfg}
}
func (b *BodyWriter) Configure(s *Server) {
	s.Use(b.log())
}
func (b *BodyWriter) log() gin.HandlerFunc {
	if !b.cfg.Server.BodyLog {
		return func(ctx *gin.Context) {
			ctx.Next()
		}
	}
	return BodyLog()
}

// BodyLog HTTP 响应体输出
func BodyLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		bw := &bodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = bw
		c.Next()
		statusCode := c.Writer.Status()
		if statusCode >= 400 {
			//ok this is an request with error, let's make a record for it
			// now print body (or log in your preferred way)
			log.Error().Msgf("Response status code: %s", statusCode)
			c.Abort()
			return
		}
		log.Info().Msgf("Response body: %s", bw.body.String())
	}
}

type bodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
