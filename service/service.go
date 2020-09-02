package service

import (
	"context"

	"github.com/bonedaddy/mychurnero/client"
	"github.com/bonedaddy/mychurnero/db"
	"github.com/bonedaddy/mychurnero/txscheduler"
	"go.uber.org/multierr"
)

type Service struct {
	scheduler *txscheduler.TxScheduler
	mc        *client.Client
	db        *db.Client
	ctx       context.Context
	cancel    context.CancelFunc
}

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
	sched := txscheduler.New(ctx)
	sched.Start()
	return &Service{sched, cl, db, ctx, cancel}, nil
}

func (s *Service) MC() *client.Client {
	return s.mc
}

func (s *Service) DB() *db.Client {
	return s.db
}

func (s *Service) Context() context.Context {
	return s.ctx
}

func (s *Service) Close() error {
	var closeErr error
	s.scheduler.Stop()
	s.cancel()
	if err := s.mc.Close(); err != nil {
		closeErr = err
	}
	if err := s.db.Close(); err != nil {
		closeErr = multierr.Combine(closeErr, err)
	}
	return closeErr
}
