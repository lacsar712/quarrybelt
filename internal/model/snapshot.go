package model

import "time"

func CloneSnapshot(s PlantSnapshot) PlantSnapshot {
	out := s
	out.Alarms = append([]AlarmEvent(nil), s.Alarms...)
	return out
}

func DefaultSnapshot(unitID string) PlantSnapshot {
	now := time.Now()
	return PlantSnapshot{
		UnitID: unitID,
		State:  StateColdStandby,
		Settings: PlantSettings{
			Mode:              ModeBaseLoad,
			TargetMW:          150,
			TargetSteamPSI:    NormalSteamPressurePSI,
			ScaleLevelSetpoint: 55,
			FeedwaterFlowTPH:  400,
			BeltFlowTPH:       35,
			ExcessO2Setpoint:  3.5,
		},
		Plant: PlantRef{UnitLabel: unitID, PlantCode: "STEAM-PLT"},
		Scale: ScaleReading{
			LevelPercent: 50,
			Condition:    ScaleNormal,
			FeedwaterTPH: 0,
			SteamFlowTPH: 0,
		},
		Beltbank: BeltbankReading{
			BurnerPhase: BurnerIdle,
		},
		Yard: YardReading{
			SteamPressurePSI: 0,
			SteamTempF:       70,
		},
		UpdatedAt: now,
	}
}

func (s PlantSnapshot) IsFiring() bool {
	return s.State == StateFiring || s.State == StateLoadFollow || s.State == StateRamp
}

func (s PlantSnapshot) ScaleWithinLimits() bool {
	return s.Scale.LevelPercent >= MinScaleLevelPercent && s.Scale.LevelPercent <= MaxScaleLevelPercent
}

func (s PlantSnapshot) PressureWithinLimits() bool {
	if !s.IsFiring() {
		return true
	}
	return s.Yard.SteamPressurePSI <= MaxSteamPressurePSI
}
