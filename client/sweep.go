package client

import "github.com/monero-ecosystem/go-monero-rpc-client/wallet"

func (c *Client) SweepDust(walletName string) (*wallet.ResponseSweepDust, error) {
	if err := c.OpenWallet(walletName); err != nil {
		return nil, err
	}
	return c.mw.SweepDust(&wallet.RequestSweepDust{})
}
