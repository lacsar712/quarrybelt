package app

import (
	"context"
	"fmt"
	"time"

	"github.com/lacsar712/quarrybelt/internal/fsm"
	"github.com/lacsar712/quarrybelt/internal/interlock"
	"github.com/lacsar712/quarrybelt/internal/model"
)

func (a *App) StartBeltprep(ctx context.Context, holder string) error {
	now := a.clk.Now()
	return a.interlock.Leases().WithLease(ctx, a.cfg.UnitID, holder, now, func(ctx context.Context) error {
		snap := a.Snapshot()
		if err := a.interlock.Gate().Allow(a.cfg.UnitID, snap.State); err != nil {
			return err
		}
		comb, err := a.beltbank.StartBeltprep(ctx, snap)
		if err != nil {
			return err
		}
		_ = a.store.UpdateBeltbank(a.cfg.UnitID, comb)
		state, err := a.fsm.Dispatch(ctx, fsm.EvStartBeltprep)
		if err != nil {
			return err
		}
		a.syncState(state)
		a.journalEvent("start_beltprep", storePayload("holder", holder))
		return nil
	})
}

func (a *App) CompleteBeltprep(ctx context.Context, holder string) error {
	now := a.clk.Now()
	return a.interlock.Leases().WithLease(ctx, a.cfg.UnitID, holder, now, func(ctx context.Context) error {
		snap := a.Snapshot()
		if err := a.beltbank.CompleteBeltprep(snap.Beltbank); err != nil {
			return err
		}
		state, err := a.fsm.Dispatch(ctx, fsm.EvBeltprepComplete)
		if err != nil {
			return err
		}
		a.syncState(state)
		a.journalEvent("beltprep_complete", "")
		return nil
	})
}

func (a *App) Ignite(ctx context.Context, holder string) error {
	now := a.clk.Now()
	return a.interlock.Leases().WithLease(ctx, a.cfg.UnitID, holder, now, func(ctx context.Context) error {
		snap := a.Snapshot()
		a.refreshPermissives(snap)
		if err := a.permissives.CheckIgnition(); err != nil {
			return err
		}
		comb, err := a.beltbank.Ignite(ctx, snap)
		if err != nil {
			return err
		}
		_ = a.store.UpdateBeltbank(a.cfg.UnitID, comb)
		state, err := a.fsm.Dispatch(ctx, fsm.EvIgnite)
		if err != nil {
			return err
		}
		a.syncState(state)
		a.journalEvent("ignite", storePayload("holder", holder))
		return nil
	})
}

func (a *App) StabilizeIgnition(ctx context.Context, holder string) error {
	now := a.clk.Now()
	return a.interlock.Leases().WithLease(ctx, a.cfg.UnitID, holder, now, func(ctx context.Context) error {
		snap := a.Snapshot()
		comb, err := a.beltbank.Stabilize(snap)
		if err != nil {
			return err
		}
		_ = a.store.UpdateBeltbank(a.cfg.UnitID, comb)
		state, err := a.fsm.Dispatch(ctx, fsm.EvIgnitionStable)
		if err != nil {
			return err
		}
		a.syncState(state)
		a.journalEvent("ignition_stable", "")
		return nil
	})
}

func (a *App) RampLoad(ctx context.Context, holder string, loadPct float64) error {
	now := a.clk.Now()
	return a.interlock.Leases().WithLease(ctx, a.cfg.UnitID, holder, now, func(ctx context.Context) error {
		snap := a.Snapshot()
		if err := a.interlock.AuthorizeFiring(a.cfg.UnitID, holder, snap.State, now); err != nil {
			return err
		}
		comb := a.beltbank.RampToLoad(snap, loadPct)
		_ = a.store.UpdateBeltbank(a.cfg.UnitID, comb)
		yardReading, err := a.yard.Tick(ctx, snap, true)
		if err != nil {
			return err
		}
		_ = a.store.UpdateYard(a.cfg.UnitID, yardReading)
		state, err := a.fsm.Dispatch(ctx, fsm.EvReachFiring)
		if err != nil {
			return err
		}
		a.syncState(state)
		a.journalEvent("ramp_load", fmt.Sprintf(`{"load_pct":%f}`, loadPct))
		return nil
	})
}

func (a *App) Trip(ctx context.Context, reason string) error {
	snap := a.Snapshot()
	comb := a.beltbank.Trip(snap.Beltbank)
	_ = a.store.UpdateBeltbank(a.cfg.UnitID, comb)
	a.interlock.Gate().Block(a.cfg.UnitID, "trip")
	state, err := a.fsm.Dispatch(ctx, fsm.EvTrip)
	if err != nil {
		return err
	}
	a.syncState(state)
	a.journalEvent("trip", storePayload("reason", reason))
	return nil
}

func (a *App) ResetTrip(ctx context.Context, holder string) error {
	now := a.clk.Now()
	return a.interlock.Leases().WithLease(ctx, a.cfg.UnitID, holder, now, func(ctx context.Context) error {
		a.interlock.Gate().Unblock(a.cfg.UnitID)
		state, err := a.fsm.Dispatch(ctx, fsm.EvResetTrip)
		if err != nil {
			return err
		}
		a.syncState(state)
		a.journalEvent("reset_trip", "")
		return nil
	})
}

func (a *App) UpdateSettings(ctx context.Context, settings model.PlantSettings) error {
	select {
	case <-ctx.Done():
		return model.ErrContextDone
	default:
	}
	if err := a.yard.ValidateSettings(settings); err != nil {
		return err
	}
	return a.store.UpdateSettings(a.cfg.UnitID, settings)
}

func (a *App) EnterService(ctx context.Context, holder string) error {
	now := a.clk.Now()
	return a.interlock.Leases().WithLease(ctx, a.cfg.UnitID, holder, now, func(ctx context.Context) error {
		state, err := a.fsm.Dispatch(ctx, fsm.EvEnterService)
		if err != nil {
			return err
		}
		a.syncState(state)
		return nil
	})
}

func (a *App) OnDriveLoss(ctx context.Context, holder string) error {
	_ = holder
	snap := a.Snapshot()
	if err := interlock.CheckDriveLoss(snap.Beltbank); err != nil {
		return fmt.Errorf("drive loss: %w", err)
	}
	return nil
}

func (a *App) Shutdown(ctx context.Context, holder string) error {
	now := a.clk.Now()
	return a.interlock.Leases().WithLease(ctx, a.cfg.UnitID, holder, now, func(ctx context.Context) error {
		a.StopTickLoop()
		a.cancelBeltLoop(holder)
		state, err := a.fsm.Dispatch(ctx, fsm.EvShutdown)
		if err != nil {
			return err
		}
		a.syncState(state)
		a.journalEvent("shutdown", "")
		return nil
	})
}

func (a *App) StartTickLoop(parent context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if cancel, ok := a.tickCancels[a.tickLabel()]; ok {
		cancel()
	}
	ctx, cancel := context.WithCancel(parent)
	a.tickCancels[a.tickLabel()] = cancel
	interval := a.cfg.TickInterval
	if interval <= 0 {
		interval = time.Second
	}
	_ = a.scheduler.Schedule(ctx, a.tickLabel(), interval, a.runTick)
}

func (a *App) StopTickLoop() {
	a.mu.Lock()
	defer a.mu.Unlock()
	if cancel, ok := a.tickCancels[a.tickLabel()]; ok {
		cancel()
		delete(a.tickCancels, a.tickLabel())
	}
}

func (a *App) runTick(ctx context.Context) error {
	snap := a.Snapshot()
	firing := a.isFiring(snap.State)
	scaleReading, err := a.scale.Tick(ctx, snap, firing)
	if err != nil {
		return err
	}
	_ = a.store.UpdateScale(a.cfg.UnitID, scaleReading)
	snap, _ = a.store.Require(a.cfg.UnitID)
	yardReading, err := a.yard.Tick(ctx, snap, firing)
	if err != nil {
		return err
	}
	_ = a.store.UpdateYard(a.cfg.UnitID, yardReading)
	a.refreshPermissives(a.Snapshot())
	a.telemetry.RecordTick(firing)
	if a.scale.TripRequired(scaleReading) {
		_ = a.Trip(ctx, "scale_level")
	}
	return nil
}
