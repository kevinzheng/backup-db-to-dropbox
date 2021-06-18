package cmd

import (
	"fmt"
	"strings"

	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"springup.xyz/backupdbtodropbox/app"
	"springup.xyz/backupdbtodropbox/config"
)

var (
	configFile    string
	keepBackingUp bool

	dropboxConfig dropbox.Config
	folder        string

	rootCmd = &cobra.Command{
		Use:   "backupdbtodropbox",
		Short: "A tool for backing up database to Dropbox",
		Long:  `A tool for backing up database to Dropbox`,
		Run:   cmdRun,
	}
)

func init() {
	rootCmd.Flags().StringVarP(&configFile, "config", "c", "", "Config file")

	rootCmd.Flags().BoolVarP(&keepBackingUp, "keep", "k", false, "Keep backing up with scheduling")
}

func readConfig() {
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
		viper.AddConfigPath("/etc/backupdbtodropbox/")
		viper.AddConfigPath("$HOME/.backupdbtodropbox")
	}
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	err = viper.Unmarshal(&config.Config)
	if err != nil {
		fmt.Printf("unable to decode into struct, %v\n", err)
	}

	var logLevel dropbox.LogLevel
	if config.Config.Dropbox.Log {
		logLevel = dropbox.LogInfo
	} else {
		logLevel = dropbox.LogOff
	}
	dropboxConfig = dropbox.Config{
		Token:    config.Config.Dropbox.Token,
		LogLevel: logLevel,
	}

	// '/' is needed at start
	if !strings.HasPrefix(config.Config.Dropbox.Folder, "/") {
		folder = "/" + config.Config.Dropbox.Folder
	} else {
		folder = config.Config.Dropbox.Folder
	}
}

func cmdRun(cmd *cobra.Command, args []string) {
	readConfig()

	if keepBackingUp {
		app.Schedule(config.Config.Backup.Cron, dropboxConfig, folder)
	} else {
		app.Backup(dropboxConfig, folder)
	}
}

func Execute() error {
	return rootCmd.Execute()
}
