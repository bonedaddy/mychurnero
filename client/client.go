package client

import (
	"fmt"
	"log"
	"sync"

	"github.com/monero-ecosystem/go-monero-rpc-client/wallet"
)

// Client is a wrapper around the monero wallet rpc
// it provides synchronous access to the RPC
type Client struct {
	mw  wallet.Client
	mux sync.RWMutex
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
	c.mux.Lock()
	defer c.mux.Unlock()
	if err := c.mw.Store(); err != nil {
		log.Println("failed to save wallet: ", err)
	}
	return c.mw.CloseWallet()
}

// Refresh triggles a total refresh of a wallet scanning
// all addresses for incoming transactions
func (c *Client) Refresh(walletName string) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	if err := c.OpenWallet(walletName); err != nil {
		return err
	}
	_, err := c.mw.Refresh(&wallet.RequestRefresh{})
	return err
}

// AddressBalance returns the unlocked funds for the given address
// TODO(bonedaddy): accept account and subaddress index
// look up balance for the given address (not the wallet)
func (c *Client) AddressBalance(walletName string, address string) (uint64, error) {
	c.mux.Lock()
	defer c.mux.Unlock()
	if err := c.OpenWallet(walletName); err != nil {
		return 0, err
	}
	resp, err := c.mw.GetBalance(&wallet.RequestGetBalance{AccountIndex: 0})
	if err != nil {
		return 0, err
	}
	for _, addr := range resp.PerSubaddress {
		if addr.Address == address {
			return addr.UnlockedBalance, nil
		}
	}
	return 0, nil
}

// Transfer is used to transfer funds from the given wallet to the destination address
// TODO(bonedaddy): filter by source address
func (c *Client) Transfer(walletName string, destAddress string, amount uint64) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	if err := c.OpenWallet(walletName); err != nil {
		return err
	}
	resp, err := c.mw.Transfer(&wallet.RequestTransfer{
		Mixing:       10,
		RingSize:     11,
		Priority:     wallet.PriorityDefault,
		GetTxHex:     true,
		GetTxKey:     true,
		AccountIndex: 0,
		Destinations: []*wallet.Destination{
			{
				Address: destAddress,
				Amount:  amount,
			},
		},
	})
	if err != nil {
		return err
	}
	fmt.Printf("%#v\n", resp)
	return nil
}
