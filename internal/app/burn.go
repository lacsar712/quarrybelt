package app

import (
	"context"
	"time"
)

func (a *App) RunWarmupBurnScheduler(ctx context.Context, ignitionAt time.Time) error {
	for !a.warmupWindow.Ready(ignitionAt) {
		if err := ctx.Err(); err != nil {
			return err
		}
		a.advanceClock(100 * time.Millisecond)
	}
	snap, err := a.store.Require(a.cfg.UnitID)
	if err != nil {
		return err
	}
	// Propagate the operator's cancellation context into the plan install so a
	// 暖机撤单 aborts subsequent drive-step appends, not just the UI. Detaching
	// to context.Background() defeated InstallBurnPlanCtx's per-step ctx.Err()
	// check and left the old drive segment queued after cancellation.
	return a.scheduler.InstallBurnPlanCtx(ctx, snap.Settings, "warmup-burn")
}

func (a *App) SchedulerItemCount() int {
	return a.scheduler.ItemCount()
}
