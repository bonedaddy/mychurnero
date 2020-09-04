package client

import "github.com/monero-ecosystem/go-monero-rpc-client/wallet"

// CreateWallet is used to create a new monero wallet
func (c *Client) CreateWallet(walletName string) error {
	return c.mw.CreateWallet(&wallet.RequestCreateWallet{
		Filename: walletName,
		Language: "English",
	})
}

// WalletBalance returns the entire unlocked balance of all accounts and subaddresses
func (c *Client) WalletBalance(walletName string) (uint64, error) {
	if err := c.OpenWallet(walletName); err != nil {
		return 0, err
	}
	resp, err := c.mw.GetBalance(&wallet.RequestGetBalance{AccountIndex: 0})
	if err != nil {
		return 0, err
	}
	return resp.UnlockedBalance, nil
}

// OpenWallet is used to open the given wallet using it for all subsequent RPC requests
func (c *Client) OpenWallet(walletName string) error {
	return c.mw.OpenWallet(&wallet.RequestOpenWallet{Filename: walletName})
}

// SaveWallet stores the state of the current actively opened wallet
func (c *Client) SaveWallet() error {
	return c.mw.Store()
}
