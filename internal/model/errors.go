package model

import "errors"

var (
	ErrContextDone      = errors.New("operation cancelled")
	ErrPlantNotFound    = errors.New("plant unit not found")
	ErrLeaseHeld        = errors.New("interlock lease held by another operator")
	ErrLeaseMissing     = errors.New("interlock lease missing or expired")
	ErrGateBlocked      = errors.New("safety gate blocked")
	ErrBeltPermissive   = errors.New("belt permissive not satisfied")
	ErrIgnitionBlocked  = errors.New("ignition sequence blocked")
	ErrScaleLevelTrip    = errors.New("scale level trip condition")
	ErrPressureTrip     = errors.New("steam pressure trip condition")
	ErrBeltbankTrip   = errors.New("beltbank trip condition")
	ErrIllegalState     = errors.New("illegal plant state transition")
	ErrSnapshotStale    = errors.New("snapshot revision stale")
	ErrWindowOpen       = errors.New("timing window still open")
	ErrBeltprepIncomplete  = errors.New("chute beltprep incomplete")
	ErrCoordinationLock = errors.New("coordination lock held")
	ErrScaleLevelLow     = errors.New("scale level below low limit")
	ErrDriveLoss        = errors.New("chute drive lost")
	ErrBypassLimit    = errors.New("bypass valve at limit")
)
