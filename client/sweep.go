package client

import "github.com/monero-ecosystem/go-monero-rpc-client/wallet"

func (c *Client) SweepDust(walletName string) (*wallet.ResponseSweepDust, error) {
	if err := c.OpenWallet(walletName); err != nil {
		return nil, err
	}
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.mw.SweepDust(&wallet.RequestSweepDust{GetTxHey: true, GetTxKeys: true})
}

func (c *Client) SweepAll(walletName, destAddress string, accountIndex uint64) (*wallet.ResponseSweepAll, error) {
	if err := c.OpenWallet(walletName); err != nil {
		return nil, err
	}
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.mw.SweepAll(&wallet.RequestSweepAll{
		Address:      destAddress,
		AccountIndex: accountIndex,
		Mixin:        10,
		RingSize:     11,
		GetTxHex:     true,
		GetTxKeys:    true,
	})
}
