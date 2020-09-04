package client

import "github.com/monero-ecosystem/go-monero-rpc-client/wallet"

// StopMining stops active mining processes
func (c *Client) StopMining(walletName string) error {
	if err := c.OpenWallet(walletName); err != nil {
		return err
	}
	return c.mw.StopMining()
}

// StartMining starts actively mining blocks with the given threads
func (c *Client) StartMining(walletName string, threads uint64) error {
	if err := c.OpenWallet(walletName); err != nil {
		return err
	}
	return c.mw.StartMining(&wallet.RequestStartMining{
		ThreadsCount: threads,
	})
}
