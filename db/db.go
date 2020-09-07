package db

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/sqlite" //include SQLite driver
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Client provides a wrapper around the gorm client connecting to a sqlite3 database
type Client struct {
	db *gorm.DB
	l  *zap.Logger
}

// NewClient returns a new database clients
func NewClient(l *zap.Logger, dbPath string) (*Client, error) {
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf(
		"file:%s?secure_delete=true&cache=shared",
		dbPath,
	)), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
			logger.Config{
				SlowThreshold: time.Second,   // Slow SQL threshold
				LogLevel:      logger.Silent, // Log level
				Colorful:      false,         // Disable color
			},
		),
	})
	if err != nil {
		return nil, err
	}
	return &Client{l: l.Named("database"), db: db}, nil
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

	// if this address already exists, update with latest balance as long as it is not scheduled
	if addr, err := c.GetAddress(address); err == nil {
		// if address has scheduled transaction skip it
		if addr.Scheduled == 1 {
			c.l.Warn("address already has scheduled transaction, try again later", zap.String("address", address))
			return nil
		}
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

// SetScheduled marks an address as having a scheduled transaction
func (c *Client) SetScheduled(address string, scheduled uint) error {
	addr, err := c.GetAddress(address)
	if err != nil {
		return err
	}
	return c.db.Model(addr).Update("scheduled", scheduled).Error
}

// GetUnscheduledAddresses returns all unscheduled addresses
func (c *Client) GetUnscheduledAddresses() ([]Address, error) {
	var addrs []Address
	return addrs, c.db.Model(&Address{}).Where("scheduled = 0").Find(&addrs).Error
}

// GetAddress returns the given address if it exists
func (c *Client) GetAddress(address string) (*Address, error) {
	var addr Address
	return &addr, c.db.First(&addr, "address = ?", address).Error
}

// GetAddresses returns all known addresses
func (c *Client) GetAddresses() ([]Address, error) {
	var addrs []Address
	return addrs, c.db.Model(&Address{}).Find(&addrs).Error
}

// ScheduleTransaction is used to persist transaction metadata information to disk, marking the
// associated address as being scheduled. This means anytime during startup, we can reschedule transactions
// in case the program exists with pending transactions
func (c *Client) ScheduleTransaction(sourceAddress, txMetadata, metadataHash string, sendTime time.Time) error {
	return c.db.Transaction(func(db *gorm.DB) error {
		var addr Address

		// make sure address exists
		if err := db.Model(&Address{}).Where("address = ?", sourceAddress).First(&addr).Error; err != nil {
			return err
		}

		if err := db.Model(addr).Update("scheduled", 1).Error; err != nil {
			return err
		}

		return db.Create(&Transfer{
			SourceAddress:  sourceAddress,
			TxMetadata:     txMetadata,
			TxMetadataHash: metadataHash,
			SendTime:       sendTime,
			Spent:          0,
		}).Error
	})
}

// DeleteTransaction is used to remove transaction data from our database
// we do this once the transaction has been confirmed and to purge evidence of the churn
func (c *Client) DeleteTransaction(sourceAddress, txHash, metaDataHash string) error {
	tx, err := c.GetTransaction(sourceAddress, metaDataHash)
	if err != nil {
		return err
	}

	if tx.TxHash != txHash {
		return errors.New("invalid transaction found")
	}

	if err := c.db.Delete(tx).Error; err != nil {
		return err
	}

	addr, err := c.GetAddress(sourceAddress)
	if err != nil {
		return err
	}

	return c.db.Delete(addr).Error
}

// AddTransaction is used to store a transaction that we need to relay
func (c *Client) AddTransaction(sourceAddress, txMetadata, metaDataHash string, sendTime time.Time) error {
	return c.db.Create(&Transfer{
		SourceAddress:  sourceAddress,
		TxMetadata:     txMetadata,
		TxMetadataHash: metaDataHash,
		SendTime:       sendTime,
		Spent:          0,
	}).Error
}

// GetRelayedTransactions returns all currently relayed transactions
func (c *Client) GetRelayedTransactions() ([]Transfer, error) {
	var txs []Transfer
	return txs, c.db.Model(&Transfer{}).Where(`tx_hash NOT NULL AND tx_hash != ""`).Find(&txs).Error
}

// SetTxHash sets the transaction hash for the corresponding churn
func (c *Client) SetTxHash(sourceAddress, metaDataHash, txHash string) error {
	tx, err := c.GetTransaction(sourceAddress, metaDataHash)
	if err != nil {
		return err
	}
	return c.db.Model(tx).Update("tx_hash", txHash).Error
}

// SetTxSpent sets the spent field on a transfer entry
func (c *Client) SetTxSpent(sourceAddress, metaDataHash string, spent uint) error {
	tx, err := c.GetTransaction(sourceAddress, metaDataHash)
	if err != nil {
		return err
	}
	return c.db.Model(tx).Update("spent", spent).Error
}

// GetTransaction returns the first matching transaction
func (c *Client) GetTransaction(sourceAddress, metaDataHash string) (*Transfer, error) {
	var tx Transfer
	return &tx, c.db.Model(&Transfer{}).First(&tx, "source_address = ? AND tx_metadata_hash = ?", sourceAddress, metaDataHash).Error
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
