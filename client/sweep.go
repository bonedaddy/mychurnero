package client

import "github.com/monero-ecosystem/go-monero-rpc-client/wallet"

// WARNING: do not use these functions without care

// SweepDust is used to sweep all unspendable amounts pre-ringCT
func (c *Client) SweepDust(walletName string) (*wallet.ResponseSweepDust, error) {
	if err := c.OpenWallet(walletName); err != nil {
		return nil, err
	}
	return c.mw.SweepDust(&wallet.RequestSweepDust{GetTxHex: true, GetTxKeys: true})
}

// SweepAll is used to sweep all funds from the given account index sending it to the destination address
func (c *Client) SweepAll(opts TransferOpts) (*wallet.ResponseSweepAll, error) {
	if err := c.OpenWallet(opts.WalletName); err != nil {
		return nil, err
	}
	var addr string
	for k := range opts.Destinations {
		addr = k
	}
	return c.mw.SweepAll(&wallet.RequestSweepAll{
		Address:        addr,
		AccountIndex:   opts.AccountIndex,
		SubaddrIndices: opts.SubaddrIndices,
		Priority:       opts.Priority,
		Mixin:          10,
		RingSize:       11,
		GetTxHex:       true,
		GetTxKeys:      true,
		GetTxMetadata:  true,
		DoNotRelay:     opts.DoNotRelay,
	})
}

// SweepSingle is used to spend all of a specified unlocked output to an address
func (c *Client) SweepSingle(opts TransferOpts) (*wallet.ResponseSweepSingle, error) {
	if err := c.OpenWallet(opts.WalletName); err != nil {
		return nil, err
	}
	var addr string
	for k := range opts.Destinations {
		addr = k
	}
	return c.mw.SweepSingle(&wallet.RequestSweepSingle{
		Address:        addr,
		AccountIndex:   opts.AccountIndex,
		SubaddrIndices: opts.SubaddrIndices,
		Priority:       opts.Priority,
		Mixin:          10,
		RingSize:       11,
		GetTxHex:       true,
		GetxKeys:       true,
		GetTxMetadata:  true,
		DoNotRelay:     opts.DoNotRelay,
	})
}
