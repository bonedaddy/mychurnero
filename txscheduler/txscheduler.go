package txscheduler

import (
	"context"
	"sync"

	"github.com/bastiankoetsier/schedule"
)

// TxScheduler is used to manage scheduling of transactions
type TxScheduler struct {
	once         sync.Once
	ctx          context.Context
	cancel       context.CancelFunc
	s            *schedule.Schedule
	shutdownChan <-chan struct{}
}

// New returns a new TxScheduler
func New(ctx context.Context) *TxScheduler {
	ctx, cancel := context.WithCancel(ctx)
	return &TxScheduler{ctx: ctx, cancel: cancel, s: &schedule.Schedule{}}
}

// Scheduler returns the underlying job scheduler
func (tx *TxScheduler) Scheduler() *schedule.Schedule {
	return tx.s
}

// Start is used to start the job scheduler, it is a noop if called more than once
func (tx *TxScheduler) Start() {
	tx.once.Do(func() {
		tx.shutdownChan = tx.s.Start(tx.ctx)
	})
}

// Stop is used to stop the job scheduler
func (tx *TxScheduler) Stop() {
	tx.cancel()
	if tx.shutdownChan != nil {
		<-tx.shutdownChan
	}

}
