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
	// the account to use for receiving churned funds
	churnAccountIndex uint64
}

// New returns a new Service starting all needed internal subprocesses
func New(ctx context.Context, churnAccountIndex uint64, dbPath, walletName, rpcAddr string) (*Service, error) {
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
	return &Service{sched, cl, db, ctx, cancel, walletName, churnAccountIndex}, nil
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
	s.createChurnAccount(s.churnAccountIndex)
	go func() {
		// call the ticker functions manually first
		// since if we dont do this this we have to wait
		// full ticker time until we can
		s.handleGetChurnTick()

		getChurnTicker := time.NewTicker(time.Minute * 20)
		defer getChurnTicker.Stop()

		for {
			select {
			case <-getChurnTicker.C:
				s.handleGetChurnTick()
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

// creates the account to churn funds ti if it does not exist
func (s *Service) createChurnAccount(churnAccountIndex uint64) {
	accts, err := s.mc.GetAccounts(s.walletName)
	if err != nil {
		log.Println("failed to get all accounts: ", err)
		return
	}
	var churnAcctExists bool
	for _, subacct := range accts.SubaddressAccounts {
		if subacct.AccountIndex == churnAccountIndex {
			churnAcctExists = true
		}
	}
	if !churnAcctExists {
		resp, err := s.mc.NewAccount(s.walletName, "churn-account")
		if err != nil {
			log.Println("failed to create churn account: ", err)
			return
		}
		if resp.AccountIndex != churnAccountIndex {
			log.Println("new created account does not match desried churn account index")
			return
		}
	}

}

func (s *Service) handleGetChurnTick() {
	addrs, err := s.mc.GetChurnableAddresses(s.walletName, s.churnAccountIndex)
	if err != nil {
		return
	}
	for _, acct := range addrs.Accounts {
		for _, sub := range acct.Subaddresses {
			bal, err := s.MC().AddressBalance(s.walletName, sub.Address)
			if err != nil {
				log.Println("failed to get balance")
				log.Fatal(err)
			}
			if err := s.db.AddAddress(
				s.walletName,
				sub.Address,
				acct.BaseAddress,
				acct.AccountIndex,
				sub.AddressIndex,
				bal); err != nil {
				log.Println("failed to add address")
				log.Fatal(err)
			}
		}
	}
	fmt.Printf("got churnable addresses\n%#v\n\n", addrs)
}
