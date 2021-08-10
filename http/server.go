package http

import (
	"errors"
	"html/template"
	stdhttp "net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/teamlint/pkg/config"
)

// Server
type Server struct {
	*gin.Engine
	Addr string
}

// ServerConfigurator 服务器配置接口
type ServerConfigurator interface {
	Configure(*Server)
}

// New 创建HTTP Sever
func New() *Server {
	e := gin.New()
	s := &Server{Engine: e}
	return s
}

// Default 创建默认HTTP Sever
func Default() *Server {
	e := gin.Default()
	s := &Server{Engine: e}
	return s
}

// NewServer 使用配置器初始化服务器
func NewServer(cs []ServerConfigurator) *Server {
	s := Default()
	s.Configure(cs...)
	return s
}

// Run 监听端口并运行服务器
func (s *Server) Run(addr ...string) error {
	if len(addr) == 0 {
		if s.Addr == "" {
			return errors.New("must be set host address")
		}
		return s.Engine.Run(s.Addr)
	}
	return s.Engine.Run(addr...)
}

// Configure 配置
func (s *Server) Configure(cs ...ServerConfigurator) {
	for _, c := range cs {
		c.Configure(s)
	}
}

// StdServer 标准HTTP服务器
func (s *Server) StdServer(conf *config.Server) *stdhttp.Server {
	readTimeout, _ := time.ParseDuration(conf.ReadTimeout)
	writeTimeout, _ := time.ParseDuration(conf.WriteTimeout)
	idleTimeout, _ := time.ParseDuration(conf.IdleTimeout)
	srv := &stdhttp.Server{
		Addr:         conf.HTTPAddr,
		Handler:      s.Engine,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}
	return srv
}

/*********************************************************************************/
type withAddr struct{ addr string }

func (c *withAddr) Configure(s *Server) {
	s.Addr = c.addr
}

// WithAddr 配置主机服务地址
func WithAddr(addr string) ServerConfigurator {
	return &withAddr{addr: addr}
}

/*********************************************************************************/
type withTemplateFuncMap struct {
	funcMap template.FuncMap
}

func (c *withTemplateFuncMap) Configure(s *Server) {
	s.SetFuncMap(c.funcMap)
}

// WithTemplateFuncMap 模板函数配置
func WithTemplateFuncMap(funcMap template.FuncMap) ServerConfigurator {
	return &withTemplateFuncMap{funcMap: funcMap}
}

/*********************************************************************************/
type withRenderTemplates struct{ pattern string }

func (c *withRenderTemplates) Configure(s *Server) {
	s.Engine.LoadHTMLGlob(c.pattern)
}

// WithRenderTemplates 配置要呈现的模板路径
func WithRenderTemplates(pattern string) ServerConfigurator {
	return &withRenderTemplates{pattern: pattern}
}

/*********************************************************************************/
