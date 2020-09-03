package db

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3" //include SQLite driver
)

type Client struct {
	// conn gorqlite.Connection
	db  *gorm.DB
	mux sync.RWMutex
}

// NewClient returns a new database clients
func NewClient(db_path string) (*Client, error) {
	os.MkdirAll(filepath.Dir(db_path), 0755)
	db, err := gorm.Open("sqlite3", fmt.Sprintf(
		"file:%s?secure_delete=true&cache=shared",
		db_path,
	))
	if err != nil {
		return nil, err
	}
	db.LogMode(true)
	return &Client{db: db}, nil
}

// Close is used to shutdown the database
func (c *Client) Close() error {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.db.Close()
}

// Destroy is used to tear down tbales if they exist
func (c *Client) Destroy() error {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.db.DropTableIfExists(Address{}, Transfer{}).Error
}

// Setup is used to create the existing tables
func (c *Client) Setup() error {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.db.AutoMigrate(Address{}, Transfer{}).Error
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
		AccountIndex: accountIndex,
		AddressIndex: addressIndex,
		BaseAddress:  baseAddress,
		Address:      address,
		Balance:      balance,
	}).Error
}

// GetAddress returns the given address if it exists
func (c *Client) GetAddress(address string) (*Address, error) {
	var addr Address
	return &addr, c.db.Model(&Address{}).First(&addr, "WHERE address = ?", address).Error
}

func (c *Client) GetTransaction(sourceAddress string) (*Transfer, error) {
	var tx Transfer
	return &tx, c.db.Model(&Transfer{}).First(&tx, "WHERE source_address = ?", sourceAddress).Error
}

func (c *Client) GetTransactions() ([]Transfer, error) {
	var txs []Transfer
	return txs, c.db.Model(&Transfer{}).Find(&txs).Error
}
