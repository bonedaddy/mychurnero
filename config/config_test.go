package config

import (
	"os"
	"testing"

	"github.com/monero-ecosystem/go-monero-rpc-client/wallet"
	"github.com/stretchr/testify/require"
)

var testPath = "test.yaml"

func TestConfig(t *testing.T) {
	t.Cleanup(func() {
		os.Remove(testPath)
	})
	cfg := DefaultConfig()
	require.Equal(t, cfg.DBPath, "mychurnero.db")
	require.Equal(t, cfg.WalletName, "testnetwallet123")
	require.Equal(t, cfg.RPCAddress, "http://127.0.0.1:6061/json_rpc")
	require.Equal(t, cfg.LogPath, "mychurnero.log")
	require.Equal(t, cfg.MinChurnAmount, wallet.Float64ToXMR(0.1))
	require.Equal(t, int(cfg.ChurnAccountIndex), 1)
	require.Equal(t, int(cfg.MinDelayMinutes), 1)
	require.Equal(t, int(cfg.MaxDelayMinutes), 10)

	require.NoError(t, Save(cfg, testPath))

	cfg2, err := Load(testPath)
	require.NoError(t, err)
	require.Equal(t, cfg2.DBPath, "mychurnero.db")
	require.Equal(t, cfg2.WalletName, "testnetwallet123")
	require.Equal(t, cfg2.RPCAddress, "http://127.0.0.1:6061/json_rpc")
	require.Equal(t, cfg.LogPath, "mychurnero.log")
	require.Equal(t, cfg2.MinChurnAmount, wallet.Float64ToXMR(0.1))
	require.Equal(t, int(cfg2.ChurnAccountIndex), 1)
	require.Equal(t, int(cfg2.MinDelayMinutes), 1)
	require.Equal(t, int(cfg2.MaxDelayMinutes), 10)

}
