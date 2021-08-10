package http

import (
	"bytes"
	"errors"
	"io"

	"github.com/gin-gonic/gin/render"
	"github.com/teamlint/pkg/log"
)

var (
	ErrHTMLRenderUndefined   = errors.New("html render is undefined")
	ErrHTMLRenderTemplateNil = errors.New("html render's template is nil")
	ErrTemplateUndefined     = errors.New("template is undefined")
)

type HTMLRender struct {
	hr render.HTMLRender
}

func NewHTMLRender() *HTMLRender {
	return &HTMLRender{}
}

func (r *HTMLRender) Configure(s *Server) {
	hr := s.Engine.HTMLRender
	if hr == nil {
		log.Error().Err(ErrHTMLRenderUndefined).Msg("HTMLRender.Configure")
		return
	}
	r.hr = hr
}

func (r *HTMLRender) Output(w io.Writer, name string, data interface{}) error {
	if r.hr == nil {
		log.Error().Err(ErrHTMLRenderUndefined).Msg("HTMLRender.Output")
		return ErrHTMLRenderUndefined
	}
	t := r.hr.Instance("", nil).(render.HTML).Template
	if t == nil {
		log.Error().Err(ErrHTMLRenderTemplateNil).Msg("HTMLRender.Output")
		return ErrHTMLRenderTemplateNil
	}
	tmpl := t.Lookup(name)
	if tmpl == nil {
		return ErrTemplateUndefined
	}
	return t.ExecuteTemplate(w, name, data)
}

func (r *HTMLRender) String(name string, data interface{}) (string, error) {
	buf := bytes.NewBufferString("")
	err := r.Output(buf, name, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (r *HTMLRender) MustString(name string, data interface{}) string {
	str, err := r.String(name, data)
	if err != nil {
		if err == ErrTemplateUndefined {
			log.Warn().Msgf(`HTMLRender: template "%s" is undefined`, name)
			return ""
		}
	}
	return str
}

func (r *HTMLRender) Bytes(name string, data interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := r.Output(buf, name, data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (r *HTMLRender) MustBytes(name string, data interface{}) []byte {
	b, err := r.Bytes(name, data)
	if err != nil {
		if err == ErrTemplateUndefined {
			log.Warn().Msgf(`HTMLRender: template "%s" is undefined`, name)
			return []byte{}
		}
	}
	return b
}
