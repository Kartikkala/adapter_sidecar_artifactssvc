package config

type StorageServerConfig struct {
	AccessKey         string
	Hostname          string
	Port              uint16
	PolicyGenEndpoint string
}

type AppConfig struct {
	Port uint16
}

type Config struct {
	MainServer StorageServerConfig
	App        AppConfig
}

func NewConfig() *Config {
	return &Config{
		MainServer: StorageServerConfig{
			AccessKey:         "myAccessKey",
			Hostname:          "127.0.0.1",
			Port:              uint16(8080),
			PolicyGenEndpoint: "/internal/policy",
		},
		App: AppConfig{
			Port: uint16(9009),
		},
	}
}
