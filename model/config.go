package model

import (
	"github.com/spf13/viper"
	"log"
	"os"
	"path/filepath"
)

var appConfig AppConfig

type Proxy struct {
	Host    string `json:"host" mapstructure:"host"`
	Backend string `json:"backend" mapstructure:"backend"`
	Cors    bool   `json:"cors" mapstructure:"cors"`
}

type AppConfig struct {
	Debug     bool    `json:"debug" mapstructure:"debug"`
	CertName  string  `json:"cert_name" mapstructure:"cert_name"`
	ProxyList []Proxy `json:"proxy_list" mapstructure:"proxy_list"`
}

func init() {
	viper.SetConfigFile(filepath.Join(appPath(), "config.json"))
	var err = viper.ReadInConfig()
	if err != nil {
		log.Println("[配置文件异常]", err.Error())
		return
	}
	if err = viper.Unmarshal(&appConfig); err != nil {
		log.Println("[配置文件异常]", err.Error())
		return
	}
	return
}

func appPath() string {
	p, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return p
}

type NewAppConfig struct {
}

func (NewAppConfig) AppConfig() AppConfig {
	return appConfig
}
