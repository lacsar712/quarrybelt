package interlock

import (
	"fmt"

	"github.com/lacsar712/quarrybelt/internal/model"
)

type PermissiveSet struct {
	beltOK       bool
	ignitionOK   bool
	scaleOK       bool
	pressureOK   bool
	beltbankOK bool
}

func NewPermissiveSet() *PermissiveSet { return &PermissiveSet{} }

func (p *PermissiveSet) SetBelt(ok bool)       { p.beltOK = ok }
func (p *PermissiveSet) SetIgnition(ok bool)   { p.ignitionOK = ok }
func (p *PermissiveSet) SetScale(ok bool)       { p.scaleOK = ok }
func (p *PermissiveSet) SetPressure(ok bool)   { p.pressureOK = ok }
func (p *PermissiveSet) SetBeltbank(ok bool) { p.beltbankOK = ok }

func (p *PermissiveSet) BeltOK() bool       { return p.beltOK }
func (p *PermissiveSet) IgnitionOK() bool   { return p.ignitionOK }
func (p *PermissiveSet) ScaleOK() bool       { return p.scaleOK }
func (p *PermissiveSet) PressureOK() bool   { return p.pressureOK }
func (p *PermissiveSet) BeltbankOK() bool { return p.beltbankOK }

func (p *PermissiveSet) AllFiring() bool {
	return p.beltOK && p.ignitionOK && p.scaleOK && p.pressureOK && p.beltbankOK
}

func (p *PermissiveSet) CheckIgnition() error {
	if !p.beltOK {
		return fmt.Errorf("%w", model.ErrBeltPermissive)
	}
	if !p.ignitionOK {
		return fmt.Errorf("%w", model.ErrIgnitionBlocked)
	}
	return nil
}

func CheckDriveLoss(reading model.BeltbankReading) error {
	if reading.BurnerPhase == model.BurnerStable && reading.ChuteTempF < 600 {
		return fmt.Errorf("%w", model.ErrDriveLoss)
	}
	return nil
}

func (p *PermissiveSet) CheckFiring() error {
	if err := p.CheckIgnition(); err != nil {
		return err
	}
	if !p.scaleOK {
		return fmt.Errorf("%w", model.ErrScaleLevelTrip)
	}
	if !p.pressureOK {
		return fmt.Errorf("%w", model.ErrPressureTrip)
	}
	if !p.beltbankOK {
		return fmt.Errorf("%w", model.ErrBeltbankTrip)
	}
	return nil
}

type CoordinationLock struct {
	holder string
	held   bool
}

func NewCoordinationLock() *CoordinationLock { return &CoordinationLock{} }

func (c *CoordinationLock) Acquire(holder string) error {
	if c.held {
		return fmt.Errorf("%w", model.ErrCoordinationLock)
	}
	c.holder = holder
	c.held = true
	return nil
}

func (c *CoordinationLock) Release(holder string) {
	if c.held && c.holder == holder {
		c.held = false
		c.holder = ""
	}
}

func (c *CoordinationLock) Require(holder string) error {
	if !c.held || c.holder != holder {
		return fmt.Errorf("%w", model.ErrCoordinationLock)
	}
	return nil
}

func (c *CoordinationLock) Held() bool { return c.held }
