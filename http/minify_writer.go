package http

import (
	"bufio"
	"io"
	"net"
	http "net/http"
	"regexp"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
	"github.com/tdewolff/minify/v2/js"
	"github.com/tdewolff/minify/v2/json"
	"github.com/tdewolff/minify/v2/svg"
	"github.com/tdewolff/minify/v2/xml"
	"github.com/teamlint/pkg/config"
	"github.com/teamlint/pkg/log"
)

const (
	noWritten     = -1
	defaultStatus = http.StatusOK
)

type MinifyWriter struct {
	m   *minify.M
	cfg *config.Config
}

func NewMinifyWriter(cfg *config.Config) *MinifyWriter {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.Add("text/html", &html.Minifier{KeepDefaultAttrVals: false, KeepDocumentTags: true})
	m.AddFunc("image/svg+xml", svg.Minify)
	m.AddFuncRegexp(regexp.MustCompile("^(application|text)/(x-)?(java|ecma)script$"), js.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), json.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), xml.Minify)
	return &MinifyWriter{m: m, cfg: cfg}
}
func (c *MinifyWriter) Configure(s *Server) {
	s.Use(c.minify("text/html"))
}

func (c *MinifyWriter) minify(mediaType string) gin.HandlerFunc {
	if !c.cfg.Server.HTMLMinify {
		return func(ctx *gin.Context) {
			ctx.Next()
		}
	}
	return Minify(c.m, mediaType)
}

func Minify(m *minify.M, mediaType string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var mw *minifyResponseWriter
		if w, ok := ctx.Writer.(gin.ResponseWriter); ok {
			log.Debug().Msgf("minifyResponseWriter new")
			mw = NewMinifyResponseWriter(m, w, ctx.Request)
			ctx.Writer = mw
			ctx.Next()
		} else {
			ctx.Next()
			return
		}
		log.Debug().Msgf("minifyResponseWriter flush()")
		mw.Flush()
	}
}

/******************************************************************************/
// NewMinifyResponseWriter 创建 NewMinifyResponseWriter
func NewMinifyResponseWriter(m *minify.M, w gin.ResponseWriter, r *http.Request) *minifyResponseWriter {
	return &minifyResponseWriter{
		ResponseWriter: w,
		status:         defaultStatus,
		// body:           new(bytes.Buffer),
		writer:    nil,
		m:         m,
		mediatype: "",
	}
}

// minifyResponseWriter 实现gin.ResponseWriter 接口
type minifyResponseWriter struct {
	gin.ResponseWriter
	status int
	size   int
	// body   *bytes.Buffer // the response content body
	// minify
	writer    *minifyWriter
	m         *minify.M
	mediatype string
}

// Returns the HTTP response status code of the current request.
func (w *minifyResponseWriter) Status() int {
	return w.status
}

func (w *minifyResponseWriter) Size() int {
	return w.size
}

func (w *minifyResponseWriter) Written() bool {
	return w.Size() != noWritten
}

// Hijack implements the http.Hijacker interface.
func (w *minifyResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.ResponseWriter.(http.Hijacker).Hijack()
}

// CloseNotify implements the http.CloseNotify interface.
func (w *minifyResponseWriter) CloseNotify() <-chan bool {
	return w.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (w *minifyResponseWriter) Header() http.Header {
	return w.ResponseWriter.Header() // use the actual response header
}

func (w *minifyResponseWriter) WriteHeader(status int) {
	w.status = status
}

func (w *minifyResponseWriter) WriteHeaderNow() {
	if !w.Written() {
		w.size = 0
		w.ResponseWriter.WriteHeader(w.status)
	}
}
func (w *minifyResponseWriter) Write(b []byte) (n int, err error) {
	if w.writer == nil {
		// first write
		if mediatype := w.Header().Get("Content-Type"); mediatype != "" {
			w.mediatype = mediatype
			log.Debug().Msgf("mediatype=%v", w.mediatype)
		}
		w.writer = PipeWriter(w.m, w.mediatype, w.ResponseWriter)
	}
	// log.Debug().Msgf("minifyResponseWriter Write: %s", string(b))
	n, err = w.writer.Write(b)
	if err != nil {
		log.Error().Err(err).Msg("minifyResponseWriter Write")
		return
	}
	w.size += n
	return
}

func (w *minifyResponseWriter) WriteString(s string) (n int, err error) {
	w.WriteHeaderNow()
	n, err = w.Write([]byte(s))
	return
}

func (w *minifyResponseWriter) Pusher() (pusher http.Pusher) {
	if pusher, ok := w.ResponseWriter.(http.Pusher); ok {
		return pusher
	}
	return nil
}

// Close must be called when writing has finished. It returns the error from the minifier.
func (w *minifyResponseWriter) Close() error {
	if w.writer != nil {
		if err := w.writer.Close(); err != nil {
			log.Warn().Err(err).Msg("minifyResponseWriter Close")
			return err
		}
		return nil
	}
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
	return nil
}

// Flush implements the http.Flush interface.
func (w *minifyResponseWriter) Flush() {
	w.WriteHeaderNow()
	w.Close()
}

func PipeWriter(m *minify.M, mediatype string, w io.Writer) *minifyWriter {
	pr, pw := io.Pipe()
	mw := &minifyWriter{pw, sync.WaitGroup{}, nil}
	mw.wg.Add(1)
	go func() {
		defer mw.wg.Done()

		if err := m.Minify(mediatype, w, pr); err != nil {
			io.Copy(w, pr)
			mw.err = err
		}
		pr.Close()
	}()
	return mw
}

// minifyWriter makes sure that errors from the minifier are passed down through Close (can be blocking).
type minifyWriter struct {
	pw  *io.PipeWriter
	wg  sync.WaitGroup
	err error
}

// Write intercepts any writes to the writer.
func (w *minifyWriter) Write(b []byte) (int, error) {
	return w.pw.Write(b)
}

// Close must be called when writing has finished. It returns the error from the minifier.
func (w *minifyWriter) Close() error {
	w.pw.Close()
	w.wg.Wait()
	return w.err
}
