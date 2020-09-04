package client

import (
	"fmt"
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
}

// Transfer is used to transfer funds from the given wallet to the destination address
func (c *Client) Transfer(opts TransferOpts) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	if err := c.OpenWallet(opts.WalletName); err != nil {
		return err
	}

	var destinations = make([]*wallet.Destination, len(opts.Destinations))

	for k, v := range opts.Destinations {
		destinations = append(destinations, &wallet.Destination{
			Address: k,
			Amount:  v,
		})
	}
	resp, err := c.mw.Transfer(&wallet.RequestTransfer{
		Mixing:       10,
		RingSize:     11,
		Priority:     opts.Priority,
		GetTxHex:     true,
		GetTxKey:     true,
		AccountIndex: opts.AccountIndex,
		Destinations: destinations,
	})
	if err != nil {
		return err
	}
	fmt.Printf("%#v\n", resp)
	return nil
}

// RandomPriority returns a random transaction priority
// note that this can potentially become expensive
func RandomPriority() wallet.Priority {
	return wallet.Priority(rand.Int63n(3))
}
