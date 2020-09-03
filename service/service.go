package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/bonedaddy/mychurnero/client"
	"github.com/bonedaddy/mychurnero/db"
	"github.com/bonedaddy/mychurnero/txscheduler"
	"go.uber.org/multierr"
)

// Service provides monero churning service that takes care of automatically scanning the wallet
// determining which addresses need to be churned, and scheduling the sending of those addresses
type Service struct {
	s          *txscheduler.TxScheduler
	mc         *client.Client
	db         *db.Client
	ctx        context.Context
	cancel     context.CancelFunc
	walletName string
}

// New returns a new Service starting all needed internal subprocesses
func New(ctx context.Context, dbPath, walletName, rpcAddr string) (*Service, error) {
	ctx, cancel := context.WithCancel(ctx)
	cl, err := client.NewClient(rpcAddr)
	if err != nil {
		cancel()
		return nil, err
	}
	// open the wallet
	if err := cl.OpenWallet(walletName); err != nil {
		cancel()
		cl.Close()
		return nil, err
	}
	db, err := db.NewClient(dbPath)
	if err != nil {
		cancel()
		cl.Close()
		return nil, err
	}
	db.Setup()
	sched := txscheduler.New(ctx)
	sched.Start()
	return &Service{sched, cl, db, ctx, cancel, walletName}, nil
}

// MC returns the underlying monero-wallet-rpc client
func (s *Service) MC() *client.Client {
	return s.mc
}

// DB returns the underlying database client
func (s *Service) DB() *db.Client {
	return s.db
}

// Context returns the underlying context
func (s *Service) Context() context.Context {
	return s.ctx
}

// Start is used to start the churning service
func (s *Service) Start() {
	go func() {
		// first time wiat one minute
		ticker := time.NewTicker(time.Minute * 1)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				addrs, err := s.mc.GetChurnableAddresses(s.walletName)
				if err != nil {
					return
				}
				for _, acct := range addrs.Accounts {
					for _, sub := range acct.Subaddresses {
						addr := sub.Address
						addrIndex := sub.AddressIndex
						address := &db.Address{
							WalletName:   s.walletName,
							AccountIndex: acct.AccountIndex,
							AddressIndex: addrIndex,
							BaseAddress:  acct.BaseAddress,
							Address:      addr,
						}
						// todo: get amount
						fmt.Printf("got churnable address\n%#v\n\n", address)
						if err := s.db.AddAddress(s.walletName, addr, acct.BaseAddress, acct.AccountIndex, addrIndex, 0); err != nil {
							log.Println("failed to add address")
							log.Fatal(err)
						}
					}
				}
				ticker.Stop()
				// now create new ticker with 20 min wait
				ticker = time.NewTicker(time.Minute * 20)
			case <-s.ctx.Done():
				return
			}
		}
	}()
}

// Close is used to close the churning service
func (s *Service) Close() error {
	var closeErr error
	s.s.Stop()
	s.cancel()
	if err := s.mc.Close(); err != nil {
		closeErr = err
	}
	if err := s.db.Close(); err != nil {
		closeErr = multierr.Combine(closeErr, err)
	}
	return closeErr
}
