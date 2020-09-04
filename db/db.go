package db

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/sqlite" //include SQLite driver
	"gorm.io/gorm"
)

type Client struct {
	db *gorm.DB
}

// NewClient returns a new database clients
func NewClient(db_path string) (*Client, error) {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf(
		"file:%s?secure_delete=true&cache=shared",
		db_path,
	)), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &Client{db: db}, nil
}

// Close is used to shutdown the database
func (c *Client) Close() error {
	d, err := c.db.DB()
	if err != nil {
		log.Println(err)
	}
	return d.Close()
}

// Destroy is used to tear down tbales if they exist
func (c *Client) Destroy() error {
	return c.db.Migrator().DropTable(Address{}, Transfer{})
}

// Setup is used to create the existing tables
func (c *Client) Setup() error {
	return c.db.Migrator().CreateTable(&Address{}, &Transfer{})
}

// AddAddress is used to store an address into the database, if a previous record with
// this address exists it will be overwritten
func (c *Client) AddAddress(walletName, address, baseAddress string, accountIndex, addressIndex, balance uint64) error {
	// if this address already exists, update with latest balance
	if addr, err := c.GetAddress(address); err == nil {
		return c.db.Model(addr).Update("balance", balance).Error
	}
	return c.db.Create(&Address{
		WalletName:   walletName,
		AccountIndex: uint(accountIndex),
		AddressIndex: uint(addressIndex),
		BaseAddress:  baseAddress,
		Address:      address,
		Balance:      uint(balance),
	}).Error
}

// GetAddress returns the given address if it exists
func (c *Client) GetAddress(address string) (*Address, error) {
	var addr Address
	return &addr, c.db.First(&addr, "address = ?", address).Error
}

// GetAddresses returns all known addresses
func (c *Client) GetAddresses() ([]Address, error) {
	var addrs []Address
	return addrs, c.db.Model(&Transfer{}).Find(&addrs).Error
}

// AddTransaction is used to store a transaction that we need to relay
func (c *Client) AddTransaction(sourceAddress, txMetadata string, sendTime time.Time) error {
	return c.db.Create(&Transfer{
		SourceAddress: sourceAddress,
		TxMetadata:    txMetadata,
		SendTime:      sendTime,
		Spent:         0,
	}).Error
}

// SetTxSpent sets the spent field on a transfer entry
func (c *Client) SetTxSpent(sourceAddress string, spent uint) error {
	tx, err := c.GetTransaction(sourceAddress)
	if err != nil {
		return err
	}
	return c.db.Model(tx).Update("spent", spent).Error
}

// GetTransaction returns the first matching transaction
func (c *Client) GetTransaction(sourceAddress string) (*Transfer, error) {
	var tx Transfer
	return &tx, c.db.Model(&Transfer{}).First(&tx, "source_address = ?", sourceAddress).Error
}

// GetTransactions returns all known transactions
func (c *Client) GetTransactions() ([]Transfer, error) {
	var txs []Transfer
	return txs, c.db.Model(&Transfer{}).Find(&txs).Error
}

// GetSendableTransactions returns all transactions we can relay
func (c *Client) GetSendableTransactions() ([]Transfer, error) {
	var txs []Transfer
	return txs, c.db.Model(&Transfer{}).Where("send_time < ? AND spent = 0", time.Now()).Find(&txs).Error
}
