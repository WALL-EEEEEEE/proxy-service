package config

type Config struct {
	Redis struct {
		Address  string `yaml:"address" envconfig:"REDIS_ADDRESS"`
		Password string `yaml:"password" envconfig:"REDIS_PASSWORD"`
	} `yaml:"redis"`
	Mysql struct {
		Host     string `yaml:"host" envconfig:"MYSQL_HOST"`
		Port     string `yaml:"port" envconfig:"MYSQL_PORT"`
		User     string `yaml:"user" envconfig:"MYSQL_USER"`
		Password string `yaml:"password" envconfig:"MYSQL_PASSWORD"`
		Database string `yaml:"database" envconfig:"MYSQL_DATABASE"`
	} `yaml:"mysql"`
}
