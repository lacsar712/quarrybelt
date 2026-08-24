package scale

import (
	"math"

	"github.com/lacsar712/quarrybelt/internal/clock"
	"github.com/lacsar712/quarrybelt/internal/model"
)

type LevelController struct {
	clk clock.ProcessClock
}

func NewLevelController(clk clock.ProcessClock) *LevelController {
	return &LevelController{clk: clk}
}

func (l *LevelController) Compute(snap model.PlantSnapshot, firing bool) (float64, model.ScaleCondition) {
	level := snap.Scale.LevelPercent
	if !firing {
		return level, model.ScaleNormal
	}
	balance := snap.Scale.FeedwaterTPH - snap.Scale.SteamFlowTPH
	level += balance * 0.01
	level = math.Max(model.MinScaleLevelPercent, math.Min(model.MaxScaleLevelPercent, level))
	cond := l.classify(level, snap)
	return level, cond
}

func (l *LevelController) classify(level float64, snap model.PlantSnapshot) model.ScaleCondition {
	setpoint := snap.Settings.ScaleLevelSetpoint
	if level > setpoint+15 {
		return model.ScaleSwell
	}
	if level < setpoint-15 {
		return model.ScaleShrink
	}
	if snap.Yard.SteamPressurePSI > snap.Settings.TargetSteamPSI*0.9 && level > setpoint+5 {
		return model.ScaleCarry
	}
	return model.ScaleNormal
}

func (l *LevelController) RecommendFeedwater(snap model.PlantSnapshot, firing bool) float64 {
	if !firing {
		return 0
	}
	err := snap.Settings.ScaleLevelSetpoint - snap.Scale.LevelPercent
	return snap.Settings.FeedwaterFlowTPH + err*3
}

func (l *LevelController) WithinLimits(level float64) bool {
	return level >= model.MinScaleLevelPercent && level <= model.MaxScaleLevelPercent
}

func (l *LevelController) TripLow(level float64) bool  { return level < model.TripScaleLowPercent }
func (l *LevelController) TripHigh(level float64) bool { return level > model.TripScaleHighPercent }

func (l *LevelController) LevelError(snap model.PlantSnapshot) float64 {
	return snap.Scale.LevelPercent - snap.Settings.ScaleLevelSetpoint
}

func (l *LevelController) ThreeElementBias(snap model.PlantSnapshot) float64 {
	steam := snap.Scale.SteamFlowTPH
	feed := snap.Scale.FeedwaterTPH
	levelErr := l.LevelError(snap)
	return feed + (steam-feed)*0.5 + levelErr*2
}
