package client

import (
	"testing"
	"time"

	"github.com/monero-ecosystem/go-monero-rpc-client/wallet"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/require"
)

var (
	testNetRPC    = "http://127.0.0.1:6061/json_rpc"
	testNetWallet = "testnetwallet123"
)

func TestClient(t *testing.T) {
	client, err := NewClient(testNetRPC)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		client.StopMining(testNetWallet)
		require.NoError(t, client.Close())
	})

	// ignore since this will likely always error
	client.CreateWallet(testNetWallet)
	// create random wallet to test no error
	kid, err := ksuid.NewRandom()
	require.NoError(t, err)
	require.NoError(t, client.CreateWallet(kid.String()))

	// start mining
	require.NoError(t, client.StartMining(testNetWallet, 2))

	time.Sleep(time.Second * 10)

	bal, err := client.WalletBalance(testNetWallet)
	require.NoError(t, err)
	t.Log("balance: ", bal)

	addr, err := client.NewAddress(testNetWallet)
	require.NoError(t, err)

	err = client.Transfer(testNetWallet, addr, wallet.Float64ToXMR(0.1))
	require.NoError(t, err)
}
