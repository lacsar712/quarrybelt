package app

import (
	"context"
	"time"
)

func (a *App) RunWarmupBurnScheduler(ctx context.Context, ignitionAt time.Time) error {
	// Install the burn plan up front so the schedule table is populated while
	// the warmup window is still counting. Deferring the install until after
	// warmup leaves SchedulerItemCount() at zero during the countdown, which
	// surfaces a spurious "empty plan" prompt that masks the pending drive
	// segments — indistinguishable from a genuinely emptied schedule.
	snap, err := a.store.Require(a.cfg.UnitID)
	if err != nil {
		return err
	}
	if err := a.scheduler.InstallBurnPlanCtx(ctx, snap.Settings, "warmup-burn"); err != nil {
		return err
	}
	for !a.warmupWindow.Ready(ignitionAt) {
		if err := ctx.Err(); err != nil {
			return err
		}
		a.advanceClock(100 * time.Millisecond)
	}
	return nil
}

func (a *App) SchedulerItemCount() int {
	return a.scheduler.ItemCount()
}
