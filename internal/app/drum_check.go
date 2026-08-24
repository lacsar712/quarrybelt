package app

import (
	"fmt"

	"github.com/lacsar712/quarrybelt/internal/model"
)

func (a *App) CheckScaleLevel(snap model.PlantSnapshot) error {
	if snap.Scale.LevelPercent < model.MinScaleLevelPercent {
		return fmt.Errorf("%w", model.ErrScaleLevelLow)
	}
	return nil
}
