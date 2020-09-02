package client

import "github.com/monero-ecosystem/go-monero-rpc-client/wallet"

func (c *Client) CreateWallet(account_name string) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.mw.CreateWallet(&wallet.RequestCreateWallet{
		Filename: account_name,
		Language: "English",
	})
}

// look up balance for all address in wallet
func (c *Client) WalletBalance(wallet_name string) (uint64, error) {
	c.mux.Lock()
	defer c.mux.Unlock()
	if err := c.OpenWallet(wallet_name); err != nil {
		return 0, err
	}
	resp, err := c.mw.GetBalance(&wallet.RequestGetBalance{AccountIndex: 0})
	if err != nil {
		return 0, err
	}
	return resp.UnlockedBalance, nil
}

func (c *Client) OpenWallet(name string) error {
	return c.mw.OpenWallet(&wallet.RequestOpenWallet{Filename: name})
}

func (c *Client) SaveWallet() error {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.mw.Store()
}
