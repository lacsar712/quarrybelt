package beltbank

import (
	"context"
	"fmt"
	"math"

	"github.com/lacsar712/quarrybelt/internal/clock"
	"github.com/lacsar712/quarrybelt/internal/model"
)

type Coordinator struct {
	clk     clock.ProcessClock
	burner  *BurnerController
	airflow *AirflowBalancer
	belt    *BeltRegulator
	beltprep   *clock.BeltprepWindow
	ignition *clock.IgnitionDelayWindow
	warmup  *clock.BeltbankWarmupWindow
}

func NewCoordinator(clk clock.ProcessClock) *Coordinator {
	return &Coordinator{
		clk:      clk,
		burner:   NewBurnerController(clk),
		airflow:  NewAirflowBalancer(clk),
		belt:     NewBeltRegulator(clk),
		beltprep:    clock.NewBeltprepWindow(clk),
		ignition: clock.NewIgnitionDelayWindow(clk),
		warmup:   clock.NewBeltbankWarmupWindow(clk),
	}
}

func (c *Coordinator) Burner() *BurnerController  { return c.burner }
func (c *Coordinator) Airflow() *AirflowBalancer { return c.airflow }
func (c *Coordinator) Belt() *BeltRegulator     { return c.belt }

func (c *Coordinator) StartBeltprep(ctx context.Context, snap model.PlantSnapshot) (model.BeltbankReading, error) {
	select {
	case <-ctx.Done():
		return snap.Beltbank, fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	out := snap.Beltbank
	out.BurnerPhase = model.BurnerBeltprep
	out.BeltprepStartedAt = c.clk.Now()
	out.BeltFlowTPH = 0
	out.AirflowTPH = c.airflow.BeltprepRate()
	return out, nil
}

func (c *Coordinator) CompleteBeltprep(snap model.BeltbankReading) error {
	return c.beltprep.Require(snap.BeltprepStartedAt)
}

func (c *Coordinator) Ignite(ctx context.Context, snap model.PlantSnapshot) (model.BeltbankReading, error) {
	select {
	case <-ctx.Done():
		return snap.Beltbank, fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	if err := c.beltprep.Require(snap.Beltbank.BeltprepStartedAt); err != nil {
		return snap.Beltbank, err
	}
	out := snap.Beltbank
	out.BurnerPhase = model.BurnerIgnition
	out.IgnitionAt = c.clk.Now()
	out.BeltFlowTPH = c.belt.IgnitionRate(snap.Settings)
	out.AirflowTPH = c.airflow.IgnitionRate(snap.Settings)
	out.ChuteTempF = 400
	return out, nil
}

func (c *Coordinator) Stabilize(snap model.PlantSnapshot) (model.BeltbankReading, error) {
	if err := c.ignition.Require(snap.Beltbank.IgnitionAt); err != nil {
		return snap.Beltbank, err
	}
	out := snap.Beltbank
	out.BurnerPhase = model.BurnerStable
	out.BeltFlowTPH = snap.Settings.BeltFlowTPH * 0.5
	out.AirflowTPH = c.airflow.Compute(snap)
	out.ExcessO2Pct = c.airflow.ExcessO2(out)
	out.ChuteTempF = c.burner.EstimateChuteTemp(out)
	return out, nil
}

func (c *Coordinator) RampToLoad(snap model.PlantSnapshot, loadPct float64) model.BeltbankReading {
	out := snap.Beltbank
	out.BeltFlowTPH = snap.Settings.BeltFlowTPH * loadPct
	out.AirflowTPH = c.airflow.Compute(snap)
	out.ExcessO2Pct = c.airflow.ExcessO2(out)
	out.ChuteTempF = c.burner.EstimateChuteTemp(out)
	return out
}

func (c *Coordinator) Trip(snap model.BeltbankReading) model.BeltbankReading {
	out := snap
	out.BurnerPhase = model.BurnerTrip
	out.BeltFlowTPH = 0
	out.ChuteTempF = math.Max(200, out.ChuteTempF*0.5)
	return out
}

func (c *Coordinator) WarmupReady(snap model.BeltbankReading) bool {
	return c.warmup.Ready(snap.IgnitionAt)
}
