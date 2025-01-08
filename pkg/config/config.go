package config

import (
	"os"

	"github.com/eastygh/webm-nas/pkg/utils/ratelimit"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server      ServerConfig           `yaml:"server"`
	DB          DBConfig               `yaml:"db"`
	Redis       RedisConfig            `yaml:"redis"`
	OAuthConfig map[string]OAuthConfig `yaml:"oauth"`
	Revers      ReversProxyConfig      `yaml:"revers"`
	Static      StaticContentConfig    `yaml:"static"`
}

type ServerConfig struct {
	ENV                    string                  `yaml:"env"`
	Address                string                  `yaml:"address"`
	Port                   int                     `yaml:"port"`
	GracefulShutdownPeriod int                     `yaml:"gracefulShutdownPeriod"`
	LimitConfigs           []ratelimit.LimitConfig `yaml:"rateLimits"`
	JWTSecret              string                  `yaml:"jwtSecret"`
}

type DBConfig struct {
	Type     string `yaml:"type"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Filename string `yaml:"filename"`
	Migrate  bool   `yaml:"migrate"`
}

type ReversProxyConfig struct {
	Enable    bool              `yaml:"enable"`
	ProxyUrls map[string]string `yaml:"proxyUrls"`
}

type StaticContentConfig struct {
	Enable   bool              `yaml:"enable"`
	Contents map[string]string `yaml:"contents"` // key: path, value: dir
	SpaPath  string            `yaml:"spaPath"`  // which is the base uri using for spa
}

type RedisConfig struct {
	Enable   bool   `yaml:"enable"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Password string `yaml:"password"`
}

type OAuthConfig struct {
	AuthType     string `yaml:"authType"`
	ClientId     string `yaml:"clientId"`
	ClientSecret string `yaml:"clientSecret"`
}

func Parse(appConfig string) (*Config, error) {
	config := &Config{}

	file, err := os.Open(appConfig)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if err := yaml.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}
