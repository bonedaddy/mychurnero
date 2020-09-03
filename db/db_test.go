package db

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	dbPath      = "somedb_Path"
	walletName  = "somewallet"
	address     = "someaddr"
	baseAddress = "somebaseaddress"
)

func TestDB(t *testing.T) {
	db, err := NewClient(dbPath)
	require.NoError(t, err)

	t.Cleanup(func() {
		os.RemoveAll(dbPath)
	})

	defer func() {
		err := db.Destroy()
		if err != nil {
			t.Error(err)
		}
		err = db.Close()
		require.NoError(t, err)
	}()
	err = db.Setup()
	require.NoError(t, err)
	err = db.AddAddress(walletName, address, baseAddress, 0, 0, 100)
	require.NoError(t, err)
	addr, err := db.GetAddress(address)
	require.NoError(t, err)
	require.Equal(t, addr.Balance, 100)
	require.Equal(t, addr.Address, addr)
	require.Equal(t, addr.WalletName, walletName)

	createdAt := addr.CreatedAt
	updatedAt := addr.UpdatedAt

	time.Sleep(time.Second * 3)

	err = db.AddAddress(walletName, address, baseAddress, 0, 0, 200)
	require.NoError(t, err)

	addr2, err := db.GetAddress(address)
	require.NoError(t, err)
	require.Equal(t, addr.Balance, 200)
	require.Equal(t, addr.Address, addr)
	require.Equal(t, addr.WalletName, walletName)

	createdAt2 := addr2.CreatedAt
	updatedAt2 := addr2.UpdatedAt

	require.True(t, createdAt.Equal(createdAt2))

	require.True(t, updatedAt2.After(updatedAt))
}
