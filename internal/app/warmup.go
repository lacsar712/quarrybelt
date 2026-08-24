package app

import (
	"context"
	"fmt"

	"github.com/lacsar712/quarrybelt/internal/model"
)

func (a *App) WarmupStatus() (ready bool, detail string) {
	snap := a.Snapshot()
	if snap.Beltbank.BeltprepStartedAt.IsZero() {
		return false, "beltprep not started"
	}
	if !a.beltprepWindow.Ready(snap.Beltbank.BeltprepStartedAt) {
		return false, "beltprep window open"
	}
	if !snap.Beltbank.IgnitionAt.IsZero() && !a.warmupWindow.Ready(snap.Beltbank.IgnitionAt) {
		return false, "beltbank warmup window open"
	}
	if !snap.Scale.LastSwellAt.IsZero() {
		if err := a.scale.RequireSettled(snap.Scale); err != nil {
			return false, "scale swell settling"
		}
	}
	return true, "ready"
}

func (a *App) WaitWarmup(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("%w", model.ErrContextDone)
		default:
		}
		ready, _ := a.WarmupStatus()
		if ready {
			return nil
		}
	}
}

func (a *App) BeltprepRemaining() string {
	snap := a.Snapshot()
	if snap.Beltbank.BeltprepStartedAt.IsZero() {
		return "not started"
	}
	if a.beltprepWindow.Ready(snap.Beltbank.BeltprepStartedAt) {
		return "complete"
	}
	return "in progress"
}

func (a *App) BeltbankWarmupRemaining() string {
	snap := a.Snapshot()
	if snap.Beltbank.IgnitionAt.IsZero() {
		return "not ignited"
	}
	if a.warmupWindow.Ready(snap.Beltbank.IgnitionAt) {
		return "complete"
	}
	return "in progress"
}
