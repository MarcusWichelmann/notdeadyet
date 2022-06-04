package config

import "github.com/spf13/viper"

type Config struct {
	Listen    string          `mapstructure:"listen"`
	Apps      []AppConfig     `mapstructure:"apps"`
	Receivers ReceiversConfig `mapstructure:"receivers"`
}

type AppConfig struct {
	Name                  string   `mapstructure:"name"`
	Token                 string   `mapstructure:"token"`
	Timeout               string   `mapstructure:"timeout"`
	RepeatInterval        string   `mapstructure:"repeat_interval"`
	NotificationReceivers []string `mapstructure:"notify"`
}

type ReceiversConfig struct {
	PushoverReceivers []PushoverReceiverConfig `mapstructure:"pushover"`
}

type ReceiverConfig struct {
	Name string `mapstructure:"name"`
}

type PushoverReceiverConfig struct {
	ReceiverConfig `mapstructure:",squash"`
	UserKey        string `mapstructure:"user_key"`
	Token          string `mapstructure:"token"`
	Priority       int    `mapstructure:"priority"`
}

func SetDefaults(v *viper.Viper) {
	v.SetDefault("listen", ":80")
}
