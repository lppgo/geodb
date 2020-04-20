package config

import (
	"github.com/spf13/viper"
)

func init() {
	Config = viper.New()
	Config.SetDefault("USERDB_PORT", ":8090")
	Config.SetDefault("USERDB_PATH", "/tmp/userdb")
	Config.SetDefault("USERDB_GC_INTERVAL", "5m")
	Config.AutomaticEnv()
}

var Config *viper.Viper
