package client

import (
	"fmt"
	"log"
	"sync"

	"github.com/monero-ecosystem/go-monero-rpc-client/wallet"
)

type Client struct {
	mw  wallet.Client
	mux sync.RWMutex
}

func NewClient(rpcAddr string) (*Client, error) {
	mclient := wallet.New(wallet.Config{
		Address: rpcAddr,
	})
	return &Client{mw: mclient}, nil
}

func (c *Client) Close() error {
	c.mux.Lock()
	defer c.mux.Unlock()
	if err := c.mw.Store(); err != nil {
		log.Println("failed to save wallet: ", err)
	}
	return c.mw.CloseWallet()
}

func (c *Client) Refresh(account_name string) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	if err := c.OpenWallet(account_name); err != nil {
		return err
	}
	_, err := c.mw.Refresh(&wallet.RequestRefresh{})
	return err
}

// look up balance for the given address (not the wallet)
func (c *Client) AddressBalance(wallet_name string, address string) (uint64, error) {
	c.mux.Lock()
	defer c.mux.Unlock()
	if err := c.OpenWallet(wallet_name); err != nil {
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

func (c *Client) Transfer(name string, dest_address string, amount uint64) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	if err := c.OpenWallet(name); err != nil {
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
			&wallet.Destination{
				Address: dest_address,
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
