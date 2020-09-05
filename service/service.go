package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/bonedaddy/mychurnero/client"
	"github.com/bonedaddy/mychurnero/config"
	"github.com/bonedaddy/mychurnero/db"
	"go.uber.org/multierr"
)

// Service provides monero churning service that takes care of automatically scanning the wallet
// determining which addresses need to be churned, and scheduling the sending of those addresses
type Service struct {
	mc         *client.Client
	db         *db.Client
	ctx        context.Context
	cancel     context.CancelFunc
	walletName string
	// the account to use for receiving churned funds
	churnAccountIndex uint64
	min               int64
	max               int64
}

// New returns a new Service starting all needed internal subprocesses
func New(ctx context.Context, cfg *config.Config) (*Service, error) {
	ctx, cancel := context.WithCancel(ctx)
	cl, err := client.NewClient(cfg.RPCAddress)
	if err != nil {
		cancel()
		return nil, err
	}
	// open the wallet
	if err := cl.OpenWallet(cfg.WalletName); err != nil {
		cancel()
		cl.Close()
		return nil, err
	}
	db, err := db.NewClient(cfg.DBPath)
	if err != nil {
		cancel()
		cl.Close()
		return nil, err
	}
	db.Setup()
	return &Service{cl, db, ctx, cancel, cfg.WalletName, cfg.ChurnAccountIndex, cfg.MinDelayMinutes, cfg.MaxDelayMinutes}, nil
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
		log.Println("getting churnable addresses")
		s.handleGetChurnTick()
		log.Println("scheduling transactions")
		s.createTransactions()

		getChurnTicker := time.NewTicker(time.Minute * 20)
		defer getChurnTicker.Stop()

		// TODO(bonedaddy): better time handling
		deleteTxTicker := time.NewTicker(time.Minute * 1)
		defer deleteTxTicker.Stop()

		for {
			select {
			case <-deleteTxTicker.C:
				log.Println("handling tx confirmation checks")
				s.deleteSpentTransfers()
			case <-getChurnTicker.C:
				log.Println("getting churnable addresses")
				s.handleGetChurnTick()
				log.Println("scheduling transactions")
				s.createTransactions()
			case <-s.ctx.Done():
				return
			}
		}

	}()
}

// Close is used to close the churning service
func (s *Service) Close() error {
	var closeErr error
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

// returns an address we can use to send churned funds to
func (s *Service) getChurnToAddress() (string, error) {
	return s.mc.NewAddress(s.walletName, s.churnAccountIndex)
}

func (s *Service) handleGetChurnTick() {
	addrs, err := s.mc.GetChurnableAddresses(s.walletName, s.churnAccountIndex)
	if err != nil {
		return
	}
	for _, acct := range addrs.Accounts {
		for _, sub := range acct.Subaddresses {
			if err := s.db.AddAddress(
				s.walletName,
				sub.Address,
				acct.BaseAddress,
				acct.AccountIndex,
				sub.AddressIndex,
				sub.Balance); err != nil {
				log.Println("failed to add address")
				log.Fatal(err)
			}
		}
	}
	fmt.Printf("got churnable addresses\n%#v\n\n", addrs)
}

func (s *Service) createTransactions() {
	addrs, err := s.db.GetUnscheduledAddresses()
	if err != nil {
		return
	}
	for _, addr := range addrs {
		churnToAddr, err := s.getChurnToAddress()
		if err != nil {
			log.Println("failed to get churn to address: ", err)
			continue
		}
		priority := client.RandomPriority()
		resp, err := s.mc.Transfer(client.TransferOpts{
			Priority:       priority,
			Destinations:   map[string]uint64{churnToAddr: uint64(addr.Balance)},
			AccountIndex:   uint64(addr.AccountIndex),
			SubaddrIndices: []uint64{uint64(addr.AddressIndex)},
			WalletName:     s.walletName,
			DoNotRelay:     true,
		})
		if err != nil {
			log.Println("failed to create transfer: ", err)
			continue
		}
		log.Printf("created unrelayed transaction with metadata hash: %s\n", s.hashMetadata(resp.TxMetadata))
		delay := s.getRandomSendDelay()
		sendTime := time.Now().Add(delay)
		if err := s.db.ScheduleTransaction(addr.Address, resp.TxMetadata, sendTime); err != nil {
			log.Println("failed to schedule transaction: ", err)
		}
		// TODO(bonedaddy): enable better scheduling instead of creating a bunch of goroutiens
		go func(sourceAddr string) {
			now := time.Now()
			diff := sendTime.Sub(now)
			ticker := time.NewTicker(diff)
			<-ticker.C
			ticker.Stop()
			txData, err := s.db.GetTransaction(sourceAddr)
			if err != nil {
				log.Println("failed to get transaction data from db: ", err)
				return
			}
			txHash, err := s.mc.Relay(s.walletName, txData.TxMetadata)
			if err != nil {
				log.Println("failed to relay transaction: ", err)
				return
			}
			log.Println("relayed transaction with hash ", txHash)
			if err := s.db.SetTxHash(sourceAddr, txHash); err != nil {
				log.Println("failed to set tx hash: ", err)
				return
			}
		}(addr.Address)
	}
}

func (s *Service) deleteSpentTransfers() {
	txs, err := s.db.GetRelayedTransactions()
	if err != nil {
		log.Println("failed to get relayed transactions: ", err)
		return
	}
	for _, tx := range txs {
		confirmed, err := s.mc.TxConfirmed(s.walletName, tx.TxHash)
		if err != nil {
			log.Println("failed to get tx confirmed status: ", err)
			continue
		}
		if confirmed {
			if err := s.db.DeleteTransaction(tx.SourceAddress, tx.TxHash); err != nil {
				log.Println("failed to delete transaction from database: ", err)
				continue
			}
			log.Printf("successfully purged tx information\nhash: %s, sender: %s\n", tx.TxHash, tx.SourceAddress)
		}
	}
}

func (s *Service) hashMetadata(txMetadata string) string {
	hashed := sha256.Sum256([]byte(txMetadata))
	return hex.EncodeToString(hashed[:])
}

func (s *Service) getRandomSendDelay() time.Duration {
	random := rand.Int63n(s.max-s.min+1) + s.min
	dur := time.Duration(random) * time.Minute
	log.Printf("using delay of %v minutes\n", dur.Minutes())
	return dur
}
