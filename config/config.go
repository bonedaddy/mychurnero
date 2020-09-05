package config

import (
	"io/ioutil"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

// Config is used to configure the mychurnero service
type Config struct {
	// specifies the path to store the sqlite3 database
	DBPath string
	// the name of the wallet to open
	WalletName string
	// the address of a monero-wallet-rpc node
	RPCAddress string
	// specifies the account index to use for receiving churned funds to
	ChurnAccountIndex uint64
	// specifies the minimum delay to use for relaying a transaction after it is created
	MinDelay time.Duration
	// specifies the maximum delay to use for relaying a transaction after it is created
	MaxDelay time.Duration
}

// DefaultConfig returns a default configuration suitable for testing
func DefaultConfig() *Config {
	return &Config{
		DBPath:            "mychurnero.db",
		WalletName:        "testnetwallet123",
		RPCAddress:        "http://127.0.0.1:6061/json_rpc",
		ChurnAccountIndex: 1,
		MinDelay:          time.Minute,
		MaxDelay:          time.Minute * 10,
	}
}

// Save stores the given config at path
func Save(config *Config, path string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, os.FileMode(0640))
}

// Load returns a config object reading contents from path
func Load(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
