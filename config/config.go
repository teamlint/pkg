package config

import (
	"fmt"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

const (
	// App
	DefaultTitle      = "teamlint"
	DefaultCopyright  = "teamlint.com"
	DefaultTimeFormat = "2006-01-02 15:04:05"
	DefaultCharset    = "UTF-8"
	DefaultDebug      = false
	DefaultLogLevel   = "error"
	// Server
	DefaultServerDebugAddr    = ":8090"
	DefaultServerHTTPAddr     = ":8080"
	DefaultServerGRPCAddr     = ":5040"
	DefaultServerNATSAddr     = ":4222"
	DefaultServerReadTimeout  = "5s"
	DefaultServerWriteTimeout = "10s"
	DefaultServerIdleTimeout  = "15s"
	DefaultServerHTMLMinify   = false
	// Databases
	DefaultDatabaseName            = "teamlint"
	DefaultDatabaseDriverName      = "pgx"
	DefaultDatabaseConnString      = "postgres://postgres:postgres@localhost/teamlint?sslmode=disable"
	DefaultDatabaseConnMaxLifetime = "3m"
	DefaultDatabaseMaxOpenConns    = 100
	DefaultDatabaseMaxIdleConns    = 10
	DefaultDatabaseLog             = false
	// Caches
)

var conf Config
var v *viper.Viper

func init() {
	v = viper.New()
	v.SetConfigType("yaml")           // REQUIRED if the config file does not have the extension in the name
	v.AddConfigPath(".")              // optionally look for config in the working directory
	v.AddConfigPath("/etc/teamlint/") // path to look for the config file in
	// v.SetEnvPrefix("")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	conf = Config{
		Viper:     v,
		App:       new(App),
		Server:    new(Server),
		Databases: make(Databases),
		Caches:    make(Caches),
	}
	// default options
	defaultOption(&conf)
	v.SetConfigName("local")

	err := v.ReadInConfig() // Find and read the config file
	if err != nil {         // Handle errors reading the config file
		v.SetConfigName("config")
		err1 := v.ReadInConfig() // Find and read the config file
		if err1 != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); ok {
				// Config file not found; ignore error if desired
				// fmt.Println("config file not found")
				if err = v.Unmarshal(&conf); err != nil {
					panic(fmt.Errorf("Fatal error config unmarshal: %s \n", err))
				}
				return
			} else {
				panic(fmt.Errorf("Fatal error config file: %s \n", err))
			}
		}
	}
	// first unmarshal
	if err = v.Unmarshal(&conf); err != nil {
		panic(fmt.Errorf("Fatal error config unmarshal: %s \n", err))
	}
	// Config file was found but another error was produced
	fmt.Println("config file was found")
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		if err = v.Unmarshal(&conf); err != nil {
			panic(fmt.Errorf("[watch]Fatal error config unmarshal: %s \n", err))
		}
	})
}

// Config 配置
type Config struct {
	*viper.Viper
	App       *App      // 应用程序
	Server    *Server   // 服务器配置
	Databases Databases // 数据库引擎列表
	Caches    Caches    // 缓存引擎列表
}

func NewConfig() *Config {
	return GetConfig()
}

func GetConfig() *Config {
	return &conf
}

// defaultOption 默认值
func defaultOption(conf *Config) {
	// App
	conf.SetDefault("App.Title", DefaultTitle)
	conf.SetDefault("App.Copyright", DefaultCopyright)
	conf.SetDefault("App.TimeFormat", DefaultTimeFormat)
	conf.SetDefault("App.Charset", DefaultCharset)
	conf.SetDefault("App.Debug", DefaultDebug)
	conf.SetDefault("App.LogLevel", DefaultLogLevel)
	conf.SetDefault("App.Settings", map[string]interface{}{})
	// Server("
	conf.SetDefault("Server.DebugAddr", DefaultServerDebugAddr)
	conf.SetDefault("Server.HTTPAddr", DefaultServerHTTPAddr)
	conf.SetDefault("Server.GRPCAddr", DefaultServerGRPCAddr)
	conf.SetDefault("Server.NATSAddr", DefaultServerNATSAddr)
	conf.SetDefault("Server.ReadTimeout", DefaultServerReadTimeout)
	conf.SetDefault("Server.WriteTimeout", DefaultServerWriteTimeout)
	conf.SetDefault("Server.IdleTimeout", DefaultServerIdleTimeout)
	conf.SetDefault("Server.HTMLMinify", DefaultServerHTMLMinify)
	// Databases
	if len(conf.Databases) == 0 {
		defaultDatabase := map[string]interface{}{
			"DriverName":      DefaultDatabaseDriverName,
			"ConnString":      DefaultDatabaseConnString,
			"ConnMaxLifetime": DefaultDatabaseConnMaxLifetime,
			"MaxOpenConns":    DefaultDatabaseMaxOpenConns,
			"MaxIdleConns":    DefaultDatabaseMaxIdleConns,
			"Log":             DefaultDatabaseLog,
		}
		conf.SetDefault("Databases", map[string]interface{}{
			DefaultDatabaseName: defaultDatabase,
		})
		// conf.SetDefault(fmt.Sprintf("Databases.%s.DriverName", DefaultDatabaseName), DefaultDatabaseDriverName)
		// conf.SetDefault(fmt.Sprintf("Databases.%s.ConnString", DefaultDatabaseName), DefaultDatabaseConnString)
		// conf.SetDefault(fmt.Sprintf("Databases.%s.ConnMaxLifetime", DefaultDatabaseName), DefaultDatabaseConnMaxLifetime)
		// conf.SetDefault(fmt.Sprintf("Databases.%s.MaxOpenConns", DefaultDatabaseName), DefaultDatabaseMaxOpenConns)
		// conf.SetDefault(fmt.Sprintf("Databases.%s.MaxIdleConns", DefaultDatabaseName), DefaultDatabaseMaxIdleConns)
		// conf.SetDefault(fmt.Sprintf("Databases.%s.Log", DefaultDatabaseName), DefaultDatabaseLog)

		// note
		// 环境变量添加的值,如果配置结构中没有指定成员,只能通过config.Get("path.path")获取,如果使用配置文件会自动加入
	}
}
