package api

import (
	"errors"

	"github.com/lacsar712/quarrybelt/internal/model"
)

func classifyScaleError(err error) (string, bool) {
	if errors.Is(err, model.ErrScaleLevelLow) {
		return "scale_level_low", true
	}
	return "", false
}
