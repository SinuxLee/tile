package config

const serverConfFile = "config/app.json"

// AppConf ...
type AppConf interface {
	// 是否调试模式运行
	IsDebugMode() bool

	// 日志等级
	GetLogLevel() string

	// 区别不同节点的id
	GetNodeID() int

	// http本地监听端口
	GetHTTPPort() int

	// consul地址
	GetConsulAddr() string
}

// appConfig 服务配置
type appConfig struct {
	DebugMode  bool   `json:"debugMode"`
	LogLevel   string `json:"logLevel"`
	HTTPPort   int    `json:"httpPort"`
	NodeID     int    `json:"nodeId"`
	ConsulAddr string `json:"consulAddr"`
}

// IsDebugMode ...
func (s *appConfig) IsDebugMode() bool {
	return s.DebugMode
}

// GetLogLevel ...
func (s *appConfig) GetLogLevel() string {
	return s.LogLevel
}

// GetNodeID ...
func (s *appConfig) GetNodeID() int {
	return s.NodeID
}

// GetHTTPPort ...
func (s *appConfig) GetHTTPPort() int {
	return s.HTTPPort
}

func (s *appConfig) GetConsulAddr() string {
	return s.ConsulAddr
}

// 加载服务相关配置
func loadServerConf(filePath string, c *config) bool {
	return loadConfFromFile(filePath, &c.appConfig)
}
