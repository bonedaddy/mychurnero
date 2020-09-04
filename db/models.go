package db

import (
	"time"

	"github.com/jinzhu/gorm"
)

// Address specifies an address, its balance and the wallet it came from
type Address struct {
	gorm.Model
	WalletName   string
	AccountIndex uint   // indicates the wallet account this is a part of
	AddressIndex uint   // indicates the subaddress index
	BaseAddress  string // indicates the base wallet account address
	Address      string `gorm:"unique"` // this is the wallet account subaddress
	Balance      uint
}

// Transfer is a single transfer to churn an address
type Transfer struct {
	gorm.Model
	TxMetadata string    // the transaction metadata we use to relay
	SendTime   time.Time // the time at which we will relay the transaction
	Spent      uint      // indicates if the tx is spent (ie broadcasted), 0 = false 1 = true
}
