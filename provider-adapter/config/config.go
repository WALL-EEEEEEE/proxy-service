package config

type Config struct {
	BrightData struct {
		Token   string `yaml:"token" envconfig:"BRIGHT_DATA_TOKEN"`
		Gateway struct {
			Host string `yaml:"host" envconfig:"BRIGHT_DATA_GATEWAY_HOST"`
			Port int64  `yaml:"port" envconfig:"BRIGHT_DATA_GATEWAY_PORT"`
		} `yaml:"gateway" `
	} `yaml:"bright_data"`
}
