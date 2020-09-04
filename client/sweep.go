package client

import "github.com/monero-ecosystem/go-monero-rpc-client/wallet"

// WARNING: do not use these functions without care

// SweepDust is used to sweep all unspendable amounts pre-ringCT
func (c *Client) SweepDust(walletName string) (*wallet.ResponseSweepDust, error) {
	if err := c.OpenWallet(walletName); err != nil {
		return nil, err
	}
	return c.mw.SweepDust(&wallet.RequestSweepDust{GetTxHey: true, GetTxKeys: true})
}

// SweepAll is used to sweep all funds from the given account index sending it to the destination address
func (c *Client) SweepAll(walletName, destAddress string, accountIndex uint64) (*wallet.ResponseSweepAll, error) {
	if err := c.OpenWallet(walletName); err != nil {
		return nil, err
	}
	return c.mw.SweepAll(&wallet.RequestSweepAll{
		Address:      destAddress,
		AccountIndex: accountIndex,
		Mixin:        10,
		RingSize:     11,
		GetTxHex:     true,
		GetTxKeys:    true,
	})
}
