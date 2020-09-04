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
		schedule     uint
	}

	tests := []struct {
		name         string
		args         args
		wantBalance  uint64
		wantSchedule uint
	}{
		{"1", args{walletName, address, baseAddress, 0, 0, 100, 0}, 100, 0},
		{"2", args{walletName, address, baseAddress, 0, 0, 200, 1}, 200, 0},
		{"3", args{walletName, address, baseAddress, 0, 0, 200, 1}, 200, 1}, // trigger already scheduled case
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
			require.Equal(t, int(addr.Balance), int(tt.wantBalance))
			require.Equal(t, addr.Address, address)
			require.Equal(t, addr.WalletName, walletName)
			require.Equal(t, addr.Scheduled, tt.wantSchedule)

			err = db.SetScheduled(address, tt.args.schedule)
			require.NoError(t, err)

			addr, err = db.GetAddress(tt.args.address)
			require.NoError(t, err)
			require.Equal(t, int(addr.Scheduled), int(tt.args.schedule))

			// now add address to trigger scheduled case handling
			require.NoError(t, db.AddAddress(
				tt.args.wallet,
				tt.args.address,
				tt.args.baseAddress,
				tt.args.accountIndex,
				tt.args.addressIndex,
				tt.args.balance,
			))

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
		name         string
		args         args
		wantErr      bool // not yet used but left for future use
		wantTxCount  int
		wantSendable int
	}{
		{"1", args{"1", "1", time.Now().AddDate(0, 0, -1), 0}, false, 1, 1},
		{"2", args{"2", "2", time.Now().Add(time.Hour), 1}, false, 2, 1},
		{"3", args{"3", "3", time.Now().Add(time.Hour * 10), 1}, false, 3, 1},
		{"4", args{"4", "4", time.Now().AddDate(0, 0, -2), 0}, false, 4, 2},
		{"5", args{"5", "5", time.Now().AddDate(0, 0, -3), 0}, false, 5, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := db.AddTransaction(tt.args.sender, tt.args.metadata, tt.args.sendTime)
			require.NoError(t, err)

			err = db.SetTxSpent(tt.args.sender, tt.args.spent)
			require.NoError(t, err)

			tx, err := db.GetTransaction(tt.args.sender)
			require.NoError(t, err)
			require.Equal(t, tx.TxMetadata, tt.args.metadata)
			require.Equal(t, int(tx.Spent), int(tt.args.spent))
			require.True(t, tx.SendTime.Equal(tt.args.sendTime))

			txs, err := db.GetTransactions()
			require.NoError(t, err)
			require.Len(t, txs, tt.wantTxCount)

			sendable, err := db.GetSendableTransactions()
			if tt.wantSendable > 0 { // otherwise for no found txs this will be an error
				require.NoError(t, err)
			}
			require.Len(t, sendable, tt.wantSendable)
		})
	}
}
