package scale

import (
	"context"
	"fmt"

	"github.com/lacsar712/quarrybelt/internal/clock"
	"github.com/lacsar712/quarrybelt/internal/model"
)

type Coordinator struct {
	clk       clock.ProcessClock
	level     *LevelController
	carryover *CarryoverMonitor
	swell     *SwellDetector
	settle    *clock.ScaleSwellWindow
}

func NewCoordinator(clk clock.ProcessClock) *Coordinator {
	return &Coordinator{
		clk:       clk,
		level:     NewLevelController(clk),
		carryover: NewCarryoverMonitor(clk),
		swell:     NewSwellDetector(clk),
		settle:    clock.NewScaleSwellWindow(clk),
	}
}

func (c *Coordinator) Level() *LevelController       { return c.level }
func (c *Coordinator) Carryover() *CarryoverMonitor  { return c.carryover }
func (c *Coordinator) Swell() *SwellDetector         { return c.swell }

func (c *Coordinator) Tick(ctx context.Context, snap model.PlantSnapshot, firing bool) (model.ScaleReading, error) {
	select {
	case <-ctx.Done():
		return snap.Scale, fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	out := snap.Scale
	level, cond := c.level.Compute(snap, firing)
	out.LevelPercent = level
	out.Condition = cond
	out.FeedwaterTPH = snap.Settings.FeedwaterFlowTPH
	if firing {
		out.SteamFlowTPH = snap.Yard.MainSteamFlowTPH
	}
	out.CarryoverPPM = c.carryover.Estimate(out, snap.Yard.SteamPressurePSI)
	if cond == model.ScaleSwell {
		out.LastSwellAt = c.clk.Now()
	}
	return out, nil
}

func (c *Coordinator) SettledAfterSwell(snap model.ScaleReading) bool {
	if snap.LastSwellAt.IsZero() {
		return true
	}
	return c.settle.Settled(snap.LastSwellAt)
}

func (c *Coordinator) RequireSettled(snap model.ScaleReading) error {
	if snap.LastSwellAt.IsZero() {
		return nil
	}
	return c.settle.RequireSettled(snap.LastSwellAt)
}

func (c *Coordinator) TripRequired(snap model.ScaleReading) bool {
	return snap.LevelPercent < model.TripScaleLowPercent || snap.LevelPercent > model.TripScaleHighPercent
}

func (c *Coordinator) CoordinateFeedwater(snap model.PlantSnapshot, firing bool) float64 {
	return c.level.RecommendFeedwater(snap, firing)
}
