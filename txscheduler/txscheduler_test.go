package txscheduler

import (
	"context"
	"testing"
	"time"

	"github.com/bastiankoetsier/schedule"
)

var (
	a   = make(chan bool, 2)
	set = false
)

func TestTxScheduler(t *testing.T) {
	s := New(context.Background())
	s.Start()
	s.Scheduler().Command(schedule.RunFunc(func(ctx context.Context) {
		if set == false {
			a <- true
		}
	})).EveryMinute()
	time.Sleep(time.Second * 65)
	b := <-a
	if !b {
		t.Error("failed to schedule task")
	}
	s.Stop()
}
