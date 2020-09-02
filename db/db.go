package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "github.com/mattn/go-sqlite3" //include SQLite driver
)

type User struct {
	Username   string
	Walletname string
	UUID       string
	Balance    float64
}

type Client struct {
	// conn gorqlite.Connection
	db  *sql.DB
	mux sync.RWMutex
}

func NewClient(db_path string) (*Client, error) {
	os.MkdirAll(filepath.Dir(db_path), 0755)
	db_string := fmt.Sprintf(
		"file:%s?secure_delete=true&cache=shared",
		db_path,
	)
	db, err := sql.Open("sqlite3", db_string)
	if err != nil {
		return nil, err
	}

	return &Client{db: db}, nil
}

func (c *Client) Close() error {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.db.Close()
}

func (c *Client) Destroy() error {
	c.mux.Lock()
	defer c.mux.Unlock()
	d1 := `DROP TABLE IF EXISTS "users";`
	d2 := `DROP TABLE IF EXISTS "wallets";`
	d3 := `DROP TABLE IF EXISTS "deposits;"`
	for _, statement := range []string{d1, d2, d3} {
		if _, err := c.db.Exec(statement); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) Setup() error {
	c.mux.Lock()
	defer c.mux.Unlock()
	create1 := `CREATE TABLE IF NOT EXISTS "addresses" (
			"id" INTEGER PRIMARY KEY AUTOINCREMENT,
			"wallet_name" TEXT NOT NULL,
			"address" TEXT NOT NULL,
			"balance" REAL NOT NULL	
	);`
	create2 := `CREATE TABLE IF NOT EXISTS "transfers" (
			"id" INTEGER PRIMARY KEY AUTOINCREMENT,
			"source_address" TEXT NOT NULL,
			"destination_address" TEXT NOT NULL,
			"tx_hash" TEXT NOT NULL,	
	);`
	for _, statement := range []string{create1, create2} {
		if _, err := c.db.Exec(statement); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) AddWallet(username, wallet_address string) error {
	usr, err := c.GetUser(username)
	if err != nil {
		return err
	}
	c.mux.Lock()
	defer c.mux.Unlock()
	statement := fmt.Sprintf(
		`INSERT INTO "wallets" ("username", "name", "address") VALUES ('%s', '%s', '%s')`,
		username, usr.Walletname, wallet_address,
	)
	_, err = c.db.Exec(statement)
	return err
}

func (c *Client) AddUser(username, wallet_name, uuid, password string) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	statement := fmt.Sprintf(
		`INSERT INTO "users" ("username", "password", "wallet_name", "uuid", "balance") VALUES ('%s', '%s', '%s', '%s', 0)`,
		username, password, wallet_name, uuid,
	)
	_, err := c.db.Exec(statement)
	return err
}

func (c *Client) GetUser(username string) (*User, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	statement := fmt.Sprintf(
		`select wallet_name, uuid, balance from users where username = '%s'`,
		username,
	)
	rows, err := c.db.Query(statement)
	if err != nil {
		return nil, err
	}
	var (
		wallet_name, uuid string
		balance           float64
	)
	for rows.Next() {
		rows.Scan(&wallet_name, &uuid, &balance)
		fmt.Printf("username %s, wallet %s, uuid %s, balance: %v", username, wallet_name, uuid, balance)
	}
	return &User{Username: username, Walletname: wallet_name, UUID: uuid}, nil
}

func (c *Client) CreateDeposit(username, address string) error {
	statement := fmt.Sprintf(
		`INSERT INTO "deposits" ("username", "address") VALUES ('%s', '%s')`,
		username, address,
	)
	c.mux.Lock()
	defer c.mux.Unlock()
	_, err := c.db.Exec(statement)
	return err
}
