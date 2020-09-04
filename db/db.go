package db

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gorm.io/driver/sqlite" //include SQLite driver
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Client struct {
	// conn gorqlite.Connection
	db  *gorm.DB
	mux sync.RWMutex
}

// NewClient returns a new database clients
func NewClient(db_path string) (*Client, error) {
	os.MkdirAll(filepath.Dir(db_path), 0755)
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?secure_delete=true&cache=shared", db_path)), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold: time.Second,  // Slow SQL threshold
				LogLevel:      logger.Error, // Log level
				Colorful:      false,        // Disable color
			},
		),
	})
	if err != nil {
		return nil, err
	}
	return &Client{db: db}, nil
}

// Close is used to shutdown the database
func (c *Client) Close() error {
	c.mux.Lock()
	defer c.mux.Unlock()
	d, err := c.db.DB()
	if err != nil {
		log.Println(err)
	}
	return d.Close()
}

// Destroy is used to tear down tbales if they exist
func (c *Client) Destroy() error {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.db.Migrator().DropTable(Address{}, Transfer{})
}

// Setup is used to create the existing tables
func (c *Client) Setup() error {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.db.Migrator().CreateTable(Address{}, Transfer{})
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

func (c *Client) GetTransaction(sourceAddress string) (*Transfer, error) {
	var tx Transfer
	return &tx, c.db.Model(&Transfer{}).First(&tx, "source_address = ?", sourceAddress).Error
}

func (c *Client) GetTransactions() ([]Transfer, error) {
	var txs []Transfer
	return txs, c.db.Model(&Transfer{}).Find(&txs).Error
}
