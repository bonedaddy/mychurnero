package client

import (
	"math/rand"

	"github.com/monero-ecosystem/go-monero-rpc-client/wallet"
)

const (

	// mixing and ring size are hard coded as they are consensus defined
	// eventually these will need to be changed when the network upgrades accordingly

	// Mixing defines the number of outputs to include with ours in the ring
	Mixing = uint64(10)
	// RingSize defines the total number of outputs in the ring (N + 1)
	RingSize = uint64(11)
)

// TransferOpts defines options used to control transfers
// TODO(bonedaddy): support more than out destination address
type TransferOpts struct {
	Priority wallet.Priority
	// maps destination address to amount
	Destinations   map[string]uint64
	AccountIndex   uint64   // defaults to 0
	SubaddrIndices []uint64 // options, default is nil which means all
	WalletName     string
	DoNotRelay     bool
}

// TxConfirmed returns whether or not the given transaction is confirmed
func (c *Client) TxConfirmed(walletName, txHash string) (bool, error) {
	if err := c.OpenWallet(walletName); err != nil {
		return false, err
	}
	resp, err := c.mw.GetTransferByTxID(&wallet.RequestGetTransferByTxID{TxID: txHash})
	if err != nil {
		return false, err
	}
	if resp.Transfer.Confirmations >= resp.Transfer.SuggestedConfirmationsThreshold {
		return true, nil
	}
	return false, nil
}

// TransferSplit allows splitting up a transaction into smaller one, useful
// for situations where Transfer returns an error due to to large of a transaction
func (c *Client) TransferSplit(opts TransferOpts) (*wallet.ResponseTransferSplit, error) {
	if err := c.OpenWallet(opts.WalletName); err != nil {
		return nil, err
	}

	var destinations []*wallet.Destination

	for k, v := range opts.Destinations {
		destinations = append(destinations, &wallet.Destination{
			Address: k,
			Amount:  v,
		})
	}

	return c.mw.TransferSplit(&wallet.RequestTransferSplit{
		Mixin:          10,
		RingSize:       11,
		Priority:       opts.Priority,
		GetTxHex:       true,
		GetxKeys:       true, // TODO: needs to change
		GetTxMetadata:  true,
		DoNotRelay:     opts.DoNotRelay,
		AccountIndex:   opts.AccountIndex,
		SubaddrIndices: opts.SubaddrIndices,
		Destinations:   destinations,
	})
}

// Transfer is used to transfer funds from the given wallet to the destination address
func (c *Client) Transfer(opts TransferOpts) (*wallet.ResponseTransfer, error) {
	if err := c.OpenWallet(opts.WalletName); err != nil {
		return nil, err
	}

	var destinations []*wallet.Destination

	for k, v := range opts.Destinations {
		destinations = append(destinations, &wallet.Destination{
			Address: k,
			Amount:  v,
		})
	}

	return c.mw.Transfer(&wallet.RequestTransfer{
		Mixing:         10, // TODO: needs to change
		RingSize:       11,
		Priority:       opts.Priority,
		GetTxHex:       true,
		GetTxKey:       true,
		GetTxMetadata:  true,
		DoNotRelay:     opts.DoNotRelay,
		AccountIndex:   opts.AccountIndex,
		SubaddrIndices: opts.SubaddrIndices,
		Destinations:   destinations,
	})
}

// Relay is used to relay an unbroadcasted transaction returning the tx hash
func (c *Client) Relay(walletName, txMetadata string) (string, error) {
	if err := c.OpenWallet(walletName); err != nil {
		return "", err
	}

	resp, err := c.mw.RelayTx(&wallet.RequestRelayTx{Hex: txMetadata})
	if err != nil {
		return "", err
	}
	return resp.TxHash, nil
}

// RandomPriority returns a random transaction priority
// note that this can potentially become expensive
func RandomPriority() wallet.Priority {
	return wallet.Priority(rand.Int63n(3))
}
