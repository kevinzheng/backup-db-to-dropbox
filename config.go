package main

import (
	"fmt"

	"github.com/spf13/viper"
)

var (
	Config struct {
		Backup struct {
			TmpDir string `yaml:"tmpDir"`
		}

		Dropbox struct {
			Token string `yaml:"token"`
			Log   bool   `yaml:"log"`
		}

		Source struct {
			Host     string   `yaml:"host"`
			Port     string   `yaml:"port"`
			Username string   `yaml:"username"`
			Password string   `yaml:"password"`
			Dbs      []string `yaml:"dbs"`
		} `yaml:"source"`
	}
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	err = viper.Unmarshal(&Config)
	if err != nil {
		fmt.Printf("unable to decode into struct, %v\n", err)
	}
}
