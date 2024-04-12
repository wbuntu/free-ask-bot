/*
Copyright © 2022 wbuntu
*/
package cmd

import (
	"fmt"
	"os"

	"gitbub.com/wbuntu/gin-template/internal/pkg/config"
	"gitbub.com/wbuntu/gin-template/internal/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gin-template",
	Short: "A playground for personal project",
	RunE:  run,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Printf("execute cmd error: %s\n", err)
		os.Exit(1)
	}
}

func init() {
	initFlags()
	initCmds()
	cobra.OnInitialize(initConfig)
}

var (
	version string
	cfgFile string
)

func initCmds() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(configCmd)
}

func initFlags() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "/etc/gin-template/config.toml", "toml config file")
	rootCmd.PersistentFlags().Int("log-level", 4, "trace=6, debug=5, info=4, error=2, fatal=1, panic=0")
}

func initConfig() {
	// set default values
	viper.BindPFlag("general.log_level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.SetDefault("general.log_format", "text")
	viper.SetDefault("general.enable_db", true)
	viper.SetDefault("general.enable_kv_db", false)
	viper.SetDefault("general.enable_leader_election", false)

	// api
	viper.SetDefault("api.addr", "0.0.0.0:8080")
	viper.SetDefault("api.tls_addr", "")
	viper.SetDefault("api.tls_crt", "")
	viper.SetDefault("api.tls_key", "")
	// db: default to sqlite for test
	// go-mysql-server -> :memory: -> gin-template:gin-template@tcp(127.0.0.1:6603)/db?charset=utf8mb4&parseTime=True&loc=Local
	// mysql -> gin-template:gin-template@tcp(127.0.0.1:3306)/db?charset=utf8mb4&parseTime=True&loc=Local
	viper.SetDefault("db.type", "sqlite")
	viper.SetDefault("db.dsn", "/var/lib/gin-template/sqlite.db")
	viper.SetDefault("db.min_idle_conns", 100)
	viper.SetDefault("db.max_active_conns", 200)
	viper.SetDefault("db.conn_lifetime", "1h0m0s")
	viper.SetDefault("db.conn_idletime", "30m0s")
	// kv_db: default to redis
	viper.SetDefault("kv_db.type", "redis")
	viper.SetDefault("kv_db.dsn", "redis://127.0.0.1:6379")
	viper.SetDefault("kv_db.min_idle_conns", 100)
	viper.SetDefault("kv_db.max_active_conns", 200)
	viper.SetDefault("kv_db.conn_lifetime", "1h0m0s")
	viper.SetDefault("kv_db.conn_idletime", "30m0s")

	// read in environment variables that match
	viper.AutomaticEnv()

	// read config file if exists
	if utils.FileExists(cfgFile) {
		viper.SetConfigFile(cfgFile)
		if err := viper.ReadInConfig(); err != nil {
			fmt.Printf("read config file error: %s\n", err)
			os.Exit(1)
		}
	}

	// unmarshal config
	if err := viper.Unmarshal(&config.C); err != nil {
		fmt.Printf("unmarshal config file error: %s\n", err)
		os.Exit(1)
	}

	// version
	config.C.Version = version
}
