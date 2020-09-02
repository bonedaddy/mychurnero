package client

type ChurnableSubAdddress struct {
	AddressIndex uint64
	Address      string
}
type ChurnableAccount struct {
	AccountIndex uint64
	BaseAddress  string
	Subaddresses []ChurnableSubAdddress
}

type ChurnableAccounts struct {
	Accounts []ChurnableAccount
}

func (c *Client) GetChurnableAddresses(walletName string) (*ChurnableAccounts, error) {
	if err := c.OpenWallet(walletName); err != nil {
		return nil, err
	}
	accts, err := c.GetAllAccounts(walletName)
	if err != nil {
		return nil, err
	}
	churns := &ChurnableAccounts{
		Accounts: make([]ChurnableAccount, 0),
	}
	for _, acct := range accts.SubaddressAccounts {
		churns.Accounts = append(churns.Accounts, ChurnableAccount{
			AccountIndex: acct.AccountIndex,
			BaseAddress:  acct.BaseAddress,
		})
	}
	for i, acct := range churns.Accounts {
		acct.Subaddresses = make([]ChurnableSubAdddress, 0)
		addrs, err := c.GetAddress(walletName, acct.AccountIndex)
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs.Addresses {
			if addr.Used {
				// todo: get balance if it is 0 no point in using
				acct.Subaddresses = append(acct.Subaddresses, ChurnableSubAdddress{
					AddressIndex: addr.AddressIndex,
					Address:      addr.Address,
				})
			}
		}
		churns.Accounts[i] = acct
	}
	return churns, nil
}
