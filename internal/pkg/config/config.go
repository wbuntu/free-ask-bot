package config

import "time"

var C Config

type Config struct {
	Version string
	General struct {
		LogLevel             int    `mapstructure:"log_level"`
		LogFormat            string `mapstructure:"log_format"`
		EnableDB             bool   `mapstructure:"enable_db"`
		EnableKVDB           bool   `mapstructure:"enable_kv_db"`
		EnableLeaderElection bool   `mapstructure:"enable_leader_election"`
	} `mapstructure:"general"`
	API struct {
		Addr    string `mapstructure:"addr"`
		TLSAddr string `mapstructure:"tls_addr"`
		TLSCrt  string `mapstructure:"tls_crt"`
		TLSKey  string `mapstructure:"tls_key"`
	} `mapstructure:"api"`
	DB struct {
		Type           string        `mapstructure:"type"`
		DSN            string        `mapstructure:"dsn"`
		MinIdleConns   int           `mapstructure:"min_idle_conns"`
		MaxActiveConns int           `mapstructure:"max_active_conns"`
		ConnLifetime   time.Duration `mapstructure:"conn_lifetime"`
		ConnIdletime   time.Duration `mapstructure:"conn_idletime"`
	} `mapstructure:"db"`
	KVDB struct {
		Type           string        `mapstructure:"type"`
		DSN            string        `mapstructure:"dsn"`
		MinIdleConns   int           `mapstructure:"min_idle_conns"`
		MaxActiveConns int           `mapstructure:"max_active_conns"`
		ConnLifetime   time.Duration `mapstructure:"conn_lifetime"`
		ConnIdletime   time.Duration `mapstructure:"conn_idletime"`
	} `mapstructure:"kv_db"`
}
