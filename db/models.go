package db

import (
	"encoding/hex"
	"encoding/json"
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
	Scheduled    uint // indicates if it is scheduled, 0 = false, 1 = true
	Spent        uint // indicates if it has been spent, 0 = false, 1 = true
}

// Transfer is a single transfer to churn an address
type Transfer struct {
	gorm.Model
	SourceAddress    string    // the sending address
	TxMetadataEncode string    // the transaction metadata we use to relay, this is hex encoded and must be decoded into the TxMetadata struct
	TxHash           string    // the hash of the transaction once relayed
	SendTime         time.Time // the time at which we will relay the transaction
	Spent            uint      // indicates if the tx is spent (ie broadcasted), 0 = false 1 = true
}

// TxMetadata used to hold one or more metadatas we need to relay
type TxMetadata struct {
	Entries []string
}

// GetMetadata is used to return a TxMetadata type from the encoded data
func (t *Transfer) GetMetadata() (*TxMetadata, error) {
	data, err := hex.DecodeString(t.TxMetadataEncode)
	if err != nil {
		return nil, err
	}
	var txm TxMetadata
	return &txm, json.Unmarshal(data, &txm)
}

// Encode marshals an hex encodes the metadata struct
func (txm *TxMetadata) Encode() (string, error) {
	data, err := json.Marshal(txm)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(data), nil
}
