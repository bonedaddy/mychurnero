package client

import "github.com/monero-ecosystem/go-monero-rpc-client/wallet"

func (c *Client) GetAllAccounts(walletName string) (*wallet.ResponseGetAccounts, error) {
	if err := c.OpenWallet(walletName); err != nil {
		return nil, err
	}
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.mw.GetAccounts(&wallet.RequestGetAccounts{})
}

func (c *Client) GetAddress(walletName string, accountIndex uint64, addressIndex ...uint64) (*wallet.ResponseGetAddress, error) {
	if err := c.OpenWallet(walletName); err != nil {
		return nil, err
	}

	return c.mw.GetAddress(&wallet.RequestGetAddress{
		AccountIndex: accountIndex,
		AddressIndex: addressIndex,
	})
}

// NewAccount is used to create a new account with an optional label
func (c *Client) NewAccount(walletName, label string) (*wallet.ResponseCreateAccount, error) {
	if err := c.OpenWallet(walletName); err != nil {
		return nil, err
	}
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.mw.CreateAccount(&wallet.RequestCreateAccount{
		Label: label,
	})
}
