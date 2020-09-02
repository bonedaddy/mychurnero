package txscheduler

import (
	"context"
	"sync"

	"github.com/bastiankoetsier/schedule"
)

type TxScheduler struct {
	once         sync.Once
	ctx          context.Context
	cancel       context.CancelFunc
	s            *schedule.Schedule
	shutdownChan <-chan struct{}
}

func New(ctx context.Context) *TxScheduler {
	ctx, cancel := context.WithCancel(ctx)
	return &TxScheduler{ctx: ctx, cancel: cancel, s: &schedule.Schedule{}}
}

func (tx *TxScheduler) Scheduler() *schedule.Schedule {
	return tx.s
}

func (tx *TxScheduler) Start() {
	tx.once.Do(func() {
		tx.shutdownChan = tx.s.Start(tx.ctx)
	})
}

func (tx *TxScheduler) Stop() {
	tx.cancel()
	if tx.shutdownChan != nil {
		<-tx.shutdownChan
	}

}
