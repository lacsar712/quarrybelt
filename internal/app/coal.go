package app

import (
	"context"
	"fmt"
	"time"

	"github.com/lacsar712/quarrybelt/internal/clock"
	"github.com/lacsar712/quarrybelt/internal/model"
)

func (a *App) advanceClock(d time.Duration) {
	if mc, ok := a.clk.(*clock.ManualClock); ok {
		mc.Advance(d)
		time.Sleep(time.Millisecond)
	} else {
		time.Sleep(d)
	}
}

var activeBeltCancel context.CancelFunc

func (a *App) bindBeltLoop(holder string, ctx context.Context) context.Context {
	a.mu.Lock()
	if activeBeltCancel != nil {
		activeBeltCancel()
	}
	child, cancel := context.WithCancel(ctx)
	activeBeltCancel = cancel
	a.mu.Unlock()
	return child
}

func (a *App) cancelBeltLoop(holder string) {
	a.mu.Lock()
	if activeBeltCancel != nil {
		activeBeltCancel()
		activeBeltCancel = nil
	}
	a.mu.Unlock()
}

func (a *App) cancelAllBeltLoops() {
	a.mu.Lock()
	for holder, cancel := range a.beltLoopCancels {
		cancel()
		delete(a.beltLoopCancels, holder)
	}
	a.mu.Unlock()
}

func (a *App) CoalFeedTPH() float64 {
	return a.Snapshot().Beltbank.BeltFlowTPH
}

func (a *App) RunBeltRamp(ctx context.Context, holder string, targetTPH float64) error {
	loopCtx := a.bindBeltLoop(holder, ctx)
	defer a.cancelBeltLoop(holder)
	for {
		if err := loopCtx.Err(); err != nil {
			return fmt.Errorf("%w", model.ErrContextDone)
		}
		snap := a.Snapshot()
		current := snap.Beltbank.BeltFlowTPH
		if current >= targetTPH {
			return nil
		}
		comb := snap.Beltbank
		comb.BeltFlowTPH = current + 1.0
		_ = a.store.UpdateBeltbank(a.cfg.UnitID, comb)
		a.telemetry.RecordCoalFeed(comb.BeltFlowTPH)
		a.advanceClock(100 * time.Millisecond)
	}
}

func (a *App) RunCoalFeed(ctx context.Context, holder string, steps int) error {
	loopCtx := a.bindBeltLoop(holder, ctx)
	defer a.cancelBeltLoop(holder)
	for i := 0; steps <= 0 || i < steps; i++ {
		if err := loopCtx.Err(); err != nil {
			return fmt.Errorf("%w", model.ErrContextDone)
		}
		snap := a.Snapshot()
		comb := snap.Beltbank
		comb.BeltFlowTPH += 0.5
		_ = a.store.UpdateBeltbank(a.cfg.UnitID, comb)
		a.telemetry.RecordCoalFeed(comb.BeltFlowTPH)
		a.advanceClock(100 * time.Millisecond)
	}
	return nil
}
