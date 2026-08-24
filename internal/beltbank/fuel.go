package beltbank

import (
	"math"

	"github.com/lacsar712/quarrybelt/internal/clock"
	"github.com/lacsar712/quarrybelt/internal/model"
)

type BeltRegulator struct {
	clk clock.ProcessClock
}

func NewBeltRegulator(clk clock.ProcessClock) *BeltRegulator {
	return &BeltRegulator{clk: clk}
}

func (f *BeltRegulator) IgnitionRate(settings model.PlantSettings) float64 {
	return settings.BeltFlowTPH * 0.08
}

func (f *BeltRegulator) ComputeForLoad(settings model.PlantSettings, loadPct float64) float64 {
	loadPct = math.Max(0, math.Min(1, loadPct))
	return settings.BeltFlowTPH * loadPct
}

func (f *BeltRegulator) Ramp(current, target, maxStep float64) float64 {
	delta := target - current
	if math.Abs(delta) <= maxStep {
		return target
	}
	if delta > 0 {
		return current + maxStep
	}
	return current - maxStep
}

func (f *BeltRegulator) BtuPerHour(flowTPH float64) float64 {
	return flowTPH * 19_500_000
}

func (f *BeltRegulator) HeatInputMW(flowTPH float64) float64 {
	return flowTPH * 11.6
}

func (f *BeltRegulator) ValidatePermissive(settings model.PlantSettings, scaleOK, beltprepOK bool) error {
	if !beltprepOK {
		return model.ErrBeltprepIncomplete
	}
	if !scaleOK {
		return model.ErrScaleLevelTrip
	}
	if settings.BeltFlowTPH <= 0 {
		return model.ErrBeltPermissive
	}
	return nil
}

func (f *BeltRegulator) MinFlow(settings model.PlantSettings) float64 {
	return settings.BeltFlowTPH * 0.2
}

func (f *BeltRegulator) MaxFlow(settings model.PlantSettings) float64 {
	return settings.BeltFlowTPH * 1.1
}
