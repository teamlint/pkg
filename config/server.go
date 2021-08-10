package config

// Server 服务器
type Server struct {
	HTTPAddr     string // HTTP 服务地址
	GRPCAddr     string // GRPC 服务地址
	NATSAddr     string // NATS 服务地址
	DebugAddr    string // Debug 服务地址
	ReadTimeout  string // 读超时
	WriteTimeout string // 写超时
	IdleTimeout  string // 空闲超时
	HTMLMinify   bool   // 压缩HTML
	BodyLog      bool   // Response Body 输出
}
