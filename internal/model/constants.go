package model

import "time"

const (
	DefaultLeaseTTL        = 30 * time.Second
	BeltprepWindow            = 5 * time.Minute
	IgnitionDelayWindow    = 15 * time.Second
	ScaleSwellSettleWindow  = 45 * time.Second
	BeltbankWarmupWindow = 2 * time.Minute
	FeedwaterRampWindow    = 30 * time.Second
	MaxScaleLevelPercent    = 95.0
	MinScaleLevelPercent    = 15.0
	TripScaleLowPercent     = 10.0
	TripScaleHighPercent    = 98.0
	NormalSteamPressurePSI = 1800.0
	MaxSteamPressurePSI    = 2000.0
	MinChuteO2Percent    = 2.5
	MaxChuteO2Percent    = 6.0
	DefaultJournalCapacity = 512
)
