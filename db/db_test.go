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
	require.Equal(t, int(addr.Balance), 100)
	require.Equal(t, addr.Address, address)
	require.Equal(t, addr.WalletName, walletName)

	createdAt := addr.CreatedAt
	updatedAt := addr.UpdatedAt

	time.Sleep(time.Second * 3)

	err = db.AddAddress(walletName, address, baseAddress, 0, 0, 200)
	require.NoError(t, err)

	addr2, err := db.GetAddress(address)
	require.NoError(t, err)
	require.Equal(t, int(addr2.Balance), 200)
	require.Equal(t, addr2.Address, address)
	require.Equal(t, addr2.WalletName, walletName)

	createdAt2 := addr2.CreatedAt
	updatedAt2 := addr2.UpdatedAt

	require.True(t, createdAt.Equal(createdAt2))

	require.True(t, updatedAt2.After(updatedAt))
}

func TestTransaction(t *testing.T) {
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

	type args struct {
		sender   string
		metadata string
		sendTime time.Time
		spent    uint
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		wantTxCount int
	}{
		{"1", args{"1", "1", time.Now().AddDate(0, 0, -1), 1}, false, 1},
		{"2", args{"2", "2", time.Now().Add(time.Hour), 0}, false, 2},
		{"3", args{"3", "3", time.Now().Add(time.Hour * 10), 0}, false, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := db.AddTransaction(tt.args.sender, tt.args.metadata, tt.args.sendTime)
			require.NoError(t, err)

			tx, err := db.GetTransaction(tt.args.sender)
			require.NoError(t, err)
			require.Equal(t, tx.TxMetadata, tt.args.metadata)
			require.Equal(t, int(tx.Spent), 0)
			require.True(t, tx.SendTime.Equal(tt.args.sendTime))

			txs, err := db.GetTransactions()
			require.NoError(t, err)
			require.Len(t, txs, tt.wantTxCount)
		})
	}
}
