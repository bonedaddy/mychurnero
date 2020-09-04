package client

import "github.com/monero-ecosystem/go-monero-rpc-client/wallet"

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

// NewAddress creates a new address under the given account index
func (c *Client) NewAddress(walletName string, accountIndex uint64) (string, error) {
	c.mux.Lock()
	defer c.mux.Unlock()
	if err := c.OpenWallet(walletName); err != nil {
		return "", err
	}
	resp, err := c.mw.CreateAddress(&wallet.RequestCreateAddress{AccountIndex: accountIndex})
	if err != nil {
		return "", err
	}
	if err != nil {
		return "", err
	}
	return resp.Address, nil
}

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
