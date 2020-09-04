package db

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	dbPath      = "somedb.db"
	walletName  = "somewallet"
	address     = "someaddr"
	baseAddress = "somebaseaddress"
)

func TestAddress(t *testing.T) {
	db, err := NewClient(dbPath)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := db.Destroy()
		if err != nil {
			t.Error(err)
		}
		err = db.Close()
		require.NoError(t, err)
		os.RemoveAll(dbPath)
	})

	require.NoError(t, db.Setup())

	type args struct {
		wallet       string
		address      string
		baseAddress  string
		accountIndex uint64
		addressIndex uint64
		balance      uint64
	}

	tests := []struct {
		name        string
		args        args
		wantBalance uint64
	}{
		{"1", args{walletName, address, baseAddress, 0, 0, 100}, 100},
		{"2", args{walletName, address, baseAddress, 0, 0, 200}, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, db.AddAddress(
				tt.args.wallet,
				tt.args.address,
				tt.args.baseAddress,
				tt.args.accountIndex,
				tt.args.addressIndex,
				tt.args.balance,
			))

			addr, err := db.GetAddress(tt.args.address)
			require.NoError(t, err)
			require.Equal(t, int(addr.Balance), 100)
			require.Equal(t, addr.Address, address)
			require.Equal(t, addr.WalletName, walletName)

			time.Sleep(time.Second * 1) // sleep let time pass for future test

			require.NoError(t, db.AddAddress(
				tt.args.wallet,
				tt.args.address,
				tt.args.baseAddress,
				tt.args.accountIndex,
				tt.args.addressIndex,
				tt.args.balance,
			)) // test update capabilities

			addr2, err := db.GetAddress(tt.args.address)
			require.NoError(t, err)
			require.Equal(t, int(addr2.Balance), 200)
			require.Equal(t, addr2.Address, address)
			require.Equal(t, addr2.WalletName, walletName)
			require.True(t, addr.CreatedAt.Equal(addr2.CreatedAt))
			require.True(t, addr2.CreatedAt.After(addr.CreatedAt))
		})
	}

}

func TestTransaction(t *testing.T) {
	db, err := NewClient(dbPath)
	require.NoError(t, err)

	t.Cleanup(func() {
		err := db.Destroy()
		if err != nil {
			t.Error(err)
		}
		err = db.Close()
		require.NoError(t, err)
		os.RemoveAll(dbPath)
	})

	require.NoError(t, db.Setup())

	type args struct {
		sender   string
		metadata string
		sendTime time.Time
		spent    uint
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool // not yet used but left for future use
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
