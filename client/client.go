package client

import (
	"log"

	"github.com/monero-ecosystem/go-monero-rpc-client/wallet"
)

// Client is a wrapper around the monero wallet rpc
type Client struct {
	mw wallet.Client
}

// NewClient returns a new initialized rpc client wrapper
func NewClient(rpcAddr string) (*Client, error) {
	mclient := wallet.New(wallet.Config{
		Address: rpcAddr,
	})
	return &Client{mw: mclient}, nil
}

// Close terminates the RPC client
func (c *Client) Close() error {
	if err := c.mw.Store(); err != nil {
		log.Println("failed to save wallet: ", err)
	}
	return c.mw.CloseWallet()
}

func (c *Client) Rescan(walletName string) error {
	if err := c.OpenWallet(walletName); err != nil {
		return err
	}
	return c.mw.RescanBlockchain()
}

// Refresh triggles a total refresh of a wallet scanning
// all addresses for incoming transactions
func (c *Client) Refresh(walletName string) error {
	if err := c.OpenWallet(walletName); err != nil {
		return err
	}
	_, err := c.mw.Refresh(&wallet.RequestRefresh{})
	return err
}
