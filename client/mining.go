package client

import "github.com/monero-ecosystem/go-monero-rpc-client/wallet"

func (c *Client) StopMining(account_name string) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	if err := c.OpenWallet(account_name); err != nil {
		return err
	}
	return c.mw.StopMining()
}

func (c *Client) StartMining(account_name string, threads uint64) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	if err := c.OpenWallet(account_name); err != nil {
		return err
	}
	return c.mw.StartMining(&wallet.RequestStartMining{
		ThreadsCount: threads,
	})
}
