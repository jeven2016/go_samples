package common

type Config struct {
	ApiServerConfig ApiServerConfig `mapstructure:"api-server,omitempty"`
	MongoConfig     MongoConfig     `mapstructure:"mongodb,omitempty"`
}

type ApiServerConfig struct {
	LogLevel      string `mapstructure:"log_level,omitempty"`
	LogPath       string `mapstructure:"log_path,omitempty"`
	OutputConsole bool   `mapstructure:"output_console,omitempty"`
	Dev           bool   `mapstructure:"dev,omitempty"`
	ServiceName   string `mapstructure:"service_name,omitempty"`
	Port          int32  `mapstructure:"port,omitempty"`
	BindAddress   string `mapstructure:"bind_address,omitempty"`
}

type MongoConfig struct {
	Uri      string `bson:"uri"`
	Database string `bson:"database"`
}