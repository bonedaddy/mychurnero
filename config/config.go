package config

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/monero-ecosystem/go-monero-rpc-client/wallet"
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
	LogPath    string
	// specifies the account index to use for receiving churned funds to
	ChurnAccountIndex uint64
	// defines the minimum balance an address must have to be churned from
	MinChurnAmount uint64
	// specifies the minimum delay in minutes to use
	MinDelayMinutes int64
	// specifies the maximum delay in minutes to use for relaying a transaction after it is created
	MaxDelayMinutes int64
	// how often we will check for churnable addresses
	ScanInterval time.Duration
}

// DefaultConfig returns a default configuration suitable for testing
func DefaultConfig() *Config {
	return &Config{
		DBPath:            "mychurnero.db",
		WalletName:        "testnetwallet123",
		RPCAddress:        "http://127.0.0.1:6061/json_rpc",
		LogPath:           "mychurnero.log",
		ChurnAccountIndex: 1,
		MinChurnAmount:    wallet.Float64ToXMR(0.1),
		MinDelayMinutes:   1,
		MaxDelayMinutes:   10,
		ScanInterval:      time.Minute,
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
