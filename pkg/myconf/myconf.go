package myconf

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	APIURL    string
	AppName   string
	ClientURL string

	AWSRegion       string
	AWSUploadBucket string
}

func Load(name string) *Config {
	config := viper.New()
	config.SetConfigName(name)
	config.AddConfigPath(".")
	err := config.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	conf := &Config{
		APIURL:    config.Get("app.api_url").(string),
		AppName:   config.Get("app.name").(string),
		ClientURL: config.Get("app.client_url").(string),

		AWSRegion:       config.Get("aws.region").(string),
		AWSUploadBucket: config.Get("aws.upload_bucket").(string),
	}

	return conf
}
