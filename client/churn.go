package client

// ChurnableSubAdddress defines a given address that we can churn funds from
type ChurnableSubAdddress struct {
	AddressIndex uint64
	Address      string
	Balance      uint64
}

// ChurnableAccount defines a group of sub addresses we can churn funds from
type ChurnableAccount struct {
	AccountIndex uint64
	BaseAddress  string
	Subaddresses []ChurnableSubAdddress
}

// ChurnableAccounts bundles together all accounts we can churn funds from
type ChurnableAccounts struct {
	Accounts []ChurnableAccount
}

// GetChurnableAddresses is used to get addresses that we can churn by sending to ourselves.
// The account index matching churnAccountIndex is skipped, as this is the account for which
// we will use to send churned funds to
func (c *Client) GetChurnableAddresses(walletName string, churnAccountIndex, minBalance uint64) (*ChurnableAccounts, error) {
	if err := c.OpenWallet(walletName); err != nil {
		return nil, err
	}
	accts, err := c.GetAccounts(walletName)
	if err != nil {
		return nil, err
	}
	churns := &ChurnableAccounts{
		Accounts: make([]ChurnableAccount, 0),
	}
	for _, acct := range accts.SubaddressAccounts {
		// skip this account index as its used for receiving churned funds
		if acct.AccountIndex == churnAccountIndex {
			continue
		}
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
			// TODO(bonedaddy): figure out a better way around this
			// primary account seems to be causing troubles using transfer and specific subaddr indices
			// there is likely a better way around
			// skip primary account
			//if addr.Label == "Primary account" {
			//	continue
			//}
			if addr.Used {
				bal, err := c.AddressBalance(walletName, addr.Address, acct.AccountIndex, addr.AddressIndex)
				if err != nil {
					return nil, err
				}
				// skip addresses with no balance or less than minimum balance
				if bal < minBalance || bal <= 0 {
					continue
				}
				// todo: get balance if it is 0 no point in using
				acct.Subaddresses = append(acct.Subaddresses, ChurnableSubAdddress{
					AddressIndex: addr.AddressIndex,
					Address:      addr.Address,
					Balance:      bal,
				})
			}
		}
		churns.Accounts[i] = acct
	}
	return churns, nil
}
