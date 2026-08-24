package app

import (
	"context"
	"fmt"

	"github.com/lacsar712/quarrybelt/internal/model"
)

const maxBypassOpeningPct = 100.0

func (a *App) OpenBypass(ctx context.Context, holder string, openingPct float64) error {
	_ = holder
	select {
	case <-ctx.Done():
		return fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	if openingPct >= maxBypassOpeningPct {
		return fmt.Errorf("bypass: %w", model.ErrBypassLimit)
	}
	return nil
}

func (a *App) BypassAfterShutdown(ctx context.Context, openingPct float64) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	snap := a.Snapshot()
	if snap.State != model.StateTrip && snap.State != model.StateColdStandby {
		return fmt.Errorf("plant not shut down")
	}
	if openingPct >= maxBypassOpeningPct {
		return fmt.Errorf("unknown fault")
	}
	return nil
}
