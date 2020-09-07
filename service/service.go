package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"strings"
	"time"

	"github.com/bonedaddy/mychurnero/client"
	"github.com/bonedaddy/mychurnero/config"
	"github.com/bonedaddy/mychurnero/db"
	"go.bobheadxi.dev/zapx/zapx"
	"go.uber.org/multierr"
	"go.uber.org/zap"
)

// Service provides monero churning service that takes care of automatically scanning the wallet
// determining which addresses need to be churned, and scheduling the sending of those addresses
type Service struct {
	mc     *client.Client
	db     *db.Client
	ctx    context.Context
	cancel context.CancelFunc
	cfg    *config.Config
	l      *zap.Logger
}

// New returns a new Service starting all needed internal subprocesses
func New(ctx context.Context, cfg *config.Config) (*Service, error) {
	// seed random number generation
	rand.Seed(time.Now().UnixNano())

	l, err := zapx.New(cfg.LogPath, true)
	if err != nil {
		return nil, err
	}

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

	db, err := db.NewClient(l, cfg.DBPath)
	if err != nil {
		cancel()
		cl.Close()
		return nil, err
	}
	db.Setup()

	return &Service{cl, db, ctx, cancel, cfg, l.Named("service")}, nil
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

	s.createChurnAccount(s.cfg.ChurnAccountIndex)
	s.l.Info("mychurnero started")
	s.rescheduleTransactions()
	go func() {
		// call the ticker functions manually first
		// since if we dont do this this we have to wait
		// full ticker time until we can
		s.l.Info("getting churnable addresses")
		s.handleGetChurnTick()
		s.l.Info("scheduling transactions")
		s.createTransactions()

		getChurnTicker := time.NewTicker(s.cfg.ScanInterval)
		defer getChurnTicker.Stop()

		// TODO(bonedaddy): better time handling
		deleteTxTicker := time.NewTicker(time.Minute * 1)
		defer deleteTxTicker.Stop()

		for {
			select {
			case <-deleteTxTicker.C:
				s.l.Info("handling tx confirmation checks")
				s.deleteSpentTransfers()

			case <-getChurnTicker.C:
				s.l.Info("getting churnable addresses")
				s.handleGetChurnTick()
				s.l.Info("scheduling transactions")
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
	accts, err := s.mc.GetAccounts(s.cfg.WalletName)
	if err != nil {
		s.l.Error("failed to get all accounts", zap.Error(err))
		return
	}

	var churnAcctExists bool
	for _, subacct := range accts.SubaddressAccounts {
		if subacct.AccountIndex == churnAccountIndex {
			churnAcctExists = true
		}
	}

	if !churnAcctExists {
		resp, err := s.mc.NewAccount(s.cfg.WalletName, "churn-account")
		if err != nil {
			s.l.Error("failed to create churn account", zap.Error(err))
			return
		}
		if resp.AccountIndex != churnAccountIndex {
			s.l.Warn("new created account does not match desried churn account index")
			return
		}
	}
}

// returns an address we can use to send churned funds to
func (s *Service) getChurnToAddress() (string, error) {
	return s.mc.NewAddress(s.cfg.WalletName, s.cfg.ChurnAccountIndex)
}

func (s *Service) handleGetChurnTick() {
	addrs, err := s.mc.GetChurnableAddresses(s.cfg.WalletName, s.cfg.ChurnAccountIndex, s.cfg.MinChurnAmount)
	if err != nil {
		return
	}

	var toChurn int

	for _, acct := range addrs.Accounts {

		for _, sub := range acct.Subaddresses {

			if err := s.db.AddAddress(
				s.cfg.WalletName,
				sub.Address,
				acct.BaseAddress,
				acct.AccountIndex,
				sub.AddressIndex,
				sub.Balance,
			); err != nil {
				s.l.Error("failed to add address to database", zap.String("address", sub.Address), zap.Error(err))
			} else {
				toChurn++
			}

		}

	}

	if toChurn > 0 {
		s.l.Info("churnable addresses found", zap.Int("count", toChurn))
	}
}

func (s *Service) createTransactions() {
	addrs, err := s.db.GetUnscheduledAddresses()
	if err != nil {
		return
	}

	for _, addr := range addrs {

		churnToAddr, err := s.getChurnToAddress()
		if err != nil {
			s.l.Error("failed to get churn to address", zap.Error(err))
			continue
		}

		sendAmt := s.getRandomBalance(uint64(addr.Balance))
		var metaDataHashes []string
		resp, err := s.mc.Transfer(client.TransferOpts{
			Priority:       client.RandomPriority(),
			Destinations:   map[string]uint64{churnToAddr: sendAmt},
			AccountIndex:   uint64(addr.AccountIndex),
			SubaddrIndices: []uint64{uint64(addr.AddressIndex)},
			WalletName:     s.cfg.WalletName,
			DoNotRelay:     true,
		})
		if err != nil && strings.Contains(err.Error(), "try /transfer_split") {
			resp, err := s.mc.TransferSplit(client.TransferOpts{
				Priority:       client.RandomPriority(),
				Destinations:   map[string]uint64{churnToAddr: sendAmt},
				AccountIndex:   uint64(addr.AccountIndex),
				SubaddrIndices: []uint64{uint64(addr.AddressIndex)},
				WalletName:     s.cfg.WalletName,
				DoNotRelay:     true,
			})
			if err != nil {
				s.l.Error("failed to create split transfer", zap.Error(err))
				continue
			}
			for _, meta := range resp.TxMetadataList {
				metaDataHashes = append(metaDataHashes, meta)
			}
		} else if err != nil {
			s.l.Error("failed to create transfer", zap.String("address", addr.Address), zap.Error(err))
			continue
		} else {
			metaDataHashes = append(metaDataHashes, resp.TxMetadata)
		}

		for _, meta := range metaDataHashes {

			txMetaHash := s.hashMetadata(meta)
			delay := s.getRandomSendDelay()
			sendTime := time.Now().Add(delay)
			s.l.Info("unrelayed transaction created", zap.String("metadata.sha256", txMetaHash), zap.Float64("delay.minutes", delay.Minutes()))
			if err := s.db.ScheduleTransaction(addr.Address, meta, txMetaHash, sendTime); err != nil {
				s.l.Error("failed to schedule transaction", zap.Error(err), zap.String("metadata.sha256", txMetaHash))
				continue
			}
			// TODO(bonedaddy): enable better scheduling instead of creating a bunch of goroutiens
			go func(sourceAddr string) {
				now := time.Now()
				diff := sendTime.Sub(now)
				ticker := time.NewTicker(diff)
				<-ticker.C
				ticker.Stop()
				txData, err := s.db.GetTransaction(sourceAddr, txMetaHash)
				if err != nil {
					s.l.Error("failed to get transaction from database", zap.Error(err), zap.String("metadata.sha256", txMetaHash))
					return
				}
				txHash, err := s.mc.Relay(s.cfg.WalletName, txData.TxMetadata)
				if err != nil {
					s.l.Error("failed to relay transaction", zap.Error(err), zap.String("metadata.sha256", txMetaHash))
					return
				}
				if err := s.db.SetTxHash(sourceAddr, txMetaHash, txHash); err != nil {
					s.l.Error("Failed to set tx hash in database", zap.Error(err))
					return
				}
				s.l.Info("successfully relayed transaction", zap.String("metadata.sha256", txMetaHash), zap.String("tx.hash", txHash))
			}(addr.Address)
		}

	}
}

func (s *Service) deleteSpentTransfers() {
	txs, err := s.db.GetRelayedTransactions()
	if err != nil {
		s.l.Error("failed to get relayed transactions from database", zap.Error(err))
		return
	}

	for _, tx := range txs {
		confirmed, err := s.mc.TxConfirmed(s.cfg.WalletName, tx.TxHash)
		if err != nil {
			s.l.Error("failed to get tx confirmation status", zap.Error(err), zap.String("tx.hash", tx.TxHash))
			continue
		}

		if confirmed {
			if err := s.db.DeleteTransaction(tx.SourceAddress, tx.TxHash, tx.TxMetadataHash); err != nil {
				s.l.Error("failed to delete transaction from database", zap.Error(err), zap.String("tx.hash", tx.TxHash))
				continue
			}
			s.l.Error("transaction purged from database", zap.String("tx.hash", tx.TxHash))
		}

	}
}

// rescheduleTransactions should only be used at startup to see
// if any existing unrelayed transactions are stored in the database, and if so
// send it
func (s *Service) rescheduleTransactions() error {
	txs, err := s.db.GetUnrelayedTransactions()
	if err != nil {
		return err
	}
	for _, tx := range txs {
		tx := tx
		go func() {
			now := time.Now()
			if now.After(tx.SendTime) {
				goto SEND
			}
			time.Sleep(tx.SendTime.Sub(now))
			goto SEND
		SEND:
			txHash, err := s.mc.Relay(s.cfg.WalletName, tx.TxMetadata)
			if err != nil {
				s.l.Error("failed to relay transaction", zap.Error(err), zap.String("metadata.sh256", tx.TxMetadataHash))
				return
			}
			if err := s.db.SetTxHash(tx.SourceAddress, tx.TxMetadataHash, txHash); err != nil {
				s.l.Error("failed to set tx hash in database", zap.Error(err))
				return
			}
		}()
	}
	return nil
}

func (s *Service) hashMetadata(txMetadata string) string {
	hashed := sha256.Sum256([]byte(txMetadata))
	return hex.EncodeToString(hashed[:])
}

func (s *Service) getRandomSendDelay() time.Duration {
	random := rand.Int63n(s.cfg.MaxDelayMinutes-s.cfg.MinDelayMinutes+1) + s.cfg.MinDelayMinutes
	dur := time.Duration(random) * time.Minute
	return dur
}

// returns random balance to send
func (s *Service) getRandomBalance(currentBalance uint64) uint64 {
	if currentBalance < s.cfg.MinChurnAmount {
		return 0
	} else if currentBalance == s.cfg.MinChurnAmount {
		return currentBalance
	}
	return uint64(rand.Int63n(
		int64(currentBalance)-int64(s.cfg.MinChurnAmount)+1,
	) + int64(s.cfg.MinChurnAmount))
}
