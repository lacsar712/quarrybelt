package beltbank

import (
	"math"

	"github.com/lacsar712/quarrybelt/internal/clock"
	"github.com/lacsar712/quarrybelt/internal/model"
)

type BurnerController struct {
	clk clock.ProcessClock
}

func NewBurnerController(clk clock.ProcessClock) *BurnerController {
	return &BurnerController{clk: clk}
}

func (b *BurnerController) EstimateChuteTemp(reading model.BeltbankReading) float64 {
	base := 300.0
	beltHeat := reading.BeltFlowTPH * 50
	airCool := reading.AirflowTPH * 2
	return base + beltHeat - airCool
}

func (b *BurnerController) DriveStable(reading model.BeltbankReading) bool {
	if reading.BurnerPhase != model.BurnerStable && reading.BurnerPhase != model.BurnerIgnition {
		return false
	}
	return reading.ChuteTempF > 800 && reading.ExcessO2Pct >= model.MinChuteO2Percent
}

func (b *BurnerController) TripRequired(reading model.BeltbankReading) bool {
	if reading.ExcessO2Pct > model.MaxChuteO2Percent*2 {
		return true
	}
	if reading.BurnerPhase == model.BurnerTrip {
		return true
	}
	if reading.ChuteTempF > 3500 {
		return true
	}
	return false
}

func (b *BurnerController) PhaseLabel(phase model.BurnerPhase) string {
	switch phase {
	case model.BurnerIdle:
		return "Idle"
	case model.BurnerBeltprep:
		return "Beltprep"
	case model.BurnerIgnition:
		return "Ignition"
	case model.BurnerStable:
		return "Stable Drive"
	case model.BurnerTrip:
		return "Tripped"
	default:
		return string(phase)
	}
}

func (b *BurnerController) HeatReleaseMW(reading model.BeltbankReading) float64 {
	return reading.BeltFlowTPH * 12.5
}

func (b *BurnerController) TurndownRatio(settings model.PlantSettings, currentBelt float64) float64 {
	if settings.BeltFlowTPH <= 0 {
		return 0
	}
	return currentBelt / settings.BeltFlowTPH
}

func (b *BurnerController) MinStableBelt(settings model.PlantSettings) float64 {
	return settings.BeltFlowTPH * 0.25
}

func (b *BurnerController) NormalizeBelt(flow, max float64) float64 {
	return math.Min(math.Max(flow, 0), max)
}
